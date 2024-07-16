// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"io"

	"github.com/platinasystems/goes/v2/pkg/xnet"
)

// https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol
type ICMP struct {
	Type uint8
	Code uint8
	Sum  uint16
}

func (p *ICMP) ReadFrom(r io.Reader) (int64, error) {
	xnet.BytePointer(&p.Type).ReadFrom(r)
	xnet.BytePointer(&p.Code).ReadFrom(r)
	_, err := xnet.BigEndianPointer(&p.Sum).ReadFrom(r)
	return SizeofICMP, err
}

func (v ICMP) WriteTo(w io.Writer) (int64, error) {
	xnet.ByteValue(v.Type).WriteTo(w)
	xnet.ByteValue(v.Code).WriteTo(w)
	_, err := xnet.BigEndianValue(v.Sum).WriteTo(w)
	return SizeofICMP, err
}
