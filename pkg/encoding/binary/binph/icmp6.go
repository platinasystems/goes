// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"io"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/ICMPv6
type ICMP6 struct {
	Type uint8
	Code uint8
	Sum  uint16
}

const SizeofICMP6 = int64(unsafe.Sizeof(ICMP6{}))

func (p *ICMP6) ReadFrom(r io.Reader) (int64, error) {
	binint.BytePointer(&p.Type).ReadFrom(r)
	binint.BytePointer(&p.Code).ReadFrom(r)
	_, err := binint.BigEndianPointer(&p.Sum).ReadFrom(r)
	return SizeofICMP6, err
}

func (v ICMP6) WriteTo(w io.Writer) (int64, error) {
	binint.ByteValue(v.Type).WriteTo(w)
	binint.ByteValue(v.Code).WriteTo(w)
	_, err := binint.BigEndianValue(v.Sum).WriteTo(w)
	return SizeofICMP6, err
}
