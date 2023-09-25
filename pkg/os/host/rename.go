// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !aix && !linux && !solaris

package host

import (
	"fmt"
	"runtime"
)

func rename(string) error {
	return fmt.Errorf("can't set %s hostname", runtime.GOOS)
}
