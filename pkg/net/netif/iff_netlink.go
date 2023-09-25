// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
)

func Admin(ifname string, with, without IFF) error {
	nl, err := netlink.Open()
	if err != nil {
		return err
	}
	defer nl.Close()

	var nif *Netif
	nifs, err := list(nl)
	if err != nil {
		return egress.Marked(err)
	}
	for _, v := range nifs {
		if v.Name == ifname {
			nif = v
		}
	}
	if nif == nil {
		return egress.Marked(ErrNotFound)
	}
	iflhdr, iflreq := netlink.Expand[netlink.NlMsghdr](nil)
	iflhdr.Type = netlink.RTM_NEWLINK
	iflhdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, iflreq := netlink.Expand[netlink.IfInfomsg](iflreq)
	ifinfo.Family = netlink.AF_UNSPEC
	ifinfo.Index = int32(nif.Index)
	ifinfo.Change = uint32(with | without)
	ifinfo.Flags |= uint32(with)
	ifinfo.Flags &^= uint32(without)
	return egress.Marked(nl.Request(iflreq, nil))
}

func Down(ifname string) error {
	return Admin(ifname, 0, IFF_UP)
}

func Up(ifname string) error {
	return Admin(ifname, IFF_UP, 0)
}
