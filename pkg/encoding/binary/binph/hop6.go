// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"io"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

type HOP6 struct {
	Type uint8
	Len  uint8
}

const SizeofHOP6 = int64(unsafe.Sizeof(HOP6{}))

func (p *HOP6) ReadFrom(r io.Reader) (int64, error) {
	binint.BytePointer(&p.Type).ReadFrom(r)
	_, err := binint.BytePointer(&p.Len).ReadFrom(r)
	return SizeofHOP6, err
}

func (v HOP6) WriteTo(w io.Writer) (int64, error) {
	binint.ByteValue(v.Type).WriteTo(w)
	_, err := binint.ByteValue(v.Len).WriteTo(w)
	return SizeofHOP6, err
}
