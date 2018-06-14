// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vfat

import (
	"bytes"
)

func checkMagic(s []byte, o int, magics ...[]byte) bool {
	for _, m := range magics {
		chk := s[o : o+len(m)]
		if bytes.Equal(chk, m) {
			return true
		}
	}
	return false
}

func Probe(s []byte) bool {
	// Too many magic numbers
	if checkMagic(s, 0,
		[]byte{0xe9},
		[]byte{0xeb}) {
		return true
	}
	if checkMagic(s, 0x36,
		[]byte("MSDOS"),
		[]byte("FAT     "),
		[]byte("FAT12   "),
		[]byte("FAT16   "),
		[]byte("MSDOS   ")) {
		return true
	}
	if checkMagic(s, 0x52,
		[]byte("MSWIN"),
		[]byte("FAT32   ")) {
		return true
	}
	return false
}
