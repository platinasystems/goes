// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"net/netip"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
)

type crate struct {
	ap  netip.AddrPort
	box box.Box
}

var inventory = sync.Pool{
	New: func() any {
		return &crate{
			box: box.New(),
		}
	},
}

func newCrate() *crate {
	crate := inventory.Get().(*crate)
	crate.box = crate.box.Expand()
	return crate
}

// Non-blocking put to channel. If channel is full, return to inventory.
func (c *crate) Put(ch chan<- *crate) {
	select {
	case ch <- c:
	default:
		inventory.Put(c)
	}
}

func (c *crate) Write(data []byte) (int, error) {
	n := len(data)
	if n > (cap(c.box) - len(c.box)) {
		return 0, ErrOverrun
	}
	c.box = append(c.box, data...)
	return n, nil
}
