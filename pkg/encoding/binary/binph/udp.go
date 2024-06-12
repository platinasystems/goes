// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"io"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/User_Datagram_Protocol
type UDP struct {
	SP  uint16
	DP  uint16
	Len uint16
	Sum uint16
}

const SizeofUDP = int64(unsafe.Sizeof(UDP{}))

func (p *UDP) ReadFrom(r io.Reader) (int64, error) {
	binint.BigEndianPointer(&p.SP).ReadFrom(r)
	binint.BigEndianPointer(&p.DP).ReadFrom(r)
	binint.BigEndianPointer(&p.Len).ReadFrom(r)
	_, err := binint.BigEndianPointer(&p.Sum).ReadFrom(r)
	return SizeofUDP, err
}

func (v UDP) WriteTo(w io.Writer) (int64, error) {
	binint.BigEndianValue(v.SP).WriteTo(w)
	binint.BigEndianValue(v.DP).WriteTo(w)
	binint.BigEndianValue(v.Len).WriteTo(w)
	_, err := binint.BigEndianValue(v.Sum).WriteTo(w)
	return SizeofUDP, err
}
