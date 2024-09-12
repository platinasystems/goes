// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package xerrors

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
)

var ErrFIXME = errors.New("FIXME")
var FileNameMutation = filepath.Base

// Mark ErrFIXME with the caller's file name and line number and args.
func FIXME(args ...any) error {
	return MarkCaller(ErrFIXME, args...)
}

// [Note] the non-nil error then [Label] that with the caller's file name and
// line number.
func Mark(err error, args ...any) error {
	if err == nil {
		return err
	}
	if len(args) > 0 {
		err = note(err, args...)
	}
	if _, f, l, ok := runtime.Caller(1); ok {
		err = MarkError{label(err, FileNameMutation(f), l)}
	}
	return err
}

// Skip 2 calls back.
func MarkCaller(err error, args ...any) error {
	if err == nil {
		return err
	}
	if len(args) > 0 {
		err = note(err, args...)
	}
	if _, f, l, ok := runtime.Caller(2); ok {
		err = MarkError{label(err, FileNameMutation(f), l)}
	}
	return err
}

// Preface fmt.Errorf with the caller's file name and line number.
func Markf(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	if _, f, l, ok := runtime.Caller(1); ok {
		err = MarkError{label(err, FileNameMutation(f), l)}
	}
	return err
}

func MarkResult[T any](v T, err error) (T, error) {
	return v, MarkCaller(err)
}

type MarkError struct{ LabelError }

func IsMarked(err error) bool {
	ne, ok := err.(NoteError)
	if ok {
		err = ne.err
	}
	_, ok = err.(MarkError)
	return ok
}
