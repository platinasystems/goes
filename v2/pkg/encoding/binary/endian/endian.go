// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package endian

type Number interface {
	uint8 | uint16 | uint32 | uint64 | float32 | float64
}

type Endian[T Number] interface {
	Put(T)
	Value() T
}

type Uint8 = Endian[uint8]
type Uint16 = Endian[uint16]
type Uint32 = Endian[uint32]
type Uint64 = Endian[uint64]
type Float32 = Endian[float32]
type Float64 = Endian[float64]
