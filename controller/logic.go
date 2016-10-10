package controller

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/coreos/pkg/capnslog"
	"github.com/reechou/x-real-control/config"
	"github.com/reechou/x-real-control/detector"
	"github.com/reechou/x-real-control/utils"
)

const (
	CacheDir = ".cache"
)

var plog = capnslog.NewPackageLogger("github.com/reezhou/x-real-control", "controller")

type DomainMapInfo struct {
	groupInfo  *DomainGroupInfo
	domainList *DomainList
	dhc        *DomainCheckHealth
	idx        int64
}

type ContentMapInfo struct {
	groupInfo   *ContentGroupInfo
	contentList *ContentList
	cg          *ContentGenerate
}

type ControllerLogic struct {
	sync.Mutex

	cfg *config.Config

	aliyunOss *config.AliyunOss

	detector *detector.Detector
	cdb      *ControllerDB
	w        *utils.TimingWheel
	xServer  *XHttpServer

	domainMap       map[int64]*DomainMapInfo
	domainGroupList []int64
	domainGroupIdx  int64
	jumpDomainGroup []int64
	jumpDomainIdx   int64
	groupMaxID      int64

	contentMap        map[int64]*ContentMapInfo
	contentGroupList  []int64
	contentGroupIdx   int64
	contentGroupMaxID int64

	stop chan struct{}
	done chan struct{}
}

func NewControllerLogic(cfg *config.Config) *ControllerLogic {
	w := utils.NewTimingWheel(500*time.Millisecond, 120)
	d := detector.NewDetector(cfg)
	cl := &ControllerLogic{
		cfg:              cfg,
		aliyunOss:        &cfg.AliyunOss,
		w:                w,
		detector:         d,
		domainMap:        make(map[int64]*DomainMapInfo),
		domainGroupList:  make([]int64, 0),
		jumpDomainGroup:  make([]int64, 0),
		contentMap:       make(map[int64]*ContentMapInfo),
		contentGroupList: make([]int64, 0),
		stop:             make(chan struct{}),
		done:             make(chan struct{}),
	}
	aliyunClient, err := oss.New(cl.aliyunOss.Endpoint, cl.aliyunOss.AccessKeyId, cl.aliyunOss.AccessKeySecret)
	if err != nil {
		plog.Panicf("aliyun oss new error: %v\n", err)
	}
	cl.aliyunOss.AliyunClient = aliyunClient
	db, err := NewControllerDB(&cfg.MysqlInfo)
	if err != nil {
		plog.Panicf("db controller new error: %v\n", err)
	}
	cl.cdb = db
	err = cl.Init()
	if err != nil {
		plog.Panicf("logic init error: %v\n", err)
	}
	go cl.run()

	cl.xServer = NewXHttpServer(cfg.ListenAddr, cfg.ListenPort, cl)
	setupLogging(cfg)
	touchCacheDir()

	return cl
}

func (cl *ControllerLogic) Start() {
	cl.xServer.Run()
}

func (cl *ControllerLogic) Stop() {
	close(cl.stop)
	<-cl.done
}

func (cl *ControllerLogic) Init() error {
	groupList, groupMaxID, err := cl.cdb.GetDomainGroupList(0)
	if err != nil {
		plog.Error("[logic] init get domain group list error: %v\n", err)
		return err
	}
	cl.groupMaxID = groupMaxID
	for _, v := range groupList {
		domainList := &DomainList{
			GroupID: v.ID,
		}
		err := cl.cdb.GetDomainList(domainList)
		if err != nil {
			plog.Error("[logic] init get domain list error: %v\n", err)
			return err
		}
		dhc := NewDomainCheckHealth(v, cl.cdb, cl.w, cl, cl.cfg)
		cl.domainMap[v.ID] = &DomainMapInfo{
			groupInfo:  v,
			domainList: domainList,
			dhc:        dhc,
		}
		switch v.Type {
		case DOMAIN_GROUP_TYPE_JUMP:
			cl.jumpDomainGroup = append(cl.jumpDomainGroup, v.ID)
		case DOMAIN_GROUP_TYPE_SHOW:
			cl.domainGroupList = append(cl.domainGroupList, v.ID)
		}
	}

	// get content list
	contentGroupList, contentGroupMaxID, err := cl.cdb.GetContentGroupList(0)
	if err != nil {
		plog.Error("[logic] init get content group list error: %v\n", err)
		return err
	}
	cl.contentGroupMaxID = contentGroupMaxID
	for _, v := range contentGroupList {
		contentList := &ContentList{
			GroupID: v.ID,
		}
		err := cl.cdb.GetContentList(contentList)
		if err != nil {
			plog.Error("[logic] init get content list error: %v\n", err)
			return err
		}
		cl.contentMap[v.ID] = &ContentMapInfo{
			groupInfo:   v,
			contentList: contentList,
			cg:          NewContentGenerate(v, cl.cdb, cl.w, cl, cl.aliyunOss),
		}
		cl.contentGroupList = append(cl.contentGroupList, v.ID)
	}

	return nil
}

