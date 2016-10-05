package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-ini/ini"
	"github.com/reechou/x-real-control/utils"
	"github.com/coreos/pkg/capnslog"
)

var plog = capnslog.NewPackageLogger("github.com/reezhou/x-real-control", "config")

type AliyunOss struct {
	Endpoint        string
	AccessKeyId     string
	AccessKeySecret string
	Bucket          string
	Url             string
	AliyunClient    *oss.Client
}

type IPFilterConfig struct {
	IPDB           string
	FilterLocation []string
}

type Config struct {
	ConfigPath string

	Debug bool

	ListenAddr string
	ListenPort int

	IfStartTimer  bool
	IfUrlEncoding bool
	
	BaiduGroups []int64
	ZhihuGroups []int64
	BaiduUrlGroup []string
	ZhihuUrlGroup []string

	utils.MysqlInfo
	AliyunOss
	IPFilterConfig
}

func NewConfig() *Config {
	c := new(Config)
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
	
	for _, v := range c.BaiduUrlGroup {
		groupId, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			continue
		}
		c.BaiduGroups = append(c.BaiduGroups, groupId)
	}
	for _, v := range c.ZhihuUrlGroup {
		groupId, err := strconv.ParseInt(v, 10, 0)
		if err != nil {
			continue
		}
		c.ZhihuGroups = append(c.ZhihuGroups, groupId)
	}
	
	plog.Info(c)

	return c
}

func initFlag(c *Config) {
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
