// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netlink

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

func Admin[With, Without integer.Int | integer.Uint](
	ctx context.Context,
	ifname string,
	with With,
	without Without,
) error {
	nl, err := Open()
	if err != nil {
		return xerrors.Mark(err)
	}
	defer nl.Close()

	ifindex, err := nl.IfIndex(ctx, ifname)
	if err != nil {
		return xerrors.Mark(err)
	}
	hdr, req := Expand[MsgHdr](nil)
	hdr.Type = rtnetlink.RTM_NEWLINK
	hdr.Flags = NLM_F_REQUEST | NLM_F_ACK
	ifinfo, req := Expand[rtnetlink.IfInfoMsg](req)
	ifinfo.Family = xnet.AF_UNSPEC
	ifinfo.Index = ifindex
	integer.Set(&ifinfo.Change, with)
	integer.Set(&ifinfo.Change, without)
	integer.Set(&ifinfo.Flags, with)
	integer.Reset(&ifinfo.Flags, without)
	if err = nl.Request(req); err == nil {
		err = nl.Wait(ctx, hdr.SEQ)
	}
	return err
}
