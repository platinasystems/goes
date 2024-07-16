// Copyright © 2015-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import "os"

const (
	Sizeof32 = 4
	Sizeof64 = 8
)

type Align int

var (
	Align32   = Align(Sizeof32).Roundup
	Align64   = Align(Sizeof64).Roundup
	PageAlign = Align(os.Getpagesize()).Roundup
)

func (a Align) Roundup(i int) int {
	return Roundup(i, a)
}

func Roundup[A ~int](i int, a A) int {
	return i + (int(a)-i%int(a))%int(a)
}
