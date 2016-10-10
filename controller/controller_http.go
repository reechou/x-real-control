package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

type XHttpServer struct {
	logic *ControllerLogic
	hs    *HttpSrv
}

type HttpHandler func(rsp http.ResponseWriter, req *http.Request) (interface{}, error)

func NewXHttpServer(addr string, port int, logic *ControllerLogic) *XHttpServer {
	xhs := &XHttpServer{
		hs: &HttpSrv{
			HttpAddr: addr,
			HttpPort: port,
			Routers:  make(map[string]http.HandlerFunc),
		},
		logic: logic,
	}
	xhs.registerHandlers()

	return xhs
}

func (xhs *XHttpServer) Run() {
	xhs.hs.Run()
}

func (xhs *XHttpServer) registerHandlers() {
	xhs.hs.Route("/", xhs.Index)

	xhs.hs.Route("/domain/add_domain_group", xhs.httpWrap(xhs.addDomainGroup))
	xhs.hs.Route("/domain/add_domain", xhs.httpWrap(xhs.addDomain))
	xhs.hs.Route("/domain/get_domain_groups", xhs.httpWrap(xhs.getDomainGroup))
	xhs.hs.Route("/domain/get_domain_group_detail", xhs.httpWrap(xhs.getDomainGroupDetail))
	xhs.hs.Route("/domain/get_domain_list", xhs.httpWrap(xhs.getDomainList))
	xhs.hs.Route("/domain/setting_domain_group", xhs.httpWrap(xhs.settingDomainGroup))
	xhs.hs.Route("/domain/off_domain", xhs.httpWrap(xhs.offDomain))
	xhs.hs.Route("/domain/set_domain_status", xhs.httpWrap(xhs.setDomainStatus))
	xhs.hs.Route("/domain/get_url", xhs.httpWrap(xhs.getURL))
	xhs.hs.Route("/domain/add_content_group", xhs.httpWrap(xhs.addContentGroup))
	xhs.hs.Route("/domain/get_content_group_detail", xhs.httpWrap(xhs.getContentGroupDetail))
	xhs.hs.Route("/domain/add_video_content", xhs.httpWrap(xhs.addVideoContent))
	xhs.hs.Route("/domain/get_content_group", xhs.httpWrap(xhs.getContentGroup))
	xhs.hs.Route("/domain/get_content_list", xhs.httpWrap(xhs.getContentList))
	xhs.hs.Route("/domain/get_data", xhs.httpWrap(xhs.getData))

	xhs.hs.Route("/domain/get_all_domains", xhs.getAllDomains)
}

func (xhs *XHttpServer) httpWrap(handler HttpHandler) func(rsp http.ResponseWriter, req *http.Request) {
	f := func(rsp http.ResponseWriter, req *http.Request) {
		logURL := req.URL.String()
		start := time.Now()
		defer func() {
			plog.Debugf("[XHttpServer][httpWrap] http: request url[%s] use_time[%v]", logURL, time.Now().Sub(start))
		}()
		obj, err := handler(rsp, req)
		// check err
	HAS_ERR:
		rsp.Header().Set("Access-Control-Allow-Origin", "*")
		rsp.Header().Set("Access-Control-Allow-Methods", "POST")
		rsp.Header().Set("Access-Control-Allow-Headers", "x-requested-with,content-type")

		if err != nil {
			plog.Debugf("[XHttpServer][httpWrap] http: request url[%s] error: %v", logURL, err)
			code := 500
			errMsg := err.Error()
			if strings.Contains(errMsg, "Permission denied") || strings.Contains(errMsg, "ACL not found") {
				code = 403
			}
			rsp.WriteHeader(code)
			rsp.Write([]byte(errMsg))
			return
		}

		// return json object
		if obj != nil {
			var buf []byte
			buf, err = json.Marshal(obj)
			if err != nil {
				goto HAS_ERR
			}
			rsp.Header().Set("Content-Type", "application/json")
			rsp.Write(buf)
		}
	}
	return f
}

func (xhs *XHttpServer) Index(rsp http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		rsp.WriteHeader(404)
		return
	}
	rsp.Write([]byte("Haunt Agent"))
}

func (xhs *XHttpServer) decodeBody(req *http.Request, out interface{}, cb func(interface{}) error) error {
	var raw interface{}
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&raw); err != nil {
		return err
	}

	if cb != nil {
		if err := cb(raw); err != nil {
			return err
		}
	}

	return mapstructure.Decode(raw, out)
}
