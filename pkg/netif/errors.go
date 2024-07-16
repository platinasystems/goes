// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"errors"
	"io/fs"
)

var (
	ErrIncomplete  = errors.New("incomplete")
	ErrInvalid     = fs.ErrInvalid
	ErrNoPrefix    = errors.New("missing <prefix>")
	ErrNotFound    = errors.New("not found")
	ErrUnsupported = errors.ErrUnsupported
	ErrWrongFamily = errors.New("wrong address family")
	FIXME          = errors.New("FIXME")
)
