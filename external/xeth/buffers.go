// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/external/xeth/internal"
)

type buffer interface {
	bytes() []byte
	pointer() unsafe.Pointer
	pool()
}

type b64 []byte
type b128 []byte
type b1024 []byte
type page []byte
type jumbo []byte

var PageSize = syscall.Getpagesize()

var buffers = struct {
	b64,
	b128,
	b1024,
	page,
	jumbo sync.Pool
}{
	b64: sync.Pool{
		New: func() interface{} {
			return b64(make([]byte, 64, 64))
		},
	},
	b128: sync.Pool{
		New: func() interface{} {
			return b128(make([]byte, 128, 128))
		},
	},
	b1024: sync.Pool{
		New: func() interface{} {
			return b1024(make([]byte, 1024, 1024))
		},
	},
	page: sync.Pool{
		New: func() interface{} {
			return page(make([]byte, PageSize, PageSize))
		},
	},
	jumbo: sync.Pool{
		New: func() interface{} {
			return jumbo(make([]byte, internal.SizeofJumboFrame,
				internal.SizeofJumboFrame))
		},
	},
}

func newBuffer(n int) buffer {
	switch {
	case n < 64:
		buf := buffers.b64.Get().(b64)
		buf = buf[:n]
		return buf
	case n < 128:
		buf := buffers.b128.Get().(b128)
		buf = buf[:n]
		return buf
	case n < 1024:
		buf := buffers.b1024.Get().(b1024)
		buf = buf[:n]
		return buf
	case n < PageSize:
		buf := buffers.page.Get().(page)
		buf = buf[:n]
		return buf
	case n < internal.SizeofJumboFrame:
		buf := buffers.jumbo.Get().(jumbo)
		buf = buf[:n]
		return buf
	default:
		panic("requested an oversized buffer")
	}
}

func cloneBuffer(b []byte) buffer {
	clone := newBuffer(len(b))
	copy(clone.bytes(), b)
	return clone
}

func (buf b64) bytes() []byte   { return []byte(buf) }
func (buf b128) bytes() []byte  { return []byte(buf) }
func (buf b1024) bytes() []byte { return []byte(buf) }
func (buf page) bytes() []byte  { return []byte(buf) }
func (buf jumbo) bytes() []byte { return []byte(buf) }

func (buf b64) pointer() unsafe.Pointer {
	return unsafe.Pointer(&buf.bytes()[0])
}

func (buf b128) pointer() unsafe.Pointer {
	return unsafe.Pointer(&buf.bytes()[0])
}

func (buf b1024) pointer() unsafe.Pointer {
	return unsafe.Pointer(&buf.bytes()[0])
}

func (buf page) pointer() unsafe.Pointer {
	return unsafe.Pointer(&buf.bytes()[0])
}

func (buf jumbo) pointer() unsafe.Pointer {
	return unsafe.Pointer(&buf.bytes()[0])
}

func (buf b64) pool()   { buffers.b64.Put(buf[:cap(buf)]) }
func (buf b128) pool()  { buffers.b128.Put(buf[:cap(buf)]) }
func (buf b1024) pool() { buffers.b1024.Put(buf[:cap(buf)]) }
func (buf page) pool()  { buffers.page.Put(buf[:cap(buf)]) }
func (buf jumbo) pool() { buffers.jumbo.Put(buf[:cap(buf)]) }
