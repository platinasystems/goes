// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package chunk provides yet another allocator for long running applications
// to avoid GC through a select, custom, or capacity matched free pools.
// Unlike [sync.Pool.Put], [Free]'d data is never abandoned to the GC,
// it's always prepended to the matching free pool.
package chunk

import (
	"os"
	"sync"
	"sync/atomic"
)

var (
	Cap16  = &Pool{Cap: 16}
	Cap64  = &Pool{Cap: 64}
	Cap256 = &Pool{Cap: 256}
	Cap512 = &Pool{Cap: 512}
	Cap1K  = &Pool{Cap: 1 << 10}
	Cap4K  = &Pool{Cap: 4 << 10}
	Cap8K  = &Pool{Cap: 8 << 10}
	Cap16K = &Pool{Cap: 16 << 10}
)

// Pools may be expanded but must remain ordered by Cap.
var Pools = []*Pool{
	Cap16,
	Cap64,
	Cap256,
	Cap512,
	Cap1K,
	Cap4K,
	Cap8K,
	Cap16K,
}

var Page = sync.OnceValue(func() *Pool {
	switch n := os.Getpagesize(); n {
	case 4 << 10:
		return Cap4K
	case 8 << 10:
		return Cap8K
	case 16 << 10:
		return Cap16K
	default:
		return Cap4K
	}
})

func Alloc(n int) []byte {
	for _, c := range Pools {
		if n < c.Cap {
			return c.Alloc(n)
		}
	}
	return make([]byte, n)
}

func Free(data []byte) {
	for _, c := range Pools {
		if cap(data) == c.Cap {
			c.Free(data)
			break
		}
	}
}

type Pool struct {
	Cap  int
	free struct {
		data, ref atomic.Pointer[ref]
	}
}

type ref struct {
	next *ref
	data []byte
}

func (p *Pool) Alloc(n int) []byte {
	if n > p.Cap {
		return make([]byte, n)
	}
	r := p.free.data.Load()
	for r != nil {
		if !p.free.data.CompareAndSwap(r, r.next) {
			r = p.free.data.Load()
			continue
		}
		data := r.data[:n]
		r.data = r.data[:0]
		r.next = p.free.ref.Load()
		for !p.free.ref.CompareAndSwap(r.next, r) {
			r.next = p.free.ref.Load()
		}
		return data
	}
	return make([]byte, n, p.Cap)
}

func (p *Pool) Free(data []byte) {
	if cap(data) != p.Cap {
		return
	}
	r := p.free.ref.Load()
	for r != nil {
		if p.free.ref.CompareAndSwap(r, r.next) {
			break
		}
		r = p.free.ref.Load()
	}
	if r == nil {
		r = new(ref)
	}
	r.data = data[:cap(data)]
	r.next = p.free.data.Load()
	for !p.free.data.CompareAndSwap(r.next, r) {
		r.next = p.free.data.Load()
	}
}
