// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"encoding/binary"
	"io"
)

// https://en.wikipedia.org/wiki/ICMPv6
type ICMP6 struct {
	Type uint8
	Code uint8
	Sum  uint16
}

func (p *ICMP6) ReadFrom(r io.Reader) (int64, error) {
	return SizeofICMP6, binary.Read(r, binary.BigEndian, p)
}

func (v ICMP6) WriteTo(w io.Writer) (int64, error) {
	return SizeofICMP6, binary.Write(w, binary.BigEndian, v)
}
