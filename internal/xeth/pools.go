/* Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */
package xeth

import (
	"fmt"
	"sync"
	"syscall"
)

const SizeofSbHdrSetStat = SizeofSbHdr + SizeofSbSetStat

var PageSize = syscall.Getpagesize()

type pools struct {
	sbSetStat, page, jumbo sync.Pool
}

var Pool = pools{
	sbSetStat: sync.Pool{
		New: func() interface{} {
			return make([]byte, SizeofSbHdrSetStat)
		},
	},
	page: sync.Pool{
		New: func() interface{} {
			return make([]byte, PageSize)
		},
	},
	jumbo: sync.Pool{
		New: func() interface{} {
			return make([]byte, SizeofJumboFrame)
		},
	},
}

func (p *pools) Get(n int) []byte {
	var buf []byte
	switch {
	case n == SizeofSbHdrSetStat:
		buf = p.sbSetStat.Get().([]byte)
	case n <= PageSize:
		buf = p.page.Get().([]byte)
	case n < SizeofJumboFrame:
		buf = p.jumbo.Get().([]byte)
	default:
		panic("can't pool > jumbo frame")
	}
	// Optimised by compiler
	for i := range buf {
		buf[i] = 0
	}
	return buf[:n]
}

func (p *pools) Put(buf []byte) {
	switch cap(buf) {
	case SizeofSbHdrSetStat:
		p.sbSetStat.Put(buf)
	case PageSize:
		buf = buf[:cap(buf)]
		p.page.Put(buf)
	case SizeofJumboFrame:
		buf = buf[:cap(buf)]
		p.jumbo.Put(buf)
	default:
		panic(fmt.Errorf("unexpected buf cap: %d", cap(buf)))
	}
}
