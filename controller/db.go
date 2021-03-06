package controller

import (
	"strconv"
	"strings"

	"github.com/reechou/x-real-control/utils"
)

type ControllerDB struct {
	db *utils.MysqlController
}

func NewControllerDB(cfg *utils.MysqlInfo) (*ControllerDB, error) {
	cdb := &ControllerDB{
		db: utils.NewMysqlController(),
	}
	err := cdb.db.InitMysql(cfg)
	if err != nil {
		plog.Error("Mysql init error: %v\n", err)
		return nil, err
	}

	return cdb, nil
}

func (cdb *ControllerDB) InsertDomainGroup(info *DomainGroupInfo) error {
	id, err := cdb.db.Insert("insert into domain_group(name,status,share_status,ads_status) values(?,?,?,?)", info.Name, info.Status, info.ShareStatus, info.AdsStatus)
	if err != nil {
		return err
	}
	info.ID = id
	return nil
}

func (cdb *ControllerDB) InsertDomain(info *DomainInfo) error {
	id, err := cdb.db.Insert("insert into domain(group_id,domain,status) values(?,?,?)", info.GroupID, info.Domain, info.Status)
	if err != nil {
		return err
	}
	info.ID = id
	return nil
}

func (cdb *ControllerDB) InsertContentGroup(info *ContentGroupInfo) error {
	id, err := cdb.db.Insert("insert into content_group(name,type) values(?,?)", info.Name, info.Type)
	if err != nil {
		return err
	}
	info.ID = id
	return nil
}

func (cdb *ControllerDB) InsertContent(info *ContentInfo) error {
	id, err := cdb.db.Insert("insert into content(group_id,value,type) values(?,?,?)", info.GroupID, info.Value, info.Type)
	if err != nil {
		return err
	}
	info.ID = id
	return nil
}

func (cdb *ControllerDB) GetAllDomain() ([]string, error) {
	rows, err := cdb.db.FetchRows("select domain from domain where status=0 group by domain")
	if err != nil {
		return nil, err
	}
	var list []string
	for _, v := range *rows {
		list = append(list, v["domain"])
	}
	return list, nil
}

func (cdb *ControllerDB) GetDomainGroupFromID(info *DomainGroupInfo) error {
	row, err := cdb.db.FetchRow("select name,status,share_status,ads_status,type,show_group_list,time from domain_group where id=?", info.ID)
	if err != nil {
		return err
	}
	status, err := strconv.ParseInt((*row)["status"], 10, 0)
	if err != nil {
		plog.Errorf("GetDomainGroupFromID parse status[%s] error: %v\n", (*row)["status"], err)
		return err
	}
	shareStatus, err := strconv.ParseInt((*row)["share_status"], 10, 0)
	if err != nil {
		plog.Errorf("GetDomainGroupFromID parse share_status[%s] error: %v\n", (*row)["share_status"], err)
		return err
	}
	adsStatus, err := strconv.ParseInt((*row)["ads_status"], 10, 0)
	if err != nil {
		plog.Errorf("GetDomainGroupFromID parse ads_status[%s] error: %v\n", (*row)["ads_status"], err)
		return err
	}
	t, err := strconv.ParseInt((*row)["type"], 10, 0)
	if err != nil {
		plog.Errorf("GetDomainGroupFromID parse type[%s] error: %v\n", (*row)["type"], err)
		return err
	}
	info.Name = (*row)["name"]
	info.Status = status
	info.ShareStatus = shareStatus
	info.AdsStatus = adsStatus
	info.Type = t
	info.Time = (*row)["time"]
	if (*row)["show_group_list"] != "" && (*row)["show_group_list"] != info.ShowListStr {
		info.ShowListStr = (*row)["show_group_list"]
		showList := strings.Split((*row)["show_group_list"], ",")
		info.ShowGroupList = nil
		for _, v := range showList {
			sId, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				plog.Errorf("GetDomainGroupFromID show_group_list[%s] strconv error: %v", v, err)
				continue
			}
			info.ShowGroupList = append(info.ShowGroupList, sId)
		}
	}

	return nil
}

func (cdb *ControllerDB) GetDomainGroupList(maxID int64) ([]*DomainGroupInfo, int64, error) {
	rows, err := cdb.db.FetchRows("select id,name,status,share_status,ads_status,type,show_group_list,time from domain_group where id>?", maxID)
	if err != nil {
		return nil, 0, err
	}
	list := make([]*DomainGroupInfo, 0)
	newMaxID := maxID
	for _, v := range *rows {
		id, err := strconv.ParseInt(v["id"], 10, 0)
		if err != nil {
			continue
		}
		status, err := strconv.ParseInt(v["status"], 10, 0)
		if err != nil {
			continue
		}
		shareStatus, err := strconv.ParseInt(v["share_status"], 10, 0)
		if err != nil {
			continue
		}
		adsStatus, err := strconv.ParseInt(v["ads_status"], 10, 0)
		if err != nil {
			continue
		}
		t, err := strconv.ParseInt(v["type"], 10, 0)
		if err != nil {
			continue
		}

		if id > newMaxID {
			newMaxID = id
		}
		info := &DomainGroupInfo{
			ID:          id,
			Name:        v["name"],
			Status:      status,
			ShareStatus: shareStatus,
			AdsStatus:   adsStatus,
			Type:        t,
			Time:        v["time"],
		}
		if v["show_group_list"] != "" {
			info.ShowListStr = v["show_group_list"]
			showList := strings.Split(v["show_group_list"], ",")
			for _, v := range showList {
				sId, err := strconv.ParseInt(v, 10, 0)
				if err != nil {
					plog.Errorf("GetDomainGroupFromID show_group_list[%s] strconv error: %v", v, err)
					continue
				}
				info.ShowGroupList = append(info.ShowGroupList, sId)
			}
		}
		list = append(list, info)
	}
	return list, newMaxID, nil
}

