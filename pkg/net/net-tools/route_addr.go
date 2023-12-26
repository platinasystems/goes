// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import "unicode"

func routeAddrIsNumeric(s string) bool {
	r := []rune(s)[0]
	return unicode.IsNumber(r) || r == ':'
}
