package controller

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/reechou/x-real-control/config"
	"github.com/reechou/x-real-control/utils"
)

const (
	DOMAIN_STATUS_OK = iota
	DOMAIN_STATUS_DOWN
	DOMAIN_STATUS_OFF
)

type DomainCheckHealth struct {
	groupInfo  *DomainGroupInfo
	updateTime int64

	cfg         *config.Config
	checkUrlIdx int

	cdb   *ControllerDB
	w     *utils.TimingWheel
	logic *ControllerLogic

	client *http.Client

	stop chan struct{}
	done chan struct{}
}

func NewDomainCheckHealth(groupInfo *DomainGroupInfo, cdb *ControllerDB, w *utils.TimingWheel, logic *ControllerLogic, cfg *config.Config) *DomainCheckHealth {
	dch := &DomainCheckHealth{
		groupInfo: groupInfo,
		cfg:       cfg,
		cdb:       cdb,
		w:         w,
		logic:     logic,
		client:    &http.Client{},
		stop:      make(chan struct{}),
		done:      make(chan struct{}),
	}
	go dch.run()

	return dch
}

func (dch *DomainCheckHealth) Stop() {
	close(dch.stop)
	<-dch.done
}

func (dch *DomainCheckHealth) run() {
	plog.Infof("domain group[%s][%d] start run.\n", dch.groupInfo.Name, dch.groupInfo.ID)
	for {
		select {
		case <-dch.w.Check(dch.groupInfo.ID):
			dch.onCheck()
		case <-dch.stop:
			close(dch.done)
			return
		}
	}
}

func (dch *DomainCheckHealth) onCheck() {
	// get group
	err := dch.cdb.GetDomainGroupFromID(dch.groupInfo)
	if err != nil {
		plog.Errorf("oncheck group[%d] get domain group error: %v\n", dch.groupInfo.ID, err)
		return
	}
	if dch.groupInfo.Status != DOMAIN_STATUS_OK {
		plog.Infof("domain group[%s][%d] is setted offline.\n", dch.groupInfo.Name, dch.groupInfo.ID)
		return
	}
	//plog.Debugf("on check get group: %v\n", dch.groupInfo)

	// get list
	list := &DomainList{
		GroupID: dch.groupInfo.ID,
	}
	err = dch.cdb.GetDomainList(list)
	if err != nil {
		plog.Errorf("oncheck group[%d] get domain list error: %v\n", dch.groupInfo.ID, err)
		return
	}

	// check
	checkUpdate := false
	for _, v := range list.DomainList {
		if v.Status == DOMAIN_STATUS_DOWN {
			continue
		}
		ok := dch.checkHealthV2(v)
		if !ok {
			if v.Status != DOMAIN_STATUS_DOWN {
				v.Status = DOMAIN_STATUS_DOWN
				dch.cdb.UpdateDomainStatus(v)
				checkUpdate = true
			}
		}
	}

	// update
	if checkUpdate || (list.UpdateTime > dch.updateTime) {
		dch.logic.UpdateDomainGroup(dch.groupInfo, list)
	}
}

const (
	DOMAIN_CHECK_OK          = "[0]"
	DOMAIN_CHECK_GRAY        = "[1]"
	DOMAIN_CHECK_BLACK       = "[2]"
	DOMAIN_CHECK_QUERY_ERROR = "[3]"
)

func (dch *DomainCheckHealth) checkHealthV2(info *DomainInfo) bool {
	if len(dch.cfg.CheckDomainUrls) == 0 {
		return true
	}
	url := "http://" + dch.cfg.CheckDomainUrls[dch.checkUrlIdx] + "/mt.do?url=" + info.Domain
	dch.checkUrlIdx = (dch.checkUrlIdx + 1) % len(dch.cfg.CheckDomainUrls)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		plog.Errorf("check health[%s] error: %v\n", info.Domain, err)
		return false
	}

	rsp, err := dch.client.Do(req)
	defer func() {
		if rsp != nil {
			rsp.Body.Close()
		}
	}()
	if err != nil {
		plog.Errorf("check health[%s] error: %v\n", info.Domain, err)
		return false
	}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		plog.Errorf("check health[%s] error: %v\n", info.Domain, err)
		return false
	}
	rspBody = bytes.Replace(rspBody, []byte(" "), []byte(""), -1)
	rspBody = bytes.Replace(rspBody, []byte("\n"), []byte(""), -1)
	result := string(rspBody)
	switch result {
	case DOMAIN_CHECK_OK:
		return true
	}
	plog.Errorf("domain[%s] check health error, check result: %s\n", url, result)
	if result == DOMAIN_CHECK_GRAY || result == DOMAIN_CHECK_BLACK {
		return false
	}
	return true
}

type DomainHealthResponse struct {
	Status  int    `json:"status"`
	Msg     string `json:"msg"`
	Domain  string `json:"domain"`
	Endtime int64  `json:"endtime"`
}

const (
	DOMAIN_HEALTH_NOT_OK = 2
)

func (dch *DomainCheckHealth) checkHealth(info *DomainInfo) bool {
	return true

	url := "http://app.nf6688.com/___check___/" + info.Domain
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		plog.Errorf("check health[%s] error: %v\n", info.Domain, err)
		return false
	}

	rsp, err := dch.client.Do(req)
	defer func() {
		if rsp != nil {
			rsp.Body.Close()
		}
	}()
	if err != nil {
		plog.Errorf("check health[%s] error: %v\n", info.Domain, err)
		return false
	}
	rspBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		plog.Errorf("check health[%s] error: %v\n", info.Domain, err)
		return false
	}

	var response DomainHealthResponse
	err = json.Unmarshal(rspBody, &response)
	if err != nil {
		plog.Errorf("check health[%s] error: %v\n", info.Domain, err)
		return false
	}
	if response.Status == DOMAIN_HEALTH_NOT_OK {
		plog.Infof("group[%s][%d] domain[%s] check unhealth.\n", dch.groupInfo.Name, dch.groupInfo.ID, info.Domain)
		return false
	}
	return true
}
