package tsc

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"

	"strings"
)

type portStatus struct {
	name          string
	live_link     bool
	signal_detect bool
	pmd_lock      bool
	sigdet_sts    uint16 `format:"0x%x"`
	speed         string
	Autonegotiate autoneg_status
	Cl72          cl72_status
}

type autoneg_status struct {
	enable bool
	done   bool
}

func (x *autoneg_status) String() (s string) {
	if x.enable {
		if x.done {
			s = "done"
		} else {
			s = "pending"
		}
	}
	return
}

type cl72_status uint16

func (x cl72_status) String() (s string) {
	if x&(1<<2) != 0 {
		s += "in-progress, "
	}
	if x&(1<<1) != 0 {
		s += "locked, "
	}
	if x&(1<<3) != 0 {
		s += "fail, "
	}
	if x&(1<<0) != 0 {
		s += "ready, "
	}
	s = strings.TrimRight(s, ", ")
	return
}

func (ss *switchSelect) showBcmPortStatus(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var ifs vnet.HwIfChooser
	ifs.Init(ss.Vnet)
	for !in.End() {
		switch {
		case in.Parse("%v", &ifs):
		default:
			err = cli.ParseError
			return
		}
	}
	stats := []portStatus{}
	ifs.Foreach(func(v *vnet.Vnet, r vnet.HwInterfacer) {
		var (
			p  m.Porter
			ok bool
		)
		if p, ok = r.(m.Porter); !ok {
			return
		}
		if !m.IsProvisioned(p) {
			return
		}
		laneMask := p.GetLaneMask()
		if p.GetPortCommon().IsManagement {
			phy := p.GetPhy().(*Tsce)
			laneMask.Foreach(func(l m.LaneMask) {
				stats = append(stats, phy.getMgmtStatus(p, l))
			})
		} else {
			phy := p.GetPhy().(*Tscf)
			laneMask.Foreach(func(l m.LaneMask) {
				stats = append(stats, phy.getStatus(p, l))
			})
		}
	})
	elib.TabulateWrite(w, stats)
	return
}
