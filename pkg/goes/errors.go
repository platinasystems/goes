// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"errors"
	"strings"
)

var (
	FIXME         = errors.New("FIXME")
	ErrEmpty      = errors.New("empty selection")
	ErrIncomplete = errors.New("incomplete")
	ErrNoRoom     = errors.New("no room")
	ErrNotFound   = errors.New("not found")
	ErrRange      = errors.New("out of range")
	ErrUnexpected = errors.New("unexpected")
)

type GoesError struct {
	Path []string
	Err  error
}

func IsGoesError(err error) bool {
	_, ok := err.(GoesError)
	return ok
}

func (e GoesError) Error() string {
	s := strings.Join(e.Path, ":") + ":"
	eee := e.Err.Error()
	if strings.IndexRune(eee, ':') < 0 {
		s += " "
	}
	s += eee
	return s
}

func (e GoesError) Unwrap() error { return e.Err }
