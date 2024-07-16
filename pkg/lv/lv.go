// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides encoding in which each data value is preceded by a
// 2-byte, big-endian unsigned integer length that doesn't exceed Max bytes.
package lv

const (
	Ebit    = 15
	Eflag   = 1 << Ebit
	Efilter = Eflag - 1
	Max     = 4096
)
