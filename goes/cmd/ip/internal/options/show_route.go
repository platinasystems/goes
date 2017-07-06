// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (opt *Options) ShowRoute(b []byte, ifnames map[int32]string) {
	var rta rtnl.Rta
	rta.Write(b)
	msg := rtnl.RtMsgPtr(b)
	detailed := opt.Flags.ByName["-d"]
	if val := rta[rtnl.RTA_DST]; len(val) > 0 {
		dstip := net.IP(rta[rtnl.RTA_DST])
		if msg.Dst_len != rtnl.AfBits[msg.Family] {
			opt.Print(dstip, "/", msg.Dst_len)
		} else {
			opt.Print(dstip)
		}
	} else if msg.Dst_len > 0 {
		opt.Print("0/", msg.Dst_len)
	} else {
		opt.Print("default")
	}
	if val := rta[rtnl.RTA_SRC]; len(val) > 0 {
		srcip := net.IP(val)
		if msg.Src_len != rtnl.AfBits[msg.Family] {
			opt.Print(" from ", srcip, "/", msg.Src_len)
		} else {
			opt.Print(" from ", srcip)
		}
	} else if msg.Src_len > 0 {
		opt.Print(" from 0/", msg.Src_len)
	}
	if val := rta[rtnl.RTA_NEWDST]; len(val) > 0 {
		opt.Print(" as to ", net.IP(val))
	}
	if val := rta[rtnl.RTA_ENCAP]; len(val) > 0 {
		opt.Print(" FIXME encap ", val)
	}
	if val := rta[rtnl.RTA_GATEWAY]; len(val) > 0 {
		opt.Print(" via ", net.IP(val))
	}
	if val := rta[rtnl.RTA_VIA]; len(val) > 0 {
		opt.Print(" FIXME via ", val)
	}
	if val := rta[rtnl.RTA_OIF]; len(val) > 0 {
		oif := rtnl.Int32(val)
		if name, found := ifnames[oif]; found {
			opt.Print(" dev ", name)
		} else {
			opt.Print(" dev ", oif)
		}
	}
	if val := rta[rtnl.RTA_TABLE]; len(val) > 0 {
		t := rtnl.Uint32(val)
		if t != uint32(rtnl.RT_TABLE_MAIN) || detailed {
			opt.Print(" table ", rtnl.RtTableName(t))
		}
	}
	if msg.Protocol != rtnl.RTPROT_UNSPEC {
		if msg.Protocol != rtnl.RTPROT_BOOT || detailed {
			opt.Print(" prot ",
				rtnl.RtProtName[msg.Protocol])
		}
	}
	if msg.Scope != rtnl.RT_SCOPE_UNIVERSE || detailed {
		opt.Print(" scope ", rtnl.RtScopeName[msg.Scope])
	}
	if val := rta[rtnl.RTA_PREFSRC]; len(val) > 0 {
		opt.Print(" src ", net.IP(val))
	}
	if val := rta[rtnl.RTA_PRIORITY]; len(val) > 0 {
		opt.Print(" metric ", rtnl.Uint32(val))
	}
	// FIXME RTNH_F_{DEAD,ONLINK,PERVASIVE,OFFLOAD,NOTIFY,LINKDOWN}
	if val := rta[rtnl.RTA_MARK]; len(val) > 0 {
		opt.Print(" mark ", rtnl.Uint32(val))
	}
	// FIXME RTA_FLOW
	// FIXME CLONED?
	// FIXME RTA_METRICS
	// FIXME RTA_IIF
	// FIXME RTA_MULTIPATH
	// FIXME RTA_PREF
}
