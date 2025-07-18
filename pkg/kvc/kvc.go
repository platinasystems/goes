// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package “kvc” provides a key/value configuration parser.
package kvc

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// Call method with line number, key, and space separated value(s) from
// the hash-tag (#) comment stripped, space trimmed, non-blank lines.
func Range(r io.Reader, method func(int, string, []string) error) error {
	sc := bufio.NewScanner(r)
	for lno := 1; sc.Scan(); lno++ {
		s := strings.TrimSpace(sc.Text())
		if len(s) == 0 || strings.HasPrefix(s, "#") {
			continue
		}
		if i := strings.Index(s, " #"); i > 0 {
			s = s[:i]
		} else if i = strings.Index(s, "\t#"); i > 0 {
			s = s[:i]
		}
		fields := strings.Fields(strings.TrimSpace(s))
		if len(fields) == 0 {
			continue
		}
		if err := method(lno, fields[0], fields[1:]); err != nil {
			return err
		}
	}
	return nil
}

func RangeFile(name string, method func(int, string, []string) error) error {
	f, err := os.Open(name)
	if err == nil {
		defer f.Close()
		err = Range(f, method)
	}
	return err
}
