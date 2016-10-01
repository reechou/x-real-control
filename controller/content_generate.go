package controller

import (
	"bytes"
	"encoding/json"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/reechou/x-real-control/utils"
)

type ContentGenerate struct {
	groupInfo *ContentGroupInfo
	w         *utils.TimingWheel
	logic     *ControllerLogic
	cdb       *ControllerDB

	updateTime int64
	aliyunInfo *AliyunOss

	stop chan struct{}
	done chan struct{}
}

func NewContentGenerate(groupInfo *ContentGroupInfo, cdb *ControllerDB, w *utils.TimingWheel, logic *ControllerLogic, aliyunInfo *AliyunOss) *ContentGenerate {
	cg := &ContentGenerate{
		groupInfo:  groupInfo,
		cdb:        cdb,
		w:          w,
		logic:      logic,
		aliyunInfo: aliyunInfo,
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}
	cg.init()
	go cg.run()

	return cg
}

func (cg *ContentGenerate) init() {
	rule1 := oss.CORSRule{
		AllowedOrigin: []string{"*"},
		AllowedMethod: []string{"PUT", "GET"},
		AllowedHeader: []string{},
		ExposeHeader:  []string{},
		MaxAgeSeconds: 200,
	}

	err := cg.aliyunInfo.aliyunClient.SetBucketCORS(cg.aliyunInfo.Bucket, []oss.CORSRule{rule1})
	if err != nil {
		plog.Panic("aliyun set oss cors rule error.", err)
	}
}

func (cg *ContentGenerate) Stop() {
	close(cg.stop)
	<-cg.done
}

func (cg *ContentGenerate) run() {
	plog.Infof("content group[%s][%d] start run.\n", cg.groupInfo.Name, cg.groupInfo.ID)
	for {
		select {
		case <-cg.w.Check(cg.groupInfo.ID):
			cg.onCheck()
		case <-cg.stop:
			close(cg.done)
			return
		}
	}
}

func (cg *ContentGenerate) onCheck() {
	list := &ContentList{
		GroupID: cg.groupInfo.ID,
	}
	err := cg.cdb.GetContentList(list)
	if err != nil {
		plog.Errorf("get content list error: %v\n", err)
		return
	}
	if list.UpdateTime > cg.updateTime {
		err = cg.saveAndPublish(list)
		if err != nil {
			plog.Errorf("save and publish error: %v\n", err)
			return
		}
		cg.updateTime = list.UpdateTime
	}
}

func (cg *ContentGenerate) saveAndPublish(list *ContentList) error {
	var dataList []interface{}
	for _, v := range list.ContentList {
		switch cg.groupInfo.Type {
		case CONTENT_TYPE_VIDEO:
			var info Video
			err := json.Unmarshal([]byte(v.Value), &info)
			if err != nil {
				continue
			}
			dataList = append(dataList, info)
		}
	}
	dataListBytes, err := json.Marshal(dataList)
	if err != nil {
		return err
	}
	filename := cg.groupInfo.Name + ".json"

	//result := string(dataListBytes)
	//f, err := os.OpenFile(CacheDir+"/"+filename, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	//if err != nil {
	//	return err
	//}
	//f.WriteString(result)
	//f.Close()
	//plog.Infof("save file[%s] success.\n", filename)

	bucket, err := cg.aliyunInfo.aliyunClient.Bucket(cg.aliyunInfo.Bucket)
	if err != nil {
		plog.Errorf("create bucket[%v] error: %v\n", cg.aliyunInfo, err)
		return err
	}
	err = bucket.PutObject(filename, bytes.NewReader(dataListBytes), []oss.Option{oss.ContentType("text/html"), oss.ContentEncoding("utf-8")}...)
	if err != nil {
		plog.Errorf("put object bucket[%v] error: %v\n", cg.aliyunInfo, err)
		return err
	}
	plog.Infof("aliyun publish file[%s] success.\n", filename)
	cg.groupInfo.JsonUrl = cg.aliyunInfo.Url + filename
	cg.cdb.UpdateContentJsonUrl(cg.groupInfo)

	return nil
}
