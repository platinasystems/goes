// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package _go_

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/platinasystems/go/internal/indent"
)

// Returns a list of machine additions
var Packages func() []map[string]string

// Return all package info
func All() []map[string]string {
	maps := []map[string]string{Package}
	if Packages != nil {
		maps = append(maps, Packages()...)
	}
	return maps
}

// Yaml formatted repos info
func Marshal() ([]byte, error) {
	buf := &bytes.Buffer{}
	ibuf := indent.New(buf, "    ")

	for _, m := range All() {
		fmt.Fprint(ibuf, m["importpath"], ":\n")
		indent.Increase(ibuf)
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
			fmt.Fprint(ibuf, k, ": ")
			if strings.Contains(s, "\n") {
				fmt.Fprint(ibuf, "|\n    ")
				indent.Increase(ibuf)
				fmt.Fprintln(ibuf, strings.TrimSpace(s))
				indent.Decrease(ibuf)
			} else {
				fmt.Fprintln(ibuf, s)
			}
		}
		indent.Decrease(ibuf)
	}
	return buf.Bytes(), nil
}
