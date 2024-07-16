// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package xnet

//go:generate textembedder iff_*.txt

import (
	"strings"

	"github.com/platinasystems/goes/v2/pkg/integer"
)

func IFFStrings() []string {
	return strings.Split(IFFText, "\n")
}

func IFFNames[T integer.Int | integer.Uint](iff T) string {
	return integer.Names(iff, IFFStrings())
}
