// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"strings"

	"github.com/platinasystems/goes/v2/pkg/integer"
)

//go:generate textembedder ift_*.txt

func IFTStrings() []string {
	return strings.Split(IFTText, "\n")
}

func IFTName[T integer.Int | integer.Uint](ift T) string {
	return integer.Name(ift, IFTStrings())
}
