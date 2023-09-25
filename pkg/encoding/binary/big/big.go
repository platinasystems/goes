// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package big

import (
	"encoding/binary"
	"math"
)

var Endian = binary.BigEndian

type Uint16 [2]byte
type Uint32 [4]byte
type Uint64 [8]byte

func (be *Uint16) Bytes() []byte { return be[:] }
func (be *Uint32) Bytes() []byte { return be[:] }
func (be *Uint64) Bytes() []byte { return be[:] }

func (be *Uint16) Put(v uint16) { Endian.PutUint16(be.Bytes(), v) }
func (be *Uint32) Put(v uint32) { Endian.PutUint32(be.Bytes(), v) }
func (be *Uint64) Put(v uint64) { Endian.PutUint64(be.Bytes(), v) }

func (be *Uint16) Value() uint16 { return Endian.Uint16(be.Bytes()) }
func (be *Uint32) Value() uint32 { return Endian.Uint32(be.Bytes()) }
func (be *Uint64) Value() uint64 { return Endian.Uint64(be.Bytes()) }

type Float32 Uint32
type Float64 Uint64

func (be *Float32) Bytes() []byte { return (*Uint32)(be).Bytes() }
func (be *Float64) Bytes() []byte { return (*Uint64)(be).Bytes() }

func (be *Float32) Put(v float32) { (*Uint32)(be).Put(math.Float32bits(v)) }
func (be *Float64) Put(v float64) { (*Uint64)(be).Put(math.Float64bits(v)) }

func (be *Float32) Uint32() uint32 { return (*Uint32)(be).Value() }
func (be *Float64) Uint64() uint64 { return (*Uint64)(be).Value() }

func (be *Float32) Value() float32 { return math.Float32frombits(be.Uint32()) }
func (be *Float64) Value() float64 { return math.Float64frombits(be.Uint64()) }
