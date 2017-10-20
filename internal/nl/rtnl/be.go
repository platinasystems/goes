// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

type Be16 [2]byte
type Be32 [4]byte
type Be64 [8]byte

func (be *Be16) Load() uint16 {
	v := uint16(be[0] << 8)
	v |= uint16(be[1])
	return v
}

func (be *Be16) Store(v uint16) {
	be[0] = byte(v >> 8)
	be[1] = byte(v)
}

func (be *Be32) Load() uint32 {
	v := uint32(be[0] << 24)
	v |= uint32(be[1] << 16)
	v |= uint32(be[2] << 8)
	v |= uint32(be[3])
	return v
}

func (be *Be32) Store(v uint32) {
	be[0] = byte(v >> 24)
	be[1] = byte(v >> 16)
	be[2] = byte(v >> 8)
	be[3] = byte(v)
}

func (be *Be64) Load() uint64 {
	v := uint64(be[0] << 56)
	v |= uint64(be[1] << 48)
	v |= uint64(be[2] << 40)
	v |= uint64(be[3] << 32)
	v |= uint64(be[4] << 24)
	v |= uint64(be[5] << 16)
	v |= uint64(be[6] << 8)
	v |= uint64(be[7])
	return v
}

func (be *Be64) Store(v uint64) {
	be[0] = byte(v >> 56)
	be[1] = byte(v >> 48)
	be[2] = byte(v >> 40)
	be[3] = byte(v >> 32)
	be[4] = byte(v >> 24)
	be[5] = byte(v >> 16)
	be[6] = byte(v >> 8)
	be[7] = byte(v)
}
