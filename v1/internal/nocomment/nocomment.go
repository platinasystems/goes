// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package nocomment strips a string trailing '#' prefaced comments along with
// its leading whitespace. For example,
//
//	nocomment.New("hello # world") returns "hello"
//	nocomment.New("# hello world") returns ""
//	nocomment.New("hello#world") returns "hello#world"
//	nocomment.New("hello #world") returns "hello"
package nocomment

import "strings"

func New(s string) string {
	var t string
	if len(s) == 0 {
		return t
	}
	if s[0] == '#' {
		return t
	}
	t = strings.TrimLeft(s, " \t")
	if len(t) == 0 {
		return t
	}
	if t[0] == '#' {
		return t
	}
	if i := strings.IndexRune(t, '#'); i > 0 {
		if t[i-1] == ' ' || t[i-1] == '\t' {
			return strings.TrimRight(t[:i], " \t")
		}
	}
	return t
}
