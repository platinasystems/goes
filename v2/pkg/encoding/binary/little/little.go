// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package little

import (
	"encoding/binary"
	"math"
)

var Endian = binary.LittleEndian

type Uint16 [2]byte
type Uint32 [4]byte
type Uint64 [8]byte

func (le *Uint16) Put(v uint16) { Endian.PutUint16(le[:], v) }
func (le *Uint32) Put(v uint32) { Endian.PutUint32(le[:], v) }
func (le *Uint64) Put(v uint64) { Endian.PutUint64(le[:], v) }

func (le *Uint16) Value() uint16 { return Endian.Uint16(le[:]) }
func (le *Uint32) Value() uint32 { return Endian.Uint32(le[:]) }
func (le *Uint64) Value() uint64 { return Endian.Uint64(le[:]) }

type Float32 Uint32
type Float64 Uint64

func (le *Float32) Put(v float32) { (*Uint32)(le).Put(math.Float32bits(v)) }
func (le *Float64) Put(v float64) { (*Uint64)(le).Put(math.Float64bits(v)) }

func (le *Float32) Uint32() uint32 { return (*Uint32)(le).Value() }
func (le *Float64) Uint64() uint64 { return (*Uint64)(le).Value() }

func (le *Float32) Value() float32 { return math.Float32frombits(le.Uint32()) }
func (le *Float64) Value() float64 { return math.Float64frombits(le.Uint64()) }
