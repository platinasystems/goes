// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package egress

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
)

var FileNameMutation = filepath.Base

// Wrap a non-nil error with the caller's file name and line number, e.g.
//
//	if err != nil {
//		return Marked(err)
//	}
func Marked(err error) error {
	if err != nil {
		if _, f, l, ok := runtime.Caller(1); ok {
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

// Record a recovered panic as an error wrapped with the panic'd file name and
// line number, e.g.
//
//	func() (err error) {
//		...
//		defer Recovery(&err)
//		}
//		...
//		panic("oops")
//	}
func Recovery(p *error) {
	r := recover()
	if r == nil {
		return
	}
	err, is_error := r.(error)
	if !is_error {
		err = errors.New(fmt.Sprint(r))
	}
	depth := 2
	if _, is_runtime := err.(runtime.Error); is_runtime {
		depth = 3
	}
	if _, f, l, ok := runtime.Caller(depth); ok {
		*p = mark{f, l, err}
	} else {
		*p = err
	}
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
