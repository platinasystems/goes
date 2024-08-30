// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"encoding/binary"
	"io"
)

// https://en.wikipedia.org/wiki/User_Datagram_Protocol
type UDP struct {
	SP  uint16
	DP  uint16
	Len uint16
	Sum uint16
}

func (p *UDP) ReadFrom(r io.Reader) (int64, error) {
	return SizeofUDP, binary.Read(r, binary.BigEndian, p)
}

func (v UDP) WriteTo(w io.Writer) (int64, error) {
	return SizeofUDP, binary.Write(w, binary.BigEndian, v)
}
