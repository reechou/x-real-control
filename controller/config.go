package controller

import (
	"flag"
	"fmt"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-ini/ini"
	"github.com/reechou/x-real-control/utils"
)

type AliyunOss struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	Bucket          string
	Url             string
	aliyunClient    *oss.Client
}

type config struct {
	ConfigPath string

	Debug bool

	ListenAddr string
	ListenPort int

	IfStartTimer bool

	utils.MysqlInfo
	AliyunOss
}

func NewConfig() *config {
	c := new(config)
	initFlag(c)

	if c.ConfigPath == "" {
		plog.Errorf("Hawk must run with config file, please check.\n")
		os.Exit(0)
	}

	cfg, err := ini.Load(c.ConfigPath)
	if err != nil {
		plog.Errorf("ini[%s] load error: %v\n", c.ConfigPath, err)
		os.Exit(1)
	}
	cfg.BlockMode = false
	err = cfg.MapTo(c)
	if err != nil {
		plog.Errorf("config MapTo error: %v\n", err)
		os.Exit(1)
	}
	plog.Info(c)

	return c
}

func initFlag(c *config) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	v := fs.Bool("v", false, "Print version and exit")
	fs.StringVar(&c.ConfigPath, "c", "", "wx-controller config file.")

	fs.Parse(os.Args[1:])
	fs.Usage = func() {
		fmt.Println("Usage: hawk -c hawk.ini")
		fmt.Printf("\nglobal flags:\n")
		fs.PrintDefaults()
	}

	if *v {
		fmt.Println("wx-controller: 0.0.1")
		os.Exit(0)
	}
}
