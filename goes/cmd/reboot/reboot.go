// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package reboot

import (
	"syscall"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/kexec"
)

const (
	Name    = "reboot"
	Apropos = "reboot system"
	Usage   = "reboot"
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
	kexec.Prepare()

	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)

	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	"en_US.UTF-8": Apropos,
}
