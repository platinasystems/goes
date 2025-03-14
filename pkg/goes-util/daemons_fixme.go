// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !android && !linux && !darwin && !ios && !netbsd && !openbsd

package goes_util

import (
	"errors"
	"os"
	"runtime"
)

const daemonCriterion = "TBD processes"

func Daemons() ([]*os.Process, error) {
	return nil, errors.New("FIXME " + runtime.GOOS)
}
