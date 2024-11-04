// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xerrors

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrBroken      = errors.New("broken")
	ErrIncomplete  = errors.New("incomplete")
	ErrInvalid     = errors.New("invalid")
	ErrNotFound    = errors.New("not found")
	ErrRange       = errors.New("out of range")
	ErrUnknown     = errors.New("unknown")
	ErrUnavailable = errors.New("unavailable")
	ErrUnsupported = errors.New("unsupported")
)

func Broken(args ...any) error {
	return Label(ErrBroken, args...)
}

func Incomplete(args ...any) error {
	return Label(ErrIncomplete, args...)
}

func Invalid(args ...any) error {
	return Label(ErrInvalid, args...)
}

func NotFound(args ...any) error {
	return Label(ErrNotFound, args...)
}

func Range(args ...any) error {
	return Label(ErrRange, args...)
}

func Unknown(args ...any) error {
	return Label(ErrUnknown, args...)
}

func Unavailable(args ...any) error {
	return Label(ErrUnavailable, args...)
}

func Unsupported(args ...any) error {
	return Label(ErrUnsupported, args...)
}

func IsLabelled(err error) bool {
	ne, ok := err.(NoteError)
	if ok {
		err = ne.err
	}
	_, ok = err.(LabelError)
	return ok
}

// Preface non-nil error with “arg:...”
func Label(err error, args ...any) error {
	if err == nil || len(args) == 0 {
		return err
	}
	return label(err, args...)
}

func label(err error, args ...any) LabelError {
	var sb strings.Builder
	for _, arg := range args {
		fmt.Fprint(&sb, arg, ":")
	}
	es := err.Error()
	colon := strings.IndexRune(es, ':')
	space := strings.IndexRune(es, ' ')
	if colon < 0 || (space > 0 && space < colon) {
		sb.WriteRune(' ')
	}
	sb.WriteString(es)
	return LabelError{err, sb.String()}
}

type LabelError struct {
	err error
	txt string
}

func (lbl LabelError) Error() string {
	return lbl.txt
}

func (lbl LabelError) Unwrap() error {
	return lbl.err
}
