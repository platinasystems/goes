// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lv

const (
	Ebit    = 15
	Eflag   = 1 << Ebit
	Efilter = Eflag - 1
)

type Nack string

func (s Nack) Error() string { return string(s) }

func nack(b []byte) Nack {
	c := make([]byte, len(b))
	copy(c, b)
	return Nack(c)
}
