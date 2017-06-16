// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package _go_

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/platinasystems/go/internal/accumulate"
	"github.com/platinasystems/go/internal/indent"
)

// Packages returns a list of repos info
var Packages func() []map[string]string

// Write Yaml formatted repos info
func WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(indent.New(w, "    "))
	defer acc.Fini()

	maps := []map[string]string{Package}
	if Packages != nil {
		maps = append(maps, Packages()...)
	}
	for _, m := range maps {
		fmt.Fprint(acc, m["importpath"], ":\n")
		indent.Increase(acc)
		keys := make([]string, 0, len(m))
		for k := range m {
			if k != "importpath" {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		for _, k := range keys {
			s := m[k]
			if len(s) == 0 {
				continue
			}
			fmt.Fprint(acc, k, ": ")
			if strings.Contains(s, "\n") {
				fmt.Fprint(acc, "|\n    ")
				indent.Increase(acc)
				fmt.Fprintln(acc, strings.TrimSpace(s))
				indent.Decrease(acc)
			} else {
				fmt.Fprintln(acc, s)
			}
		}
		indent.Decrease(acc)
	}
	return acc.Tuple()
}
