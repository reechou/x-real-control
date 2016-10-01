package controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/reechou/x-real-control/utils"
)

const (
	DOMAIN_STATUS_OK = iota
	DOMAIN_STATUS_DOWN
	DOMAIN_STATUS_OFF
)

type DomainCheckHealth struct {
	groupInfo *DomainGroupInfo

	cdb   *ControllerDB
	w     *utils.TimingWheel
	logic *ControllerLogic

	client *http.Client

	stop chan struct{}
	done chan struct{}
}

func NewDomainCheckHealth(groupInfo *DomainGroupInfo, cdb *ControllerDB, w *utils.TimingWheel, logic *ControllerLogic) *DomainCheckHealth {
	dch := &DomainCheckHealth{
		groupInfo: groupInfo,
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
	for _, v := range list.DomainList {
		ok := dch.checkHealth(v)
		if !ok {
			if v.Status != DOMAIN_STATUS_DOWN {
				v.Status = DOMAIN_STATUS_DOWN
				dch.cdb.UpdateDomainStatus(v)
			}
		}
	}

	// update
	dch.logic.UpdateDomainGroup(dch.groupInfo, list)
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
