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
)

func Broken(txt ...string) error {
	return Label(ErrBroken, txt...)
}

func Incomplete(txt ...string) error {
	return Label(ErrIncomplete, txt...)
}

func Invalid(txt ...string) error {
	return Label(ErrInvalid, txt...)
}

func NotFound(txt ...string) error {
	return Label(ErrNotFound, txt...)
}

func Range(txt ...string) error {
	return Label(ErrRange, txt...)
}

func Unknown(txt ...string) error {
	return Label(ErrUnknown, txt...)
}

func Unavailable(txt ...string) error {
	return Label(ErrUnavailable, txt...)
}

type LabelError struct {
	err error
	txt []string
}

func IsLabelled(err error) bool {
	_, ok := err.(*LabelError)
	return ok
}

func Label(err error, txt ...string) error {
	if err == nil || IsMarked(err) {
		return err
	}
	if len(txt) > 0 {
		err = &LabelError{err, txt}
	}
	return err
}

func (lbl *LabelError) Error() string {
	var sb strings.Builder
	fmt.Fprint(&sb, strings.Join(lbl.txt, ":"))
	sb.WriteRune(':')
	es := lbl.err.Error()
	colon := strings.IndexRune(es, ':')
	space := strings.IndexRune(es, ' ')
	if colon < 0 || (space > 0 && space < colon) {
		sb.WriteRune(' ')
	}
	sb.WriteString(es)
	return sb.String()
}

func (lbl *LabelError) Unwrap() error {
	return lbl.err
}
