// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"errors"
	"io/fs"
	"runtime"
)

var (
	FIXME           = errors.New("FIXME")
	ErrCantTap      = errors.New(runtime.GOOS + " can't TAP")
	ErrExists       = errors.New("exists")
	ErrHasPI        = errors.New(runtime.GOOS + " has unwanted packet info")
	ErrIncomplete   = errors.New("incomplete")
	ErrInvalid      = fs.ErrInvalid
	ErrKeyType      = errors.New("key type mismatch")
	ErrKeyParm      = errors.New("key parameter mismatch")
	ErrNilLeaf      = errors.New("nil leaf")
	ErrNoDNSNames   = errors.New("no DNS names")
	ErrNoName       = errors.New("no name")
	ErrNotFound     = errors.New("not found")
	ErrPrivateKey   = errors.New("missing private key")
	ErrOccupied     = errors.New("occupied")
	ErrRange        = errors.New("out of range")
	ErrTooShort     = errors.New("too short")
	ErrUnavailable  = errors.New("unavailable")
	ErrUnconfirmed  = errors.New("unconfirmed")
	ErrUnderrun     = errors.New("underrun")
	ErrUnidentified = errors.New("unidentified")
	ErrUnnamed      = errors.New("unnamed")
)
