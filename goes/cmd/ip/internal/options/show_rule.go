// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (opt *Options) ShowRule(b []byte, ifnames map[int32]string) {
	var fra rtnl.Fra
	var hostlen uint8
	fra.Write(b)
	msg := rtnl.FibRuleMsgPtr(b)

	switch msg.Family {
	case rtnl.AF_INET6:
		hostlen = 128
	case rtnl.AF_INET:
		hostlen = 32
	}

	if val := fra[rtnl.FRA_PRIORITY]; len(val) > 0 {
		opt.Print(rtnl.Uint32(val), ":\t")
	} else {
		opt.Print("0:\t")
	}

	if (msg.Flags & rtnl.FIB_RULE_INVERT) != 0 {
		opt.Print("not ")
	}

	if val := fra[rtnl.FRA_SRC]; len(val) > 0 {
		if msg.Src_len != hostlen {
			opt.Print("from ", net.IP(val), "/", msg.Src_len, " ")
		} else {
			opt.Print("from ", net.IP(val), " ")
		}
	} else if msg.Src_len != 0 {
		opt.Print("from 0/", msg.Src_len, " ")
	} else {
		opt.Print("from all ")
	}

	if val := fra[rtnl.FRA_DST]; len(val) > 0 {
		if msg.Dst_len != hostlen {
			opt.Print("to ", net.IP(val), "/", msg.Dst_len, " ")
		} else {
			opt.Print("to ", net.IP(val), " ")
		}
	} else if msg.Dst_len != 0 {
		opt.Print("to 0/", msg.Dst_len, " ")
	}

	if msg.Tos != 0 {
		opt.Print("tos ", msg.Tos, " ")
	}

	vmark, vmask := fra[rtnl.FRA_FWMARK], fra[rtnl.FRA_FWMASK]
	if len(vmark) > 0 || len(vmask) > 0 {
		mark := rtnl.Uint32(vmark)
		mask := rtnl.Uint32(vmask)
		if len(vmask) > 0 && mask != ^uint32(0) {
			fmt.Printf("fwmark %#x/%#x ", mark, mask)
		} else {
			fmt.Printf("fwmark %#x ", mark)
		}
	}

	if val := fra[rtnl.FRA_IFNAME]; len(val) > 0 {
		opt.Print("iif ", rtnl.Kstring(val), " ")
		if (msg.Flags & rtnl.FIB_RULE_IIF_DETACHED) != 0 {
			opt.Print("[detatched] ")
		}
	}

	if val := fra[rtnl.FRA_OIFNAME]; len(val) > 0 {
		opt.Print("oif ", rtnl.Kstring(val), " ")
		if (msg.Flags & rtnl.FIB_RULE_OIF_DETACHED) != 0 {
			opt.Print("[detatched] ")
		}
	}

	if val := fra[rtnl.FRA_L3MDEV]; len(val) > 0 {
		if rtnl.Uint8(val) != 0 {
			opt.Print("lookup [l3mdev-table] ")
		}
	}

	if val := fra[rtnl.FRA_UID_RANGE]; len(val) > 0 {
		r := rtnl.FibRuleUidRangePtr(val)
		opt.Print("uidrange ", r.Start, "-", r.End, " ")
	}

	// FIXME table

	if val := fra[rtnl.FRA_FLOW]; len(val) > 0 {
		to := rtnl.Uint32(val)
		from := to >> 16
		to &= 0xFFFF
		if from != 0 {
			opt.Print("realms ", from, "/")
		}
		opt.Print(to, " ")
	}

	// FIXME RTN_NAT
}
