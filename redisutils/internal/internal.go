// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package internal

import (
	"fmt"
	"io"
	"unicode"
)

var these struct {
	keys  []string
	hkeys map[string][]string
}

func Fprintln(w io.Writer, s string) {
	s = Quotes(s)
	if len(s) == 0 {
		return
	}
	w.Write([]byte(s))
	if s[len(s)-1] != '\n' {
		w.Write([]byte{'\n'})
	}
}

func Quotes(s string) string {
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return fmt.Sprintf("%q", s)
		}
	}
	return s
}

/* FIXME
func Complete(line string) []string {
	args := strings.Fields(line)
	switch len(args) {
	case 1:
		return completeKeys(line)
	case 2:
		if line[len(line)-1] == ' ' {
			return completeHkeys(args[1], line)
		} else {
			return completeKeys(line)
		}
	case 3:
		return completeHkeys(args[1], line)
	}
	return []string{}
}

func CompleteKeys(line string) []string {
	if len(these.keys) == 0 {
		these.keys, _ = redis.Keys(".*")
	}
	return goes.Complete.List(line, these.keys)
}

func CompleteHkeys(key, line string) []string {
	if these.hkeys == nil {
		these.hkeys = make(map[string][]string)
	}
	l, found := these.hkeys[key]
	if !found {
		l, _ = redis.Hkeys(key)
		these.hkeys[key] = l
	}
	return goes.Complete.List(line, l)
}
*/
