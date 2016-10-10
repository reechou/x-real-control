package detector

import (
	"github.com/reechou/x-real-control/config"
	"github.com/wangtuanjie/ip17mon"
)

type IPFilter struct {
	cfg *config.IPFilterConfig
}

func NewIPFilter(cfg *config.IPFilterConfig) *IPFilter {
	f := &IPFilter{
		cfg: cfg,
	}
	f.init()

	return f
}

func (f *IPFilter) init() {
	if err := ip17mon.Init(f.cfg.IPDB); err != nil {
		plog.Panic("init ip db error:", err)
	}
}

func (f *IPFilter) Check(info *DetectorInfo) bool {
	loc, err := ip17mon.Find(info.IP)
	if err != nil {
		plog.Errorf("ip[%s] find error: %v\n", info.IP, err)
		return false
	}
	for _, v := range f.cfg.FilterLocation {
		if v == loc.City {
			plog.Debugf("ip[%s] check in[%s], check ok.\n", info.IP, v)
			return true
		}
	}

	return false
}
