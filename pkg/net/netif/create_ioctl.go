// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"context"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netioctl"
)

const CreateParameters = ""

var Cloneable = []string{
	"feth",
	"bridge",
	"bond",
	"vlan",
	"pflog",
	"gif",
	"iptap",
	"pktap",
}

func Create(ctx context.Context, name string, args ...string) (*NetIf, error) {
	nifs, err := List(ctx)
	if err != nil {
		return nil, egress.Mark(err)
	}
	before := make(map[int]*NetIf)
	for _, nif := range nifs {
		before[nif.Index] = nif
	}
	req := netioctl.NewIfReqNothing(name)
	err = netioctl.Inet(syscall.SIOCIFCREATE2, req)
	if err != nil {
		return nil, egress.Mark(err)
	}
	after, err := List(ctx)
	if err != nil {
		return nil, egress.Mark(err)
	}
	for _, nif := range after {
		if _, existed := before[nif.Index]; !existed {
			return nif, nil
		}
	}
	return nil, egress.Mark(ErrNotFound)
}
