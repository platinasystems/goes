// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"errors"
	"fmt"
)

var (
	ErrUnavailable   = errors.New("unavailable")
	ErrInvalidAddr   = errors.New("invalid address")
	ErrInvalidLength = errors.New("invalid length")
	ErrInvalidPrefix = errors.New("invalid prefix")
	ErrNoAdmin       = fmt.Errorf("admin %w", ErrUnavailable)
	ErrNoIPv6        = fmt.Errorf("ipv6 %w", ErrUnavailable)
)
