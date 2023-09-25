// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

const CreateParameters = `
Create Parameters
  FIXME
`

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

func Create(name string, args ...string) (*Netif, error) {
	nifs, err := List()
	if err != nil {
		return nil, egress.Marked(err)
	}
	existing := make(map[int]*Netif)
	for _, nif := range nifs {
		existing[nif.Index] = nif
	}
	inet, err := af.Open[af.Inet]()
	if err != nil {
		return nil, egress.Marked(err)
	}
	defer af.Close(inet)
	req := NewIfreq[Nothing](name)
	if err = IOCTL(inet, syscall.SIOCIFCREATE2, req); err != nil {
		return nil, egress.Marked(err)
	}
	if nifs, err = List(); err != nil {
		return nil, egress.Marked(err)
	}
	for _, nif := range nifs {
		if _, existed := existing[nif.Index]; !existed {
			return nif, nil
		}
	}
	return nil, egress.Marked(ErrNotFound)
}
