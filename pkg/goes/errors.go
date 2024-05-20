// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import "errors"

var (
	FIXME          = errors.New("FIXME")
	ErrDisabled    = errors.New("disabled")
	ErrEmpty       = errors.New("empty selection")
	ErrIncomplete  = errors.New("incomplete")
	ErrNoRoom      = errors.New("no room")
	ErrNotFound    = errors.New("not found")
	ErrRange       = errors.New("out of range")
	ErrUnavailable = errors.New("unavailable")
	ErrUnexpected  = errors.New("unexpected")
)
