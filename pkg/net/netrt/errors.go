// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "errors"

var (
	ErrNoDst     = errors.New("missing destination <addr|prefix>")
	ErrSRCH      = errors.New("not in table")
	ErrBUSY      = errors.New("entry in use")
	ErrNOBUFS    = errors.New("not enough memory")
	ErrADDRINUSE = errors.New("gateway uses the same route")
	ErrEXIST     = errors.New("route already in table")
	FIXME        = errors.New("FIXME")
)
