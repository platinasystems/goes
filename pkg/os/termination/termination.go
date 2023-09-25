// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package termination

import (
	"os"
	"syscall"
)

var Signals = []os.Signal{
	os.Interrupt,
	os.Signal(syscall.SIGTERM),
}
