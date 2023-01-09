// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package chunk

import "sync"

const (
	SizeBlock = 512
	SizeK     = 1 << 10
	SizeFrame = 1500
	Size2K    = 2 << 10
	Size4K    = 4 << 10
	Size8K    = 8 << 10
	SizeJumbo = 9728
	Size16K   = 16 << 10
	Size64K   = 64 << 10
	SizeM     = 1 << 20
)

func chunk(n int) []byte { return make([]byte, n, n) }

var (
	poolBlock = sync.Pool{New: func() any { return chunk(SizeBlock) }}
	poolK     = sync.Pool{New: func() any { return chunk(SizeK) }}
	poolFrame = sync.Pool{New: func() any { return chunk(SizeFrame) }}
	pool2K    = sync.Pool{New: func() any { return chunk(Size2K) }}
	pool4K    = sync.Pool{New: func() any { return chunk(Size4K) }}
	pool8K    = sync.Pool{New: func() any { return chunk(Size8K) }}
	poolJumbo = sync.Pool{New: func() any { return chunk(SizeJumbo) }}
	pool16K   = sync.Pool{New: func() any { return chunk(Size16K) }}
	pool64K   = sync.Pool{New: func() any { return chunk(Size64K) }}
	poolM     = sync.Pool{New: func() any { return chunk(SizeM) }}
)

// Free* will panic if cap(b) <= Size*
func FreeBlock(b []byte) { poolBlock.Put(b[:SizeBlock]) }
func FreeK(b []byte)     { poolK.Put(b[:SizeK]) }
func FreeFrame(b []byte) { poolFrame.Put(b[:SizeFrame]) }
func Free2K(b []byte)    { pool2K.Put(b[:Size2K]) }
func Free4K(b []byte)    { pool4K.Put(b[:Size4K]) }
func Free8K(b []byte)    { pool8K.Put(b[:Size8K]) }
func FreeJumbo(b []byte) { poolJumbo.Put(b[:SizeJumbo]) }
func Free16K(b []byte)   { pool16K.Put(b[:Size16K]) }
func Free64K(b []byte)   { pool64K.Put(b[:Size64K]) }
func FreeM(b []byte)     { poolM.Put(b[:SizeM]) }

// Unrecognized cap(b) are garbage collected.
func Free(b []byte) {
	switch cap(b) {
	case SizeBlock:
		FreeBlock(b)
	case SizeK:
		FreeK(b)
	case SizeFrame:
		FreeFrame(b)
	case Size2K:
		Free2K(b)
	case Size4K:
		Free4K(b)
	case Size8K:
		Free8K(b)
	case SizeJumbo:
		FreeJumbo(b)
	case Size16K:
		Free16K(b)
	case Size64K:
		Free64K(b)
	case SizeM:
		FreeM(b)
	}
}

func NewBlock() []byte { return poolBlock.Get().([]byte) }
func NewK() []byte     { return poolK.Get().([]byte) }
func NewFrame() []byte { return poolFrame.Get().([]byte) }
func New2K() []byte    { return pool2K.Get().([]byte) }
func New4K() []byte    { return pool4K.Get().([]byte) }
func New8K() []byte    { return pool8K.Get().([]byte) }
func NewJumbo() []byte { return poolJumbo.Get().([]byte) }
func New16K() []byte   { return pool16K.Get().([]byte) }
func New64K() []byte   { return pool64K.Get().([]byte) }
func NewM() []byte     { return poolM.Get().([]byte) }

func New(n int) []byte {
	switch {
	case n <= SizeBlock:
		return NewBlock()[:n]
	case n <= SizeK:
		return NewK()[:n]
	case n <= SizeFrame:
		return NewFrame()[:n]
	case n <= Size2K:
		return New2K()[:n]
	case n <= Size4K:
		return New4K()[:n]
	case n <= Size8K:
		return New8K()[:n]
	case n <= SizeJumbo:
		return NewJumbo()[:n]
	case n <= Size16K:
		return New16K()[:n]
	case n <= Size64K:
		return New64K()[:n]
	case n <= SizeM:
		return NewM()[:n]
	default:
		return make([]byte, n, n)
	}
}
