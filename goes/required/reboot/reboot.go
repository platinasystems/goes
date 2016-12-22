// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package reboot

import (
	"syscall"

	"github.com/platinasystems/go/goes/internal/kexec"
)

const Name = "reboot"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd) Main(args ...string) error {
	kexec.Prepare()

	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)

	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "reboot system",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	reboot - reboots system

SYNOPSIS
	reboot

DESCRIPTION
	Reboots system.`,
	}
}
