// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"io"

	"github.com/platinasystems/goes/v2/pkg/xnet"
)

// https://en.wikipedia.org/wiki/User_Datagram_Protocol
type UDP struct {
	SP  uint16
	DP  uint16
	Len uint16
	Sum uint16
}

func (p *UDP) ReadFrom(r io.Reader) (int64, error) {
	xnet.BigEndianPointer(&p.SP).ReadFrom(r)
	xnet.BigEndianPointer(&p.DP).ReadFrom(r)
	xnet.BigEndianPointer(&p.Len).ReadFrom(r)
	_, err := xnet.BigEndianPointer(&p.Sum).ReadFrom(r)
	return SizeofUDP, err
}

func (v UDP) WriteTo(w io.Writer) (int64, error) {
	xnet.BigEndianValue(v.SP).WriteTo(w)
	xnet.BigEndianValue(v.DP).WriteTo(w)
	xnet.BigEndianValue(v.Len).WriteTo(w)
	_, err := xnet.BigEndianValue(v.Sum).WriteTo(w)
	return SizeofUDP, err
}
