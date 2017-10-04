// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/internal/sysconf"
)

func (opt *Options) ShowNeigh(b []byte, ifnames map[int32]string) {
	var nda rtnl.Nda
	nda.Write(b)
	msg := rtnl.NdMsgPtr(b)

	dst := nda[rtnl.NDA_DST]
	if len(dst) == 0 {
		return
	}

	opt.Print(net.IP(dst), " dev ")
	if name, found := ifnames[msg.Index]; found {
		opt.Print(name)
	} else {
		opt.Print(msg.Index)
	}
	if lladdr := nda[rtnl.NDA_LLADDR]; lladdr != nil {
		opt.Print(" lladdr ", net.HardwareAddr(lladdr[:6]))
	}
	if opt.Flags.ByName["-s"] {
		if val := nda[rtnl.NDA_CACHEINFO]; len(val) > 0 {
			ci := rtnl.NdaCacheInfoPtr(val)
			if ci != nil {
				if ci.RefCnt > 0 {
					opt.Print(" ref ", ci.RefCnt)
				}
				hz := sysconf.Hz()
				opt.Print(" used ",
					uint64(ci.Used)/hz, "/",
					uint64(ci.Confirmed)/hz, "/",
					uint64(ci.Updated))
			}
		}
		if val := nda[rtnl.NDA_PROBES]; len(val) > 0 {
			opt.Print(" probes ", rtnl.Uint32(val))
		}
	}
	{
		sep := " "
		for _, x := range []struct {
			flag uint16
			name string
		}{
			{rtnl.NUD_INCOMPLETE, "incomplete"},
			{rtnl.NUD_REACHABLE, "reachable"},
			{rtnl.NUD_STALE, "stale"},
			{rtnl.NUD_DELAY, "delay"},
			{rtnl.NUD_PROBE, "probe"},
			{rtnl.NUD_FAILED, "failed"},
			{rtnl.NUD_NOARP, "noarp"},
			{rtnl.NUD_PERMANENT, "permanent"},
		} {
			if (msg.State & x.flag) == x.flag {
				opt.Print(sep, x.name)
				sep = ","
			}
		}
	}
}
