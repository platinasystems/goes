// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import "github.com/platinasystems/goes/internal/nl/rtnl"

func (opt *Options) ShowIfStats(val []byte) {
	var ifstats64 rtnl.IfStats64
	opt.Println()
	opt.Nprint(4)
	if len(val) >= rtnl.SizeofIfStats64 {
		ifstats64 = *rtnl.IfStats64Attr(val)
	} else if len(val) >= rtnl.SizeofIfStats {
		ifstats32 := *rtnl.IfStatsAttr(val)
		for i := 0; i < rtnl.N_link_stat; i++ {
			ifstats64[i] = uint64(ifstats32[i])
		}
	} else {
		opt.Print("can't show these stats: ", val)
		return
	}
	opt.Println("RX: bytes  packets  errors  dropped overrun mcast")
	opt.Nprint(4)
	opt.Nprint(11, Stat(ifstats64[rtnl.Rx_bytes]))
	opt.Nprint(9, Stat(ifstats64[rtnl.Rx_packets]))
	opt.Nprint(8, Stat(ifstats64[rtnl.Rx_errors]))
	opt.Nprint(8, Stat(ifstats64[rtnl.Rx_dropped]))
	opt.Nprint(8, Stat(ifstats64[rtnl.Rx_over_errors]))
	opt.Println(Stat(ifstats64[rtnl.Multicast]))
	opt.Nprint(4)
	opt.Println("TX: bytes  packets  errors  dropped carrier collsns")
	opt.Nprint(4)
	opt.Nprint(11, Stat(ifstats64[rtnl.Tx_bytes]))
	opt.Nprint(9, Stat(ifstats64[rtnl.Tx_packets]))
	opt.Nprint(8, Stat(ifstats64[rtnl.Tx_errors]))
	opt.Nprint(8, Stat(ifstats64[rtnl.Tx_dropped]))
	opt.Nprint(8, Stat(ifstats64[rtnl.Tx_carrier_errors]))
	opt.Print(Stat(ifstats64[rtnl.Collisions]))
}
