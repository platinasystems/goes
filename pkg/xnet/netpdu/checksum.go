// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

type Checksummer interface {
	Checksum(uint8, []byte) uint16
}

// http://www.faqs.org/rfcs/rfc1071.html
func Checksum(data []byte) (v uint32) {
	for len(data) > 1 {
		v += uint32(data[0])<<8 | uint32(data[1])
		data = data[2:]
	}
	if len(data) == 1 {
		v += uint32(data[0]) << 8
	}
	return
}

func CarryOver(v uint32) uint16 {
	for carry := v >> 16; carry != 0; carry = v >> 16 {
		v = (v & 0xffff) + carry
	}
	return uint16(v)
}
