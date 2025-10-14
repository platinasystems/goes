// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package “kvc” provides a key/value configuration parser.
package kvc

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
	"unicode"
)

type ProcKeyVals = func(int, string, []string) error
type Split = func(string) []string

var DefaultSplit = func(s string) []string {
	s = strings.TrimSpace(s)
	if len(s) == 0 || !unicode.IsLetter([]rune(s)[0]) {
		return nil
	}
	return strings.Fields(s)
}

// Range calls “split” or, if that's nil, [DefaultSplit]
// to get fields from each line scanned from reader.
// It then processes the sliced keyword and remaining value(s).
// This skips lines with nil or empty fields and
// stops at the first error.
func Range(r io.Reader, split Split, proc ProcKeyVals) error {
	if proc == nil {
		return nil
	}
	if split == nil {
		split = DefaultSplit
	}
	sc := bufio.NewScanner(r)
	for lno := 1; sc.Scan(); lno++ {
		if args := split(sc.Text()); len(args) > 0 {
			if err := proc(lno, args[0], args[1:]); err != nil {
				return err
			}
		}
	}
	return nil
}

// [Range] named file.
func RangeFile(name string, split Split, proc ProcKeyVals) error {
	f, err := os.Open(name)
	if err == nil {
		err = Range(f, split, proc)
		f.Close()
	}
	return err
}

func RangeString(s string, split Split, proc ProcKeyVals) error {
	return Range(strings.NewReader(s), split, proc)
}

func RangeText(data []byte, split Split, proc ProcKeyVals) error {
	return Range(bytes.NewReader(data), split, proc)
}
