// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xerrors

import (
	"errors"
	"runtime"
)

var ErrFalse = errors.New("false")

func assert(err error) {
	if err != nil {
		if _, f, l, ok := runtime.Caller(2); ok {
			err = MarkError{label(err, FileNameMutation(f), l)}
		}
		panic(err)
	}
}

// Panic w/ file name:line of caller if err != nil
func Assert(err error) {
	assert(err)
}

func AssertResult[T any](v T, err error) T {
	assert(err)
	return v
}

func AssertResultPair[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	assert(err)
	return v1, v2
}

func AssertTrue(t bool) {
	if !t {
		err := ErrFalse
		if _, f, l, ok := runtime.Caller(1); ok {
			err = MarkError{label(err, FileNameMutation(f), l)}
		}
		panic(err)
	}
}
