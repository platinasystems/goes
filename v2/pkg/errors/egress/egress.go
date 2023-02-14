// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package egress

import (
	"fmt"
	"path/filepath"
	"runtime"
)

var FilenameMutation = filepath.Base

// After calling all funcs, this returns any non-nil error wrapped with the
// caller's file name and line number, e.g.
//
//	if err != nil {
//		return Marked(err, f.Close)
//	}
func Marked(err error, funcs ...func() error) error {
	if err = Unmarked(err, funcs...); err != nil {
		if _, f, l, ok := runtime.Caller(1); ok {
			err = &mark{f, l, err}
		}
	}
	return err
}

// After calling all funcs, this returns any non-nil error without with
// caller's file name and line number wapper, e.g.
//
//	if err != nil {
//		panic(Unmarked(err, f.Close)
//	}
func Unmarked(err error, funcs ...func() error) error {
	for _, f := range funcs {
		if t := f(); err == nil {
			err = t
		}
	}
	return err
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
	return fmt.Sprint(FilenameMutation(m.file), ":", m.line, sep, m.err)
}

func (m mark) Unwrap() error {
	return m.err
}
