// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !unix

package vpn

import (
	"errors"

	"golang.org/x/sys/unix"
)

func ErrIsNOENT(err error) bool {
	return errors.Is(err, unix.ENOENT)
}
