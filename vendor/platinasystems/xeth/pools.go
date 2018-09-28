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

var PageSize = syscall.Getpagesize()

type BufPool struct {
	cap int
	sync.Pool
}

type BufPools []*BufPool

var Pool = BufPools{
	&BufPool{64, sync.Pool{
		New: func() interface{} {
			return make([]byte, 64, 64)
		},
	}},
	&BufPool{128, sync.Pool{
		New: func() interface{} {
			return make([]byte, 128, 128)
		},
	}},
	&BufPool{1024, sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024, 1024)
		},
	}},
	&BufPool{PageSize, sync.Pool{
		New: func() interface{} {
			return make([]byte, PageSize, PageSize)
		},
	}},
	&BufPool{SizeofJumboFrame, sync.Pool{
		New: func() interface{} {
			return make([]byte, SizeofJumboFrame, SizeofJumboFrame)
		},
	}},
}

func (pools BufPools) Get(n int) []byte {
	for _, pool := range pools {
		if n < pool.cap {
			buf := pool.Get().([]byte)
			// Optimised by compiler
			for i := range buf {
				buf[i] = 0
			}
			return buf[:n]
		}
	}
	panic(fmt.Errorf("no pool for %d byte buffer", n))
}

func (pools BufPools) Put(buf []byte) {
	for _, pool := range pools {
		if cap(buf) == pool.cap {
			pool.Put(buf)
			return
		}
	}
	panic(fmt.Errorf("no pool for %d cap buffer", cap(buf)))
}
