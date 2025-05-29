// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package chunk provides byte slice allocation w/ capacity matched pools.
package chunk

import (
	"context"
	"sync"
)

type Chunk = *[]byte

func mk(n uint) *[]byte {
	chunk := new([]byte)
	*chunk = make([]byte, n, n)
	return chunk
}

var (
	pool16  = sync.Pool{New: func() any { return mk(16) }}
	pool64  = sync.Pool{New: func() any { return mk(64) }}
	pool256 = sync.Pool{New: func() any { return mk(256) }}
	pool512 = sync.Pool{New: func() any { return mk(512) }}
	pool1K  = sync.Pool{New: func() any { return mk(1 << 10) }}
	pool2K  = sync.Pool{New: func() any { return mk(2 << 10) }}
	pool4K  = sync.Pool{New: func() any { return mk(4 << 10) }}
	pool8K  = sync.Pool{New: func() any { return mk(8 << 10) }}
)

func Discard(chunk *[]byte) {
	if chunk == nil {
		return
	}
	switch n := cap(*chunk); n {
	case 0:
	case 16:
		pool16.Put(chunk)
	case 64:
		pool64.Put(chunk)
	case 256:
		pool256.Put(chunk)
	case 512:
		pool512.Put(chunk)
	case 1 << 10:
		pool1K.Put(chunk)
	case 2 << 10:
		pool2K.Put(chunk)
	case 4 << 10:
		pool4K.Put(chunk)
	case 8 << 10:
		pool8K.Put(chunk)
	default:
		*chunk = (*chunk)[:0]
	}
}

func New(n uint) *[]byte {
	var chunk *[]byte
	if n <= 16 {
		chunk = pool16.Get().(*[]byte)
	} else if n <= 64 {
		chunk = pool64.Get().(*[]byte)
	} else if n <= 256 {
		chunk = pool256.Get().(*[]byte)
	} else if n <= 512 {
		chunk = pool512.Get().(*[]byte)
	} else if n <= 1<<10 {
		chunk = pool1K.Get().(*[]byte)
	} else if n <= 2<<10 {
		chunk = pool2K.Get().(*[]byte)
	} else if n <= 4<<10 {
		chunk = pool4K.Get().(*[]byte)
	} else if n <= 8<<10 {
		chunk = pool8K.Get().(*[]byte)
	} else {
		chunk = mk(n)
	}
	*chunk = (*chunk)[:n]
	return chunk
}

// Returns false and discards chunk if it isn't channeled before
// context is done.
func Queue(ctx context.Context, ch chan<- *[]byte, chunk *[]byte) bool {
	select {
	case <-ctx.Done():
		Discard(chunk)
		return false
	case ch <- chunk:
		return true
	}
}
