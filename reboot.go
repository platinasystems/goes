// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package reboot

import (
	"os"
	"syscall"

	"github.com/platinasystems/oops"
)

type reboot struct{ oops.Id }

var Reboot = &reboot{"reboot"}

func (*reboot) Usage() string { return "reboot" }

func Prepare() {
	for _, f := range []*os.File{
		os.Stdout,
		os.Stderr,
	} {
		syscall.Fsync(int(f.Fd()))
	}
	syscall.Sync()
}

func (p *reboot) Main(args ...string) {
	Prepare()

	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)

	err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	if err != nil {
		p.Panic(err)
	}
}
