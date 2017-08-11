// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (opt *Options) ShowIflaVf(b []byte) {
	var vf rtnl.IflaVf

	rtnl.IndexAttrByType(vf[:], b)

	vfmac := rtnl.IflaVfMacPtr(&vf)
	if vfmac == nil {
		return
	}
	opt.Println()
	opt.Print("    vf ", vfmac.Vf)
	opt.Print(" MAC ", net.HardwareAddr(vfmac.Mac[:6]))
	if vfvlan := rtnl.IflaVfVlanPtr(&vf); vfvlan != nil {
		if vfvlan.Vlan != 0 {
			opt.Print(", vlan ", vfvlan.Vlan)
		}
		if vfvlan.Qos != 0 {
			opt.Print(", qos ", vfvlan.Qos)
		}
	}
	if vftxrate := rtnl.IflaVfTxRatePtr(&vf); vftxrate != nil {
		if vftxrate.Rate != 0 {
			opt.Print(", tx rate ", vftxrate.Rate, " (Mbps)")
		}
	}
	if vfrate := rtnl.IflaVfRatePtr(&vf); vfrate != nil {
		if vfrate.MaxTxRate != 0 {
			opt.Print(", max_tx_rate ",
				vfrate.MaxTxRate, "Mbps")
		}
		if vfrate.MinTxRate != 0 {
			opt.Print(", min_tx_rate ",
				vfrate.MinTxRate, "Mbps")
		}
	}
	if vfspoofchk := rtnl.IflaVfSpoofchkPtr(&vf); vfspoofchk != nil {
		if vfspoofchk.Setting != ^uint32(0) {
			opt.Print(", spoof checking ")
			if vfspoofchk.Setting != 0 {
				opt.Print("on")
			} else {
				opt.Print("off")
			}
		}
	}
	if vflinkstate := rtnl.IflaVfLinkStatePtr(&vf); vflinkstate != nil {
		opt.Print(", link-state ")
		s, found := map[uint32]string{
			rtnl.IFLA_VF_LINK_STATE_AUTO:    "auto",
			rtnl.IFLA_VF_LINK_STATE_ENABLE:  "enable",
			rtnl.IFLA_VF_LINK_STATE_DISABLE: "disable",
		}[vflinkstate.LinkState]
		if !found {
			s = "unknown"
		}
		opt.Print(s)
	}
	if vftrust := rtnl.IflaVfTrustPtr(&vf); vftrust != nil {
		if vftrust.Setting != ^uint32(0) {
			opt.Print(", trust ")
			if vftrust.Setting != 0 {
				opt.Print("on")
			} else {
				opt.Print("off")
			}
		}
	}
	if opt.Flags.ByName["-s"] && len(vf[rtnl.IFLA_VF_STATS]) > 0 {
		var vfstats rtnl.IflaVfStats
		rtnl.IndexAttrByType(vfstats[:], vf[rtnl.IFLA_VF_STATS])
		opt.ShowVfStats(vfstats[:])
	}
}
