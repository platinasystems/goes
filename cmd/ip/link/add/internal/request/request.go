// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package request

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/cmd/ip/internal/options"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
)

func New(opt *options.Options, args []string) (*Add, error) {
	err := opt.OnlyName(args)
	if err != nil {
		return nil, err
	}
	ifname := opt.Parms.ByName["name"]
	if len(args) == 1 {
		ifname = args[0]
	}
	if len(ifname) == 0 {
		return nil, fmt.Errorf("missing IFNAME")
	}
	if len(args) > 1 {
		return nil, fmt.Errorf("%v: unexpected", args[1:])
	}
	add := new(Add)
	add.Hdr.Type = rtnl.RTM_NEWLINK
	add.Hdr.Flags = nl.NLM_F_REQUEST | nl.NLM_F_ACK | nl.NLM_F_CREATE |
		nl.NLM_F_EXCL
	add.Msg.Family = rtnl.AF_UNSPEC
	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_IFNAME,
		nl.KstringAttr(ifname)})
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"address", rtnl.IFLA_ADDRESS},
		{"broadcast", rtnl.IFLA_BROADCAST},
	} {
		s := opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		mac, err := net.ParseMAC(s)
		if err != nil {
			return add, fmt.Errorf("%s: %q %v", x.name, s, err)
		}
		add.Attrs = append(add.Attrs, nl.Attr{x.t, nl.BytesAttr(mac)})
	}
	if s := opt.Parms.ByName["index"]; len(s) > 0 {
		if _, err := fmt.Sscan(s, &add.Msg.Index); err != nil {
			return add, fmt.Errorf("index: %q %v", s, err)
		}
	}
	if s := opt.Parms.ByName["link"]; len(s) > 0 {
		link, found := rtnl.If.IndexByName[s]
		if !found {
			return add, fmt.Errorf("link: %q not found", s)
		}
		add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINK,
			nl.Int32Attr(link)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"mtu", rtnl.IFLA_MTU},
		{"numtxqueues", rtnl.IFLA_NUM_TX_QUEUES},
		{"numrxqueues", rtnl.IFLA_NUM_RX_QUEUES},
		{"txqueuelen", rtnl.IFLA_TXQLEN},
	} {
		var u32 uint32
		s := opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		_, err := fmt.Sscan(s, &u32)
		if err != nil {
			return add, fmt.Errorf("%s: %q %v", x.name, s, err)
		}
		add.Attrs = append(add.Attrs, nl.Attr{x.t, nl.Uint32Attr(u32)})
	}
	return add, nil
}

type Add struct {
	Hdr   nl.Hdr
	Msg   rtnl.IfInfoMsg
	Attrs nl.Attrs
}

func (add *Add) Message() ([]byte, error) {
	return nl.NewMessage(add.Hdr, add.Msg, add.Attrs...)
}
