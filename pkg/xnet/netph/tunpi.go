// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"encoding/binary"
	"io"
)

type TunPI struct {
	Flags,
	Proto uint16
}

func (p *TunPI) ReadFrom(r io.Reader) (int64, error) {
	return SizeofTunPI, binary.Read(r, binary.BigEndian, p)
}

func (v TunPI) WriteTo(w io.Writer) (int64, error) {
	return SizeofTunPI, binary.Write(w, binary.BigEndian, v)
}
