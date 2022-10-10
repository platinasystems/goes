// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lv

import "strings"

var (
	ErrTooLarge = Nack("encoded length is too large")
	ErrTooSmall = Nack("receiving buf is too small")
)

type Nack string

func NewNack(s string) Nack { return Nack(strings.Clone(s)) }

func (s Nack) Error() string { return string(s) }
