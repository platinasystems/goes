// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package native

import (
	"encoding/binary"
	"math"
)

var Endian = binary.NativeEndian

type Uint16 [2]byte
type Uint32 [4]byte
type Uint64 [8]byte

func (ne *Uint16) Bytes() []byte { return ne[:] }
func (ne *Uint32) Bytes() []byte { return ne[:] }
func (ne *Uint64) Bytes() []byte { return ne[:] }

func (ne *Uint16) Put(v uint16) { Endian.PutUint16(ne.Bytes(), v) }
func (ne *Uint32) Put(v uint32) { Endian.PutUint32(ne.Bytes(), v) }
func (ne *Uint64) Put(v uint64) { Endian.PutUint64(ne.Bytes(), v) }

func (ne *Uint16) Value() uint16 { return Endian.Uint16(ne.Bytes()) }
func (ne *Uint32) Value() uint32 { return Endian.Uint32(ne.Bytes()) }
func (ne *Uint64) Value() uint64 { return Endian.Uint64(ne.Bytes()) }

type Float32 Uint32
type Float64 Uint64

func (ne *Float32) Bytes() []byte { return (*Uint32)(ne).Bytes() }
func (ne *Float64) Bytes() []byte { return (*Uint64)(ne).Bytes() }

func (ne *Float32) Put(v float32) { (*Uint32)(ne).Put(math.Float32bits(v)) }
func (ne *Float64) Put(v float64) { (*Uint64)(ne).Put(math.Float64bits(v)) }

func (ne *Float32) Uint32() uint32 { return (*Uint32)(ne).Value() }
func (ne *Float64) Uint64() uint64 { return (*Uint64)(ne).Value() }

func (ne *Float32) Value() float32 { return math.Float32frombits(ne.Uint32()) }
func (ne *Float64) Value() float64 { return math.Float64frombits(ne.Uint64()) }
