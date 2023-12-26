// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package iff

//go:generate textembedder iff_*.txt

import (
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/integer"
)

var Lines = sync.OnceValue(func() []string {
	return strings.Split(Text, "\n")
})

func Names[T integer.Int | integer.Uint](iff T) string {
	return integer.Names(iff, Lines())
}
