// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import "errors"

var (
	ErrEmptyRequest      = errors.New("empty request")
	ErrExit              = errors.New("exit")
	ErrMissingAddress    = errors.New("missing <address>:<port>")
	ErrNameOrSKINotFound = errors.New("name or subject-key-id not found")
	ErrNoPeer            = errors.New("no peer certificates")
	ErrNoSubjectKeyId    = errors.New("missing subject-key-id")
	ErrNotNamePort       = errors.New("expect <name>:<port>")
	ErrNotOK             = errors.New("unsuccessful")
	ErrSKINotFound       = errors.New("subject-key-id not found")
	ErrUnexpectedArgs    = errors.New("unexpected argument(s)")
	ErrUnknownCommand    = errors.New("unknown command")
)
