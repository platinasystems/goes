// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import "errors"

var (
	ErrIncomplete     = errors.New("incomplete")
	ErrNoExchange     = errors.New("missing <exchange>")
	ErrNoCommand      = errors.New("missing <command>")
	ErrUnexpectedArgs = errors.New("unexpected argument(s)")
	ErrNoDNSNames     = errors.New("no DNS names")
	ErrNoName         = errors.New("no name")
)
