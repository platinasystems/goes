// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"io"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

type TunPI struct {
	Flags,
	Proto uint16
}

const SizeofTunPI = int64(unsafe.Sizeof(TunPI{}))

func (p *TunPI) ReadFrom(r io.Reader) (int64, error) {
	binint.BigEndianPointer(&p.Flags).ReadFrom(r)
	_, err := binint.BigEndianPointer(&p.Proto).ReadFrom(r)
	return SizeofTunPI, err
}

func (v TunPI) WriteTo(w io.Writer) (int64, error) {
	binint.BigEndianValue(v.Flags).WriteTo(w)
	_, err := binint.BigEndianValue(v.Proto).WriteTo(w)
	return SizeofTunPI, err
}
