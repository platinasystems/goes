// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"io"

	"github.com/platinasystems/goes/v2/pkg/xnet"
)

type TunPI struct {
	Flags,
	Proto uint16
}

func (p *TunPI) ReadFrom(r io.Reader) (int64, error) {
	xnet.BigEndianPointer(&p.Flags).ReadFrom(r)
	_, err := xnet.BigEndianPointer(&p.Proto).ReadFrom(r)
	return SizeofTunPI, err
}

func (v TunPI) WriteTo(w io.Writer) (int64, error) {
	xnet.BigEndianValue(v.Flags).WriteTo(w)
	_, err := xnet.BigEndianValue(v.Proto).WriteTo(w)
	return SizeofTunPI, err
}
