// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xutf8

import . "unicode/utf8"

// Assures that non-empty slice has one and only one trailing newline.
func AlineBytes(s []byte) []byte {
	if len(s) == 0 || !Valid(s) {
		return s
	}
	r, size := DecodeLastRune(s)
	if r != '\n' {
		return AppendRune(s, '\n')
	}
	for {
		i := len(s) - size
		r, size = DecodeLastRune(s[:i])
		if r != '\n' {
			break
		}
		s = s[:i]
	}
	return s
}

// Assures that non-empty slice has one and only one trailing newline.
func AlineRunes(s []rune) []rune {
	if len(s) == 0 {
		return s
	}
	if s[len(s)-1] != '\n' {
		return append(s, '\n')
	}
	for len(s) > 1 {
		if s[len(s)-2] != '\n' {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}

// Assures that non-empty string has one and only one trailing newline.
func AlineString(s string) string {
	if len(s) == 0 || !ValidString(s) {
		return s
	}
	r, size := DecodeLastRuneInString(s)
	if r != '\n' {
		return s + "\n"
	}
	for {
		i := len(s) - size
		r, size = DecodeLastRuneInString(s[:i])
		if r != '\n' {
			break
		}
		s = s[:i]
	}
	return s
}
