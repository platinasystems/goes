// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package egress

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var FileNameMutation = filepath.Base

func IsMarked(err error) bool {
	_, ok := err.(*mark)
	return ok
}

// Wrap a non-nil error with the caller's file name and line number, e.g.
//
//	if err != nil {
//		return Marked(err)
//	}
func Mark(err error) error {
	if err != nil {
		if _, f, l, ok := runtime.Caller(1); ok {
			err = &mark{f, l, err}
		}
	}
	return err
}

// Skip 2 calls back.
func MarkCaller(err error) error {
	if err != nil {
		if _, f, l, ok := runtime.Caller(2); ok {
			err = &mark{f, l, err}
		}
	}
	return err
}

// Wrap fmt.Errorf with the caller's file name and line number, e.g.
//
//	return Markf(format, args...)
func Markf(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	if _, f, l, ok := runtime.Caller(1); ok {
		err = &mark{f, l, err}
	}
	return err
}

func MarkResult[T any](v T, err error) (T, error) {
	return v, MarkCaller(err)
}

type mark struct {
	file string
	line int
	err  error
}

func (m mark) Error() string {
	sep := ": "
	if _, ok := m.err.(*mark); ok {
		sep = "/"
	}
	return fmt.Sprint(FileNameMutation(m.file), ":", m.line, sep, m.err)
}

func (m mark) Unwrap() error {
	return m.err
}