func (cl *ControllerLogic) run() {
	for {
		select {
		case <-time.After(30 * time.Second):
			cl.onRefresh()
		case <-cl.stop:
			close(cl.done)
			return
		}
	}
}

func (cl *ControllerLogic) onRefresh() {
	groupList, groupMaxID, err := cl.cdb.GetDomainGroupList(cl.groupMaxID)
	if err != nil {
		plog.Error("[onRefresh] get domain group list error: %v\n", err)
	} else {
		cl.groupMaxID = groupMaxID
		for _, v := range groupList {
			domainList := &DomainList{
				GroupID: v.ID,
			}
			err := cl.cdb.GetDomainList(domainList)
			if err != nil {
				plog.Error("[onRefresh] get domain list error: %v\n", err)
			} else {
				dhc := NewDomainCheckHealth(v, cl.cdb, cl.w, cl, cl.cfg)
				cl.Lock()
				cl.domainMap[v.ID] = &DomainMapInfo{
					groupInfo:  v,
					domainList: domainList,
					dhc:        dhc,
				}
				switch v.Type {
				case DOMAIN_GROUP_TYPE_JUMP:
					cl.jumpDomainGroup = append(cl.jumpDomainGroup, v.ID)
				case DOMAIN_GROUP_TYPE_SHOW:
					cl.domainGroupList = append(cl.domainGroupList, v.ID)
				}
				cl.Unlock()
			}
		}
	}

	contentGroupList, contentGroupMaxID, err := cl.cdb.GetContentGroupList(cl.contentGroupMaxID)
	if err != nil {
		plog.Error("[logic] init get content group list error: %v\n", err)
	} else {
		cl.contentGroupMaxID = contentGroupMaxID
		for _, v := range contentGroupList {
			contentList := &ContentList{
				GroupID: v.ID,
			}
			err := cl.cdb.GetContentList(contentList)
			if err != nil {
				plog.Error("[logic] init get content list error: %v\n", err)
			} else {
				cl.Lock()
				cl.contentMap[v.ID] = &ContentMapInfo{
					groupInfo:   v,
					contentList: contentList,
					cg:          NewContentGenerate(v, cl.cdb, cl.w, cl, cl.aliyunOss),
				}
				cl.contentGroupList = append(cl.contentGroupList, v.ID)
				cl.Unlock()
			}
		}
	}
}

func (cl *ControllerLogic) UpdateDomainGroup(groupInfo *DomainGroupInfo, domainList *DomainList) {
	cl.Lock()
	defer cl.Unlock()

	v := cl.domainMap[groupInfo.ID]
	if v != nil {
		v.domainList = domainList
	} else {
		cl.domainMap[groupInfo.ID] = &DomainMapInfo{
			groupInfo:  groupInfo,
			domainList: domainList,
		}
		cl.domainGroupList = append(cl.domainGroupList, groupInfo.ID)
	}
}

func (cl *ControllerLogic) UpdateContentGroup(groupInfo *ContentGroupInfo) {
	cl.Lock()
	defer cl.Unlock()

	v := cl.contentMap[groupInfo.ID]
	if v != nil {
		v.groupInfo = groupInfo
	}
}

func (cl *ControllerLogic) GetDomainInfo(id, t int64) (*DomainInfo, error) {
	cl.Lock()
	defer cl.Unlock()

	if t == DOMAIN_GROUP_TYPE_JUMP {
		oldJumpGroupIdx := cl.jumpDomainIdx
		for {
			groupID := cl.jumpDomainGroup[cl.jumpDomainIdx]
			cl.jumpDomainIdx = (cl.jumpDomainIdx + 1) % int64(len(cl.jumpDomainGroup))
			domain, err := cl.getDomainFromGroupID(groupID, t)
			if err == nil {
				return domain, nil
			}
			if cl.jumpDomainIdx == oldJumpGroupIdx {
				plog.Errorf("no useful jump domain!")
				return nil, fmt.Errorf("no useful jump domain!")
			}
		}
	}

	if id != 0 {
		return cl.getDomainFromGroupID(id, t)
	}

	oldGroupIdx := cl.domainGroupIdx
	for {
		groupID := cl.domainGroupList[cl.domainGroupIdx]
		cl.domainGroupIdx = (cl.domainGroupIdx + 1) % int64(len(cl.domainGroupList))
		domain, err := cl.getDomainFromGroupID(groupID, t)
		if err == nil {
			return domain, nil
		}
		if cl.domainGroupIdx == oldGroupIdx {
			plog.Errorf("no useful domain!")
			return nil, fmt.Errorf("no useful domain!")
		}
	}
}

