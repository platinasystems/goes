// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nl

import "syscall"

var fdSetBits int

func FdSetBits() int {
	if fdSetBits == 0 {
		var set syscall.FdSet
		fdSetBits = syscall.FD_SETSIZE / len(set.Bits)
	}
	return fdSetBits
}

func FD_SET(p *syscall.FdSet, i int) {
	p.Bits[i/FdSetBits()] |= 1 << uint(i%FdSetBits())
}

func FD_ISSET(p *syscall.FdSet, i int) bool {
	return (p.Bits[i/FdSetBits()] & (1 << uint(i%FdSetBits()))) != 0
}

func FD_ZERO(p *syscall.FdSet) {
	for i := range p.Bits {
		p.Bits[i] = 0
	}
}
