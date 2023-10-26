// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import "github.com/platinasystems/goes/v2/pkg/net/netlink"

func (nif *Netif) Destroy() error {
	nl, err := netlink.Open()
	if err != nil {
		return err
	}
	defer nl.Close()

	req, msg := netlink.ExpandNlMsghdr(nil)
	req.Type = netlink.RTM_DELLINK
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifinfo, msg := netlink.ExpandIfInfomsg(msg)
	ifinfo.Family = netlink.AF_UNSPEC
	ifinfo.Index = int32(nif.Index)
	if err = nl.Request(msg); err == nil {
		err = nl.Wait(req.Seq)
	}
	return err
}
