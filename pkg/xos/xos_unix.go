// Copyright © 2015-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package xos

import (
	. "os"

	"golang.org/x/sys/unix"
)

var Termination = []Signal{Interrupt, Signal(unix.SIGTERM)}
