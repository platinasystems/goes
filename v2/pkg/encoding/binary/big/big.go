// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package big

import (
	"encoding/binary"
	"math"
)

var Endian = binary.BigEndian

type Uint8 []byte
type Uint16 []byte
type Uint32 []byte
type Uint64 []byte

type Float32 []byte
type Float64 []byte

func (be Uint8) Put(v uint8)   { be[0] = byte(v) }
func (be Uint16) Put(v uint16) { Endian.PutUint16([]byte(be), v) }
func (be Uint32) Put(v uint32) { Endian.PutUint32([]byte(be), v) }
func (be Uint64) Put(v uint64) { Endian.PutUint64([]byte(be), v) }

func (be Uint8) Value() uint8   { return uint8(be[0]) }
func (be Uint16) Value() uint16 { return Endian.Uint16([]byte(be)) }
func (be Uint32) Value() uint32 { return Endian.Uint32([]byte(be)) }
func (be Uint64) Value() uint64 { return Endian.Uint64([]byte(be)) }

func (be Float32) Put(v float32) { Uint32(be).Put(math.Float32bits(v)) }
func (be Float64) Put(v float64) { Uint64(be).Put(math.Float64bits(v)) }

func (be Float32) Value() float32 {
	return math.Float32frombits(Uint32(be).Value())
}

func (be Float64) Value() float64 {
	return math.Float64frombits(Uint64(be).Value())
}
