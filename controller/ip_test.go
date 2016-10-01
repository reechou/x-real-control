package controller

import (
	"testing"
	"github.com/wangtuanjie/ip17mon"
	"fmt"
)

func TestIP(t *testing.T) {
	if err := ip17mon.Init("/Users/reezhou/Desktop/xman/src/github.com/reechou/x-real-control/17monipdb.dat"); err != nil {
		panic(err)
	}
	loc, err := ip17mon.Find("60.177.43.30")
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	fmt.Println(loc)
}
