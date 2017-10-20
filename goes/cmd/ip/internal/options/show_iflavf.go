// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"net"

	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

func (opt *Options) ShowIflaVf(b []byte) {
	var vf rtnl.IflaVf

	nl.IndexAttrByType(vf[:], b)

	printflag := func(h string, t uint16) {
		if v := rtnl.IflaVfFlagPtr(vf[t]); v != nil {
			if v.Setting != ^uint32(0) {
				opt.Print(h)
				if v.Setting != 0 {
					opt.Print("on")
				} else {
					opt.Print("off")
				}
			}
		}
	}

	vfmac := rtnl.IflaVfMacPtr(vf[rtnl.IFLA_VF_MAC])
	if vfmac == nil {
		return
	}
	opt.Println()
	opt.Print("    vf ", vfmac.Vf)
	opt.Print(" MAC ", net.HardwareAddr(vfmac.Mac[:6]))
	vfvlan := rtnl.IflaVfVlanPtr(vf[rtnl.IFLA_VF_VLAN])
	if vfvlan != nil {
		if vfvlan.Vlan != 0 {
			opt.Print(", vlan ", vfvlan.Vlan)
		}
		if vfvlan.Qos != 0 {
			opt.Print(", qos ", vfvlan.Qos)
		}
	}
	vftxrate := rtnl.IflaVfTxRatePtr(vf[rtnl.IFLA_VF_TX_RATE])
	if vftxrate != nil {
		if vftxrate.Rate != 0 {
			opt.Print(", tx rate ", vftxrate.Rate, " (Mbps)")
		}
	}
	vfrate := rtnl.IflaVfRatePtr(vf[rtnl.IFLA_VF_RATE])
	if vfrate != nil {
		if vfrate.MaxTxRate != 0 {
			opt.Print(", max_tx_rate ",
				vfrate.MaxTxRate, "Mbps")
		}
		if vfrate.MinTxRate != 0 {
			opt.Print(", min_tx_rate ",
				vfrate.MinTxRate, "Mbps")
		}
	}
	printflag(", spoof checking ", rtnl.IFLA_VF_SPOOFCHK)
	vflinkstate := rtnl.IflaVfLinkStatePtr(vf[rtnl.IFLA_VF_LINK_STATE])
	if vflinkstate != nil {
		opt.Print(", link-state ")
		s, found := rtnl.IflaVfLinkStateName[vflinkstate.LinkState]
		if !found {
			s = "unknown"
		}
		opt.Print(s)
	}
	printflag(", trust ", rtnl.IFLA_VF_TRUST)
	if opt.Flags.ByName["-s"] && len(vf[rtnl.IFLA_VF_STATS]) > 0 {
		var vfstats rtnl.IflaVfStats
		nl.IndexAttrByType(vfstats[:], vf[rtnl.IFLA_VF_STATS])
		opt.ShowVfStats(vfstats[:])
	}
}
