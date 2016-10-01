package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
		Type    int64       `json:"type"`
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
		Type:    info.Type,
	}
	err = xhs.logic.cdb.InsertContent(content)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("add content failed: %v", err)
		return response, nil
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

func (xhs *XHttpServer) offDomainGroup(rsp http.ResponseWriter, req *http.Request) (interface{}, error) {
	response := &Response{Code: RES_OK}
	var info DomainGroupInfo
	if err := xhs.decodeBody(req, &info, nil); err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("Request decode failed: %v", err)
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

	data, err := xhs.logic.GetContent(info.GroupID)
	if err != nil {
		response.Code = RES_ERR
		response.Msg = fmt.Sprintf("get content failed: %v", err)
		return response, nil
	}
	response.Data = data

	return response, nil
}
