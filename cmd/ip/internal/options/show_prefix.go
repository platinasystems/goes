// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"net"

	"github.com/platinasystems/goes/internal/nl/rtnl"
)

func (opt *Options) ShowPrefix(b []byte) {
	var prefixa rtnl.Prefixa
	prefixa.Write(b)
	msg := rtnl.PrefixMsgPtr(b)

	if msg.Family != rtnl.AF_INET6 {
		opt.Print("incorrect protocol family: ",
			rtnl.AfName(msg.Family))
		return
	}

	// FIXME type?

	if val := prefixa[rtnl.PREFIX_ADDRESS]; len(val) > 0 {
		opt.Print("prefix ", net.IP(val), "/", msg.Len, " ")
	}
	if name, found := rtnl.If.NameByIndex[int32(msg.IfIndex)]; found {
		opt.Print("dev ", name)
	} else {
		opt.Print("dev ", msg.IfIndex)
	}
	if (msg.Flags & rtnl.IF_PREFIX_ONLINK) != 0 {
		opt.Print("onlink ")
	}
	if (msg.Flags & rtnl.IF_PREFIX_AUTOCONF) != 0 {
		opt.Print("autoconf ")
	}
	if val := prefixa[rtnl.PREFIX_CACHEINFO]; len(val) > 0 {
		ci := rtnl.PrefixCacheInfoPtr(val)
		opt.Print("valid ", ci.ValidTime, " ")
		opt.Print("preferred ", ci.PreferredTime, " ")

	}
}
