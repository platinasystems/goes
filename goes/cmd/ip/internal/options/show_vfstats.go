// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

func (opt *Options) ShowVfStats(vfstats [][]byte) {
	opt.Println()
	opt.Nprint(4)
	opt.Println("RX: bytes  packets  mcast   bcast")
	opt.Nprint(4)
	for _, x := range []struct {
		w, i int
	}{
		{11, rtnl.IFLA_VF_STATS_RX_BYTES},
		{9, rtnl.IFLA_VF_STATS_RX_PACKETS},
		{8, rtnl.IFLA_VF_STATS_MULTICAST},
		{8, rtnl.IFLA_VF_STATS_BROADCAST},
	} {
		opt.Nprint(x.w, Stat(nl.Uint64(vfstats[x.i])))
	}
	opt.Println()
	opt.Nprint(4)
	opt.Println("TX: bytes  packets")
	opt.Nprint(4)
	for _, x := range []struct {
		w, i int
	}{
		{11, rtnl.IFLA_VF_STATS_TX_BYTES},
		{9, rtnl.IFLA_VF_STATS_TX_PACKETS},
	} {
		opt.Nprint(x.w, Stat(nl.Uint64(vfstats[x.i])))
	}
}
