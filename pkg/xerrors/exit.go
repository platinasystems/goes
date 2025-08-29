// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xerrors

type ExitError struct {
	code int
	err  error
}

func NewExitError(code int, err error) ExitError {
	return ExitError{code, err}
}

func (x ExitError) Error() (s string) {
	if x.err != nil {
		s = x.err.Error()
	}
	return
}

func (x ExitError) ExitCode() int {
	return x.code
}

func (x ExitError) Unwrap() error {
	return x.err
}
