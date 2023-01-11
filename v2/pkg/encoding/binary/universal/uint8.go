// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package universal

type Uint8 [1]byte

func NewUint8(b []byte) (*Uint8, []byte) { return (*Uint8)(b), b[1:] }

func (h *Uint8) Put(v uint8)  { h[0] = byte(v) }
func (h *Uint8) Value() uint8 { return uint8(h[0]) }
