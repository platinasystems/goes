// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ift

import (
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/integer"
)

//go:generate textembedder ift_*.txt

var Lines = sync.OnceValue(func() []string {
	return strings.Split(Text, "\n")
})

func Name[T integer.Int | integer.Uint](ift T) string {
	return integer.Name(ift, Lines())
}
