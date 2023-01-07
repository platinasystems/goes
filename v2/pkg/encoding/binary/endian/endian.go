// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package endian

type Numeric interface {
	uint8 | uint16 | uint32 | uint64 | float32 | float64
}

type Number[T Numeric] interface {
	Put(T)
	Value() T
}
