// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides encoding in which each data value is preceded by a
// 2-byte, big-endian unsigned integer length that doesn't exceed Max bytes.
package lv

import "errors"

const Max = 4096

var (
	ErrTooLarge = errors.New("encoded length is too large")
	ErrTooSmall = errors.New("receiving buf is too small")
)
