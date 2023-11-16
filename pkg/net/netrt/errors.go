// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "errors"

var (
	ErrNoDst = errors.New("missing destination <addr|prefix>")
	FIXME    = errors.New("FIXME")
)
