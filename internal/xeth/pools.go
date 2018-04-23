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

const CacheLineSize = 64

var PageSize = syscall.Getpagesize()

var size = struct{ small, medium, large int }{
	small:  64,
	medium: PageSize,
	large:  SizeofJumboFrame,
}

type pools struct {
	small, medium, large sync.Pool
}

var Pool = pools{
	small: sync.Pool{
		New: func() interface{} {
			return make([]byte, size.small, size.small)
		},
	},
	medium: sync.Pool{
		New: func() interface{} {
			return make([]byte, size.medium, size.medium)
		},
	},
	large: sync.Pool{
		New: func() interface{} {
			return make([]byte, size.large, size.large)
		},
	},
}

func (p *pools) Get(n int) []byte {
	var buf []byte
	switch {
	case n <= size.small:
		buf = p.small.Get().([]byte)
	case n <= size.medium:
		buf = p.medium.Get().([]byte)
	case n <= size.large:
		buf = p.large.Get().([]byte)
	default:
		panic(fmt.Errorf("can't pool %d byte buffer", n))
	}
	// Optimised by compiler
	for i := range buf {
		buf[i] = 0
	}
	return buf[:n]
}

func (p *pools) Put(buf []byte) {
	switch cap(buf) {
	case size.small:
		p.small.Put(buf)
	case size.medium:
		p.medium.Put(buf)
	case size.large:
		p.large.Put(buf)
	default:
		panic(fmt.Errorf("unexpected buf cap: %d", cap(buf)))
	}
}
