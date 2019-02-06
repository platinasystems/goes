// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build !linux, linux,!amd64

package iocmd

func Io_reg_wr(addr uint64, dat uint64, wid uint64) (err error) {
	panic("Unexpected Io_reg_wr")
}

func Io_reg_rd(addr uint64, wid uint64) (b []byte, err error) {
	panic("Unexpected Io_reg_rd")
}
