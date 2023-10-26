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
	ErrIncomplete = errors.New("incomplete")
	ErrNoRoom     = errors.New("no room")
	ErrNotFound   = errors.New("not found")
	ErrRange      = errors.New("out of range")
	ErrUnexpected = errors.New("unexpected")
)

type Error struct {
	Path []string
	Err  error
}

func (e Error) Error() string {
	s := strings.Join(e.Path, "/") + ":"
	eee := e.Err.Error()
	if strings.IndexRune(eee, ':') < 0 {
		s += " "
	}
	s += eee
	return s
}

func (e Error) Unwrap() error { return e.Err }
