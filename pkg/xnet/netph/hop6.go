// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"io"

	"github.com/platinasystems/goes/v2/pkg/xnet"
)

type HOP6 struct {
	Type uint8
	Len  uint8
}

func (p *HOP6) ReadFrom(r io.Reader) (int64, error) {
	xnet.BytePointer(&p.Type).ReadFrom(r)
	_, err := xnet.BytePointer(&p.Len).ReadFrom(r)
	return SizeofHOP6, err
}

func (v HOP6) WriteTo(w io.Writer) (int64, error) {
	xnet.ByteValue(v.Type).WriteTo(w)
	_, err := xnet.ByteValue(v.Len).WriteTo(w)
	return SizeofHOP6, err
}
