// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcert

import (
	"errors"
	"io/fs"
)

var (
	ErrInvalid    = fs.ErrInvalid
	ErrNilLeaf    = errors.New("nil leaf")
	ErrNoName     = errors.New("no name")
	ErrNoDNSNames = errors.New("no DNS names")
	ErrPrivateKey = errors.New("missing private key")
	ErrKeyType    = errors.New("key type mismatch")
	ErrKeyParm    = errors.New("key parameter mismatch")
)
