// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package iso9660

const (
	MagicOff1 = 0x8001
	MagicOff2 = 0x8801
	MagicOff3 = 0x9001

	MagicVal1 = 'C'
	MagicVal2 = 'D'
	MagicVal3 = '0'
	MagicVal4 = '0'
	MagicVal5 = '1'
)

func isValidSignature(s []byte, off int) bool {
	return s[off] == MagicVal1 &&
		s[off+1] == MagicVal2 &&
		s[off+2] == MagicVal3 &&
		s[off+3] == MagicVal4 &&
		s[off+4] == MagicVal5
}

func Probe(s []byte) bool {
	return isValidSignature(s, MagicOff1) ||
		isValidSignature(s, MagicOff2) ||
		isValidSignature(s, MagicOff3)
}
