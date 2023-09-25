// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"errors"
)

var (
	ErrIncomplete  = errors.New("incomplete")
	ErrInvalid     = errors.New("invalid")
	ErrNoPrefix    = errors.New("missing <prefix>")
	ErrNotFound    = errors.New("not found")
	ErrUnsupported = errors.New("unsupported")
	FIXME          = errors.New("FIXME")
)
