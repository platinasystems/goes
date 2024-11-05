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
	ErrOverrun     = errors.New("overrun")
	ErrRange       = errors.New("out of range")
	ErrUnderrun    = errors.New("underrun")
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

func IsLabelled(err error) bool {
	ne, ok := err.(NoteError)
	if ok {
		err = ne.err
	}
	_, ok = err.(LabelError)
	return ok
}

func IsBroken(err error) bool      { return errors.Is(err, ErrBroken) }
func IsIncomplete(err error) bool  { return errors.Is(err, ErrIncomplete) }
func IsInvalid(err error) bool     { return errors.Is(err, ErrInvalid) }
func IsNotFound(err error) bool    { return errors.Is(err, ErrNotFound) }
func IsOverrun(err error) bool     { return errors.Is(err, ErrOverrun) }
func IsRange(err error) bool       { return errors.Is(err, ErrRange) }
func IsUnderrun(err error) bool    { return errors.Is(err, ErrUnderrun) }
func IsUnknown(err error) bool     { return errors.Is(err, ErrUnknown) }
func IsUnavailable(err error) bool { return errors.Is(err, ErrUnavailable) }
func IsUnsupported(err error) bool { return errors.Is(err, ErrUnsupported) }

func NotFound(args ...any) error {
	return Label(ErrNotFound, args...)
}

func Overrun(args ...any) error {
	return Label(ErrOverrun, args...)
}

func Range(args ...any) error {
	return Label(ErrRange, args...)
}

func Underrun(args ...any) error {
	return Label(ErrUnderrun, args...)
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
