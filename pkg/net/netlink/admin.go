// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netlink

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/iflink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

func Admin(
	ctx context.Context,
	ifname string, with,
	without iflink.NetDeviceFlag,
) error {
	nl, err := Open()
	if err != nil {
		return egress.Mark(err)
	}
	defer nl.Close()

	ifindex, err := nl.IfIndex(ctx, ifname)
	if err != nil {
		return egress.Mark(err)
	}
	hdr, req := Expand[MsgHdr](nil)
	hdr.Type = rtnetlink.RTM_NEWLINK
	hdr.Flags = NLM_F_REQUEST | NLM_F_ACK
	ifinfo, req := Expand[rtnetlink.IfInfoMsg](req)
	ifinfo.Family = af.UNSPEC
	ifinfo.Index = ifindex
	ifinfo.Change = uint32(with | without)
	ifinfo.Flags |= uint32(with)
	ifinfo.Flags &^= uint32(without)
	if err = nl.Request(req); err == nil {
		err = nl.Wait(ctx, hdr.SEQ)
	}
	return err
}