func (cdb *ControllerDB) GetDomainList(list *DomainList) error {
	rows, err := cdb.db.FetchRows("select id,domain,status,time,UNIX_TIMESTAMP(time) as utime from domain where group_id=?", list.GroupID)
	if err != nil {
		return err
	}
	for _, v := range *rows {
		id, err := strconv.ParseInt(v["id"], 10, 0)
		if err != nil {
			continue
		}
		status, err := strconv.ParseInt(v["status"], 10, 0)
		if err != nil {
			continue
		}
		uTime, err := strconv.ParseInt(v["utime"], 10, 0)
		if err != nil {
			continue
		}
		if uTime > list.UpdateTime {
			list.UpdateTime = uTime
		}
		info := &DomainInfo{
			ID:      id,
			GroupID: list.GroupID,
			Domain:  v["domain"],
			Status:  status,
			Time:    v["time"],
		}
		list.DomainList = append(list.DomainList, info)
	}
	return nil
}

func (cdb *ControllerDB) GetContentGroupFromID(info *ContentGroupInfo) error {
	row, err := cdb.db.FetchRow("select name,json_url,type,main_content,time,UNIX_TIMESTAMP(time) as utime from content_group where id=?", info.ID)
	if err != nil {
		return err
	}
	t, err := strconv.ParseInt((*row)["type"], 10, 0)
	if err != nil {
		return err
	}
	utime, err := strconv.ParseInt((*row)["utime"], 10, 0)
	if err != nil {
		return err
	}
	info.Name = (*row)["name"]
	info.JsonUrl = (*row)["json_url"]
	info.Type = t
	info.Time = (*row)["time"]
	info.UpdateTime = utime
	if (*row)["main_content"] != "" {
		mainList := strings.Split((*row)["main_content"], ",")
		info.MainContent = nil
		for _, v := range mainList {
			cId, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				plog.Errorf("main content[%s] strconv error: %v", v, err)
				continue
			}
			info.MainContent = append(info.MainContent, cId)
		}
		plog.Debugf("content_group_id[%d] main_content: %v\n", info.ID, info.MainContent)
	}

	return nil
}

func (cdb *ControllerDB) GetContentGroupList(maxID int64) ([]*ContentGroupInfo, int64, error) {
	rows, err := cdb.db.FetchRows("select id,name,json_url,time from content_group where id>?", maxID)
	if err != nil {
		return nil, 0, err
	}
	list := make([]*ContentGroupInfo, 0)
	newMaxID := maxID
	for _, v := range *rows {
		id, err := strconv.ParseInt(v["id"], 10, 0)
		if err != nil {
			continue
		}
		if id > newMaxID {
			newMaxID = id
		}
		info := &ContentGroupInfo{
			ID:      id,
			Name:    v["name"],
			JsonUrl: v["json_url"],
			Time:    v["time"],
		}
		list = append(list, info)
	}
	return list, newMaxID, nil
}

func (cdb *ControllerDB) GetContentList(list *ContentList) error {
	rows, err := cdb.db.FetchRows("select id,value,type,time,UNIX_TIMESTAMP(time) as utime from content where group_id=?", list.GroupID)
	if err != nil {
		return err
	}
	for _, v := range *rows {
		id, err := strconv.ParseInt(v["id"], 10, 0)
		if err != nil {
			continue
		}
		cType, err := strconv.ParseInt(v["type"], 10, 0)
		if err != nil {
			continue
		}
		uTime, err := strconv.ParseInt(v["utime"], 10, 0)
		if err != nil {
			continue
		}
		if uTime > list.UpdateTime {
			list.UpdateTime = uTime
		}
		info := &ContentInfo{
			ID:      id,
			GroupID: list.GroupID,
			Value:   v["value"],
			Type:    cType,
			Time:    v["time"],
		}
		list.ContentList = append(list.ContentList, info)
	}
	return nil
}

func (cdb *ControllerDB) UpdateDomainStatus(info *DomainInfo) error {
	_, err := cdb.db.Exec("update domain set status=? where id=?", info.Status, info.ID)
	if err != nil {
		return err
	}
	return nil
}

func (cdb *ControllerDB) UpdateDomainsStatus(info *DomainInfo) error {
	_, err := cdb.db.Exec("update domain set status=? where domain=?", info.Status, info.Domain)
	if err != nil {
		return err
	}
	return nil
}

func (cdb *ControllerDB) UpdateDomainGroupStatus(info *DomainGroupInfo) error {
	_, err := cdb.db.Exec("update domain_group set status=?,share_status=?,ads_status=? where id=?", info.Status, info.ShareStatus, info.AdsStatus, info.ID)
	if err != nil {
		return err
	}
	return nil
}

func (cdb *ControllerDB) UpdateContentJsonUrl(info *ContentGroupInfo) error {
	_, err := cdb.db.Exec("update content_group set json_url=? where id=?", info.JsonUrl, info.ID)
	if err != nil {
		return err
	}
	return nil
}
