// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !linux && !netbsd

package goes_util

import (
	"errors"
	"runtime"
)

func Daemons() (pids []int, err error) {
	err = errors.New("FIXME " + runtime.GOOS)
	return
}
