// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package selection

import "strings"

type Error struct {
	Path []string
	Err  error
}

func (e Error) Error() string {
	return strings.Join(e.Path, ":") + ": " + e.Err.Error()
}

func (e Error) Unwrap() error { return e.Err }
