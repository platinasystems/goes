// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package kexec

import (
	"os"
	"syscall"
)

func Prepare() {
	for _, f := range []*os.File{
		os.Stdout,
		os.Stderr,
	} {
		syscall.Fsync(int(f.Fd()))
	}
	syscall.Sync()
}
