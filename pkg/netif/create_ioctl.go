// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/netioctl"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
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
		return nil, xerrors.Mark(err)
	}
	before := make(map[int]*NetIf)
	for _, nif := range nifs {
		before[nif.Index] = nif
	}
	req := netioctl.NewIfReqNothing(name)
	err = netioctl.Inet(SIOCIFCREATE, req)
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	after, err := List(ctx)
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	for _, nif := range after {
		if _, existed := before[nif.Index]; !existed {
			return nif, nil
		}
	}
	return nil, xerrors.NotFound(name)
}
