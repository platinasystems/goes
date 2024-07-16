// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package xerrors

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

var ErrFIXME = errors.New("FIXME")
var FileNameMutation = filepath.Base

func IsMarked(err error) bool {
	_, ok := err.(*MarkError)
	return ok
}

// Mark ErrFIXME with the caller's file name and line number and args.
func FIXME(args ...any) error {
	return MarkCaller(ErrFIXME, args...)
}

// Wrap a non-nil error with the caller's file name and line number and args.
func Mark(err error, args ...any) error {
	if err != nil {
		if _, f, l, ok := runtime.Caller(1); ok {
			err = &MarkError{f, l, err, args}
		} else {
			err = &MarkError{"unavailable", -1, err, args}
		}
	}
	return err
}

// Skip 2 calls back.
func MarkCaller(err error, args ...any) error {
	if err != nil {
		if _, f, l, ok := runtime.Caller(2); ok {
			err = &MarkError{f, l, err, args}
		} else {
			err = &MarkError{"unavailable", -1, err, args}
		}
	}
	return err
}

// Wrap fmt.Errorf with the caller's file name and line number.
func Markf(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	if _, f, l, ok := runtime.Caller(1); ok {
		err = &MarkError{f, l, err, nil}
	}
	return err
}

func MarkResult[T any](v T, err error) (T, error) {
	return v, MarkCaller(err)
}

type MarkError struct {
	file string
	line int
	err  error
	args []any
}

func (m *MarkError) Error() string {
	if m == nil {
		return ""
	}
	var sb strings.Builder
	if m.line >= 0 {
		fmt.Fprint(&sb, FileNameMutation(m.file), ":", m.line)
		if IsMarked(m.err) {
			fmt.Fprint(&sb, "/ ")
		} else {
			fmt.Fprint(&sb, ": ")
		}
	}
	fmt.Fprint(&sb, m.err)
	if len(m.args) > 0 {
		fmt.Fprint(&sb, " [")
		for i, arg := range m.args {
			if i > 0 {
				fmt.Fprint(&sb, " ")
			}
			fmt.Fprint(&sb, arg)
		}
		fmt.Fprint(&sb, "]")
	}
	return sb.String()
}

func (m *MarkError) Unwrap() error {
	return m.err
}
