package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"strconv"
	"html/template"
)

func (xhs *XHttpServer) addDomainGroup(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	var info DomainGroupInfo
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	err := xhs.logic.cdb.InsertDomainGroup(&info)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("add domain group failed: %v", err)
		return response, nil
	}

	return response, nil
}

func (xhs *XHttpServer) addDomain(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	var info DomainInfo
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	err := xhs.logic.cdb.InsertDomain(&info)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("add domain failed: %v", err)
		return response, nil
	}

	return response, nil
}

func (xhs *XHttpServer) addContentGroup(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	var info ContentGroupInfo
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	err := xhs.logic.cdb.InsertContentGroup(&info)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("add content group failed: %v", err)
		return response, nil
	}

	return response, nil
}

func (xhs *XHttpServer) addVideoContent(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	type AddVideoReq struct {
		GroupID int64       `json:"groupID"`
		VInfo   interface{} `json:"video"`
	}
	result, err := ioutil.ReadAll(req.Body)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("add content group ioutil.ReadAll failed: %v", err)
		return response, nil
	}
	var info AddVideoReq
	err = json.Unmarshal(result, &info)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("add content group json unmarshal failed: %v", err)
		return response, nil
	}

	valueBytes, err := json.Marshal(&info.VInfo)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request json marshal failed: %v", err)
		return response, nil
	}

	content := &ContentInfo{
		GroupID: info.GroupID,
		Value:   string(valueBytes),
		Type:    CONTENT_TYPE_VIDEO,
	}
	err = xhs.logic.cdb.InsertContent(content)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("add content failed: %v", err)
		return response, nil
	}

	return response, nil
}

func (xhs *XHttpServer) getDomainGroupDetail(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	type GetDomainGroupReq struct {
		GroupID int64 `json:"groupID"`
	}
	var info GetDomainGroupReq
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	domainGroupInfo := &DomainGroupInfo{
		ID: info.GroupID,
	}
	err := xhs.logic.cdb.GetDomainGroupFromID(domainGroupInfo)
	if err != nil {
		plog.Errorf("get domain group detail error: %v\n", err)
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get domain group detail error: %v\n", err)
	} else {
		response.Data = domainGroupInfo
	}

	return response, nil
}

func (xhs *XHttpServer) getDomainGroup(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	list, _, err := xhs.logic.cdb.GetDomainGroupList(0)
	if err != nil {
		plog.Errorf("get domain groups error: %v\n", err)
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get domain groups error: %v\n", err)
	} else {
		response.Data = list
	}

	return response, nil
}

func (xhs *XHttpServer) getDomainList(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	var list DomainList
	if err := xhs.decodeBody(req, &list, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	err := xhs.logic.cdb.GetDomainList(&list)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get domain list failed: %v", err)
		return response, nil
	}
	response.Data = list

	return response, nil
}

func (xhs *XHttpServer) offDomain(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	var info DomainInfo
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	err := xhs.logic.cdb.UpdateDomainStatus(&info)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("off domain failed: %v", err)
		return response, nil
	}

	return response, nil
}

func (xhs *XHttpServer) settingDomainGroup(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	var info DomainGroupInfo
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	if info.ID == 0 {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("domain group id cannot be 0.")
		return response, nil
	}

	err := xhs.logic.cdb.UpdateDomainGroupStatus(&info)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("off domain group failed: %v", err)
		return response, nil
	}

	return response, nil
}

func (xhs *XHttpServer) getContentGroupDetail(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	type GetContentGroupReq struct {
		GroupID int64 `json:"groupID"`
	}
	var info GetContentGroupReq
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	contentGroupInfo := &ContentGroupInfo{
		ID: info.GroupID,
	}
	err := xhs.logic.cdb.GetContentGroupFromID(contentGroupInfo)
	if err != nil {
		plog.Errorf("get content group detail error: %v\n", err)
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get content group detail error: %v\n", err)
	} else {
		response.Data = contentGroupInfo
	}

	return response, nil
}

