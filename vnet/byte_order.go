// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

// Network byte order helpers.
type Uint16 uint16
type Uint32 uint32
type Uint64 uint64

func (x Uint16) ToHost() uint16   { return swap16(uint16(x)) }
func (x Uint16) FromHost() Uint16 { return Uint16(swap16(uint16(x))) }
func (x *Uint16) Set(v uint)      { *x = Uint16(swap16(uint16(v))) }

func (x Uint32) ToHost() uint32   { return swap32(uint32(x)) }
func (x Uint32) FromHost() Uint32 { return Uint32(swap32(uint32(x))) }
func (x *Uint32) Set(v uint)      { *x = Uint32(swap32(uint32(v))) }

func (x Uint64) ToHost() uint64   { return swap64(uint64(x)) }
func (x Uint64) FromHost() Uint64 { return Uint64(swap64(uint64(x))) }
func (x *Uint64) Set(v uint)      { *x = Uint64(swap64(uint64(v))) }
