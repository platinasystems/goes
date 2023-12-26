// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"errors"
	"runtime"
)

var (
	ErrCantTAP         = errors.New(runtime.GOOS + " can't TAP")
	ErrCantPersist     = errors.New(runtime.GOOS + " can't persist")
	ErrCantChangeOwner = errors.New(runtime.GOOS + " can't change owner")
	ErrCantChangeGroup = errors.New(runtime.GOOS + " can't change group")
	FIXME              = errors.New("unsupported on " + runtime.GOOS)
)
