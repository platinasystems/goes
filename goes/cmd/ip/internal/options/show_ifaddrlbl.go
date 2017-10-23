// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"net"

	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

func (opt *Options) ShowIfAddrLbl(b []byte) {
	var ifal rtnl.Ifal
	var space string
	ifal.Write(b)
	msg := rtnl.IfAddrLblMsgPtr(b)

	if val := ifal[rtnl.IFAL_ADDRESS]; len(val) > 0 {
		opt.Print("prefix ", net.IP(val), "/", msg.PrefixLen)
		space = " "
	}

	if name, found := rtnl.If.NameByIndex[int32(msg.IfIndex)]; found {
		opt.Print(space, "dev ", name)
		space = " "
	}

	if val := ifal[rtnl.IFAL_LABEL]; len(val) > 0 {
		opt.Print(space, "label ", nl.Uint32(val))
	}
}
