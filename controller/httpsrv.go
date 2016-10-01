package controller

import (
	"fmt"
	"net/http"
)

type HttpSrv struct {
	HttpAddr string
	HttpPort int
	Routers  map[string]http.HandlerFunc
}

func (hs *HttpSrv) Route(pattern string, f http.HandlerFunc) {
	hs.Routers[pattern] = f
}

func (hs *HttpSrv) Run() {
	addr := hs.HttpAddr
	if hs.HttpPort != 0 {
		addr = fmt.Sprintf("%s:%d", hs.HttpAddr, hs.HttpPort)
	}
	for p, f := range hs.Routers {
		http.Handle(p, f)
	}
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		panic(err)
	}
}
