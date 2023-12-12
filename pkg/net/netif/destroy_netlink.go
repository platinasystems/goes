// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/net/netlink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

func (nif *NetIf) Destroy(ctx context.Context) error {
	nl, err := netlink.Open()
	if err != nil {
		return err
	}
	defer nl.Close()

	req, msg := netlink.ExpandMsgHdr(nil)
	req.Type = rtnetlink.RTM_DELLINK
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, msg := netlink.ExpandIfInfoMsg(msg)
	ifinfo.Family = af.UNSPEC
	ifinfo.Index = int32(nif.Index)
	if err = nl.Request(msg); err == nil {
		err = nl.Wait(ctx, req.SEQ)
	}
	return err
}
