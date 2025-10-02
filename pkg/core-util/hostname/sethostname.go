// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !linux

package hostname

import (
	"fmt"
	"runtime"
)

func Sethostname(string) error {
	return fmt.Errorf("can't set %s hostname", runtime.GOOS)
}