func (cl *ControllerLogic) getDomainFromGroupID(groupID, t int64) (*DomainInfo, error) {
	v := cl.domainMap[groupID]
	if v != nil {
		if v.groupInfo.Status == DOMAIN_STATUS_OK {
			if len(v.domainList.DomainList) > 0 {
				oldDomainIdx := v.idx
				for {
					if v.domainList.DomainList[v.idx].Status == DOMAIN_STATUS_OK {
						resultIdx := v.idx
						v.idx = (v.idx + 1) % int64(len(v.domainList.DomainList))
						var domain string
						if cl.cfg.IfUrlEncoding {
							ok := false
							for _, gv := range cl.cfg.BaiduGroups {
								if gv == groupID {
									domain = BaiduEncoding(v.domainList.DomainList[resultIdx].Domain)
									ok = true
									break
								}
							}
							if !ok {
								for _, gv := range cl.cfg.ZhihuGroups {
									if gv == groupID {
										domain = ZhihuEncoding(v.domainList.DomainList[resultIdx].Domain)
										ok = true
										break
									}
								}
							}
							if !ok {
								domain = v.domainList.DomainList[resultIdx].Domain
							}
						} else {
							domain = v.domainList.DomainList[resultIdx].Domain
						}
						result := &DomainInfo{
							ID:      v.domainList.DomainList[resultIdx].ID,
							GroupID: v.domainList.DomainList[resultIdx].GroupID,
							Domain:  domain,
							Status:  v.domainList.DomainList[resultIdx].Status,
							Time:    v.domainList.DomainList[resultIdx].Time,
						}
						if t == DOMAIN_GROUP_TYPE_JUMP {
							ifHasShowGroup := false
							for _, jv := range v.groupInfo.ShowGroupList {
								jvg := cl.domainMap[jv]
								if jvg != nil {
									if jvg.groupInfo.Status == DOMAIN_STATUS_OK {
										ifHasShowGroup = true
										result.ShowGroupID = jvg.groupInfo.ID
										break
									}
								}
							}
							if ifHasShowGroup {
								return result, nil
							}
							return nil, fmt.Errorf("no useful jump domain!")
						}
						return result, nil
					}
					v.idx = (v.idx + 1) % int64(len(v.domainList.DomainList))
					if v.idx == oldDomainIdx {
						break
					}
				}
			}
		}
	}

	return nil, fmt.Errorf("no useful domain!")
}

func (cl *ControllerLogic) GetContent(id, contentGroupID int64, clientIP string) (*RealContentInfo, error) {
	cl.Lock()
	defer cl.Unlock()

	if contentGroupID != 0 {
		list := cl.contentMap[contentGroupID]
		if list == nil {
			return nil, fmt.Errorf("no this[%d] content group!", contentGroupID)
		}
		rci := &RealContentInfo{
			ContentGroupID: contentGroupID,
			ContentUrl:     list.groupInfo.JsonUrl,
			IfForceShare:   true,
			IfShowAds:      true,
		}
		v := cl.domainMap[id]
		if v != nil {
			rci.IfForceShare = (v.groupInfo.ShareStatus == 0)
			rci.IfShowAds = (v.groupInfo.AdsStatus == 0)
			if v.groupInfo.Status != DOMAIN_STATUS_OK {
				rci.IfOffLine = true
			}
		}
		if cl.detector.Check(&detector.DetectorInfo{GroupID: id, IP: clientIP}) {
			rci.IfForceShare = false
		}

		return rci, nil
	}

	var idx int64
	idx = -1
	if len(cl.contentGroupList) > 0 {
		idx = cl.contentGroupIdx
		cl.contentGroupIdx = (cl.contentGroupIdx + 1) % int64(len(cl.contentGroupList))
	}
	if idx == -1 {
		return nil, fmt.Errorf("no content group!")
	}
	list := cl.contentMap[cl.contentGroupList[idx]]
	if list == nil {
		return nil, fmt.Errorf("content map error!")
	}
	rci := &RealContentInfo{
		ContentGroupID: list.groupInfo.ID,
		ContentUrl:     list.groupInfo.JsonUrl,
		IfForceShare:   true,
		IfShowAds:      true,
	}
	v := cl.domainMap[id]
	if v != nil {
		rci.IfForceShare = (v.groupInfo.ShareStatus == 0)
		rci.IfShowAds = (v.groupInfo.AdsStatus == 0)
		if v.groupInfo.Status != DOMAIN_STATUS_OK {
			rci.IfOffLine = true
		}
	}
	if cl.detector.Check(&detector.DetectorInfo{GroupID: id, IP: clientIP}) {
		rci.IfForceShare = false
	}

	return rci, nil
}

func touchCacheDir() {
	fi, err := os.Stat(CacheDir)
	if err != nil {
		if os.IsExist(err) == false {
			os.MkdirAll(CacheDir, 0777)
		}
	} else {
		if fi.IsDir() == false {
			os.MkdirAll(CacheDir, 0777)
		}
	}
}

func setupLogging(cfg *config.Config) {
	capnslog.SetGlobalLogLevel(capnslog.INFO)
	if cfg.Debug {
		capnslog.SetGlobalLogLevel(capnslog.DEBUG)
	}
}
