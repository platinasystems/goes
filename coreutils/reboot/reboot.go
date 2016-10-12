// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package reboot

import (
	"os"
	"syscall"
)

type reboot struct{}

func New() reboot { return reboot{} }

func (reboot) String() string { return "reboot" }
func (reboot) Usage() string  { return "reboot" }

func (reboot) Main(args ...string) error {
	for _, f := range []*os.File{
		os.Stdout,
		os.Stderr,
	} {
		syscall.Fsync(int(f.Fd()))
	}
	syscall.Sync()
	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func (reboot) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "reboot system",
	}
}

func (reboot) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	reboot - reboots system

SYNOPSIS
	reboot

DESCRIPTION
	Reboots system.`,
	}
}
