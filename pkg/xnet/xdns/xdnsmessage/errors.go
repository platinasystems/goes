// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import "errors"

var (
	ErrIncomplete  = errors.New("incomplete")
	ErrInvalid     = errors.New("invalid")
	ErrOverrun     = errors.New("overrun")
	ErrUnderrun    = errors.New("underrun")
	ErrUnsupported = errors.New("unsupported")

	FIXME = errors.New("FIXME")
)