func (xhs *XHttpServer) getContentGroup(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	list, _, err := xhs.logic.cdb.GetContentGroupList(0)
	if err != nil {
		plog.Errorf("get content groups error: %v\n", err)
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get content groups error: %v\n", err)
	} else {
		response.Data = list
	}

	return response, nil
}

func (xhs *XHttpServer) getContentList(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	var list ContentList
	if err := xhs.decodeBody(req, &list, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	err := xhs.logic.cdb.GetContentList(&list)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get content list failed: %v", err)
		return response, nil
	}
	response.Data = list

	return response, nil
}

func (xhs *XHttpServer) getURL(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	type GetURLReq struct {
		GroupID int64 `json:"groupID"`
	}
	var info GetURLReq
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	data, err := xhs.logic.GetDomainInfo(info.GroupID)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get url failed: %v", err)
		return response, nil
	}
	response.Data = data

	return response, nil
}

func (xhs *XHttpServer) getData(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	type GetContentReq struct {
		GroupID int64 `json:"groupID"`
	}
	var info GetContentReq
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
		return response, nil
	}

	data, err := xhs.logic.GetContent(info.GroupID, strings.Split(req.RemoteAddr, ":")[0])
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get content failed: %v", err)
		return response, nil
	}
	response.Data = data

	return response, nil
}

func (xhs *XHttpServer) setDomainStatus(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	req.ParseForm()
	var domain string
	var status int64
	domainV := req.Form["domain"]
	if domainV == nil {
		plog.Errorf("set domain status domain is nil\n")
		return nil, nil
	}
	statusV := req.Form["status"]
	if statusV == nil {
		plog.Errorf("set domain status status is nil \n")
		return nil, nil
	}
	domain = domainV[0]
	status, _ = strconv.ParseInt(statusV[0], 10, 0)
	
	response := &Response{Code: RES_OK}
	info := &DomainInfo{
		Domain: domain,
		Status: status,
	}
	
	err := xhs.logic.cdb.UpdateDomainsStatus(info)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("set domain failed: %v", err)
		return response, nil
	}
	
	return response, nil
}

func (xhs *XHttpServer) getAllDomains(rsp http.ResponseWriter, req *http.Request) {
	//req.ParseForm()
	//var offset int64
	//var num int64
	//offsetV := req.Form["offset"]
	//if offsetV == nil {
	//	plog.Errorf("get all domains offset is nil\n")
	//	rsp.WriteHeader(500)
	//	rsp.Write([]byte("get all domains offset is nil."))
	//	return
	//}
	//numV := req.Form["num"]
	//if numV == nil {
	//	plog.Errorf("get all domains num is nil \n")
	//	rsp.WriteHeader(500)
	//	rsp.Write([]byte("get all domains num is nil."))
	//	return
	//}
	//offset, _ = strconv.ParseInt(offsetV[0], 10, 0)
	//num, _ = strconv.ParseInt(numV[0], 10, 0)
	
	list, err := xhs.logic.cdb.GetAllDomain()
	if err != nil {
		rsp.WriteHeader(500)
		rsp.Write([]byte("get all domain error."))
		return
	}
	type HtmlDomains struct {
		Title string
		Domains []string
	}
	htmlDomains := &HtmlDomains{
		Title: "domains",
		Domains: list,
	}
	tpl, err := template.New("domains.tpl").ParseFiles("/Users/reezhou/Desktop/xman/src/github.com/reechou/x-real-control/tpl/domains.tpl")
	if err != nil {
		fmt.Println(err)
		rsp.WriteHeader(500)
		rsp.Write([]byte("tpl parse error."))
		return
	}
	err = tpl.Execute(rsp, htmlDomains)
	if err != nil {
		rsp.WriteHeader(500)
		rsp.Write([]byte("tpl parse error."))
		return
	}
}
