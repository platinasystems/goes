// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/parse"

	"unsafe"
)

type Rewrite struct {
	// Software interface to mark re-written packets with.
	Si Si

	// Node where packet will be rewritten.
	NodeIndex uint32

	// Next node to feed packet after rewrite.
	NextIndex uint32

	// Max packet size layer 3 (MTU) for output interface.
	// Used for MTU check after packet rewrite.
	// Avoids having to lookup egress interface's MTU.
	MaxL3PacketSize uint16

	// Number of bytes in rewrite data.
	dataLen uint16

	data [hw.BufferRewriteBytes]byte
}

func (r *Rewrite) String(v *Vnet) (lines []string) {
	swt := r.Si.GetType(v)
	lines = append(lines, r.Si.Name(v))
	lines = append(lines, swt.SwInterfaceRewriteString(v, r)...)
	return
}

func (r *Rewrite) ParseWithArgs(in *parse.Input, args *parse.Args) {
	v := args.Get().(*Vnet)
	var line parse.Input
	if !in.Parse("%v %l", &r.Si, v, &line) {
		in.ParseError()
	}
	sw := v.SwIf(r.Si)
	hw := v.SupHwIf(sw)
	h := v.HwIfer(hw.hi)
	h.ParseRewrite(r, &line)
}

func (r *Rewrite) Len() uint        { return uint(r.dataLen) }
func (r *Rewrite) SetLen(l uint)    { r.dataLen = uint16(l) }
func (r *Rewrite) SetData(d []byte) { r.dataLen = uint16(copy(r.data[:], d)) }
func (r *Rewrite) ResetData()       { r.SetData(nil) }
func (r *Rewrite) Data() []byte     { return r.data[:] }
func (r *Rewrite) Slice() []byte    { return r.data[:r.dataLen] }
func (r *Rewrite) AddData(p unsafe.Pointer, size uintptr) (l uintptr) {
	l = uintptr(r.dataLen)
	r.dataLen += uint16(size)
	for i := uintptr(0); i < size; i++ {
		r.data[l+i] = *(*uint8)(elib.PointerAdd(p, i))
	}
	return l + size
}
func (r *Rewrite) getData() []byte           { return r.data[:r.dataLen] }
func (r *Rewrite) GetData() unsafe.Pointer   { return unsafe.Pointer(&r.data[0]) }
func (r *Rewrite) SetMaxPacketSize(hw *HwIf) { r.MaxL3PacketSize = uint16(hw.maxPacketSize) }

func (v *Vnet) SetRewrite(rw *Rewrite, si Si, noder Noder, t PacketType, dstAddr []byte) {
	sw := v.SwIf(si)
	hw := v.SupHwIf(sw)
	h := v.HwIfer(hw.hi)
	n := noder.GetNode()
	rw.Si = si
	rw.NodeIndex = uint32(n.Index())
	x, _ := v.loop.AddNext(noder, h)
	rw.NextIndex = uint32(x)
	rw.MaxL3PacketSize = uint16(hw.maxPacketSize)
	h.SetRewrite(v, rw, t, dstAddr)
}

func PerformRewrite(r0 *Ref, rw0 *Rewrite) {
	r0.Advance(-int(rw0.dataLen))
	copy(r0.DataSlice(), rw0.getData())
}

func Perform2Rewrites(r0, r1 *Ref, rw0, rw1 *Rewrite) {
	r0.Advance(-int(rw0.dataLen))
	r1.Advance(-int(rw1.dataLen))
	copy(r0.DataSlice(), rw0.getData())
	copy(r1.DataSlice(), rw1.getData())
}
