// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xerrors

import (
	"fmt"
	"strings"
)

// Suffix err with " [arg  ...]"
func Note(err error, args ...any) error {
	if err == nil || len(args) == 0 {
		return err
	}
	return note(err, args...)
}

func note(err error, args ...any) NoteError {
	var sb strings.Builder
	sb.WriteString(err.Error())
	sep := " ["
	for _, arg := range args {
		fmt.Fprint(&sb, sep, arg)
		sep = " "
	}
	sb.WriteRune(']')
	return NoteError{err, sb.String()}
}

type NoteError struct {
	err error
	txt string
}

func (e NoteError) Error() string {
	return e.txt
}

func (e NoteError) Unwrap() error {
	return e.err
}
