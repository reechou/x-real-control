package detector

import (
	"github.com/coreos/pkg/capnslog"
	"github.com/reechou/x-real-control/config"
)

var plog = capnslog.NewPackageLogger("github.com/reezhou/x-real-control", "detector")

type DetectorInfo struct {
	GroupID int64
	IP      string
}

type CheckFilter interface {
	Check(info *DetectorInfo) bool
}

type Detector struct {
	filters []CheckFilter
}

func NewDetector(cfg *config.Config) *Detector {
	d := &Detector{}
	d.filters = append(d.filters, NewIPFilter(&cfg.IPFilterConfig))

	return d
}

func (d *Detector) Check(info *DetectorInfo) bool {
	for _, v := range d.filters {
		if v.Check(info) {
			return true
		}
	}
	return false
}
