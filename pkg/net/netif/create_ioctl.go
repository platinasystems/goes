// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
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

func Create(name string, args ...string) (*NetIf, error) {
	beforeNifs, beforeByIndex, beforeByName, err := List()
	if err != nil {
		return nil, egress.Marked(err)
	}
	_ = beforeNifs
	_ = beforeByName
	req := netioctl.NewIfReqNothing(name)
	err = netioctl.Inet(syscall.SIOCIFCREATE2, req)
	if err != nil {
		return nil, egress.Marked(err)
	}
	afterNifs, afterByIndex, afterByName, err := List()
	if err != nil {
		return nil, egress.Marked(err)
	}
	_ = afterByIndex
	_ = afterByName
	for _, nif := range afterNifs {
		if _, existed := beforeByIndex[nif.Index]; !existed {
			return nif, nil
		}
	}
	return nil, egress.Marked(ErrNotFound)
}
