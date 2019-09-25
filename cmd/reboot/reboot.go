// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package reboot

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/internal/kexec"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (Command) String() string { return "reboot" }

func (Command) Usage() string { return "reboot" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		"en_US.UTF-8": "reboot system",
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) (err error) {
	kexec.Prepare()

	err = c.g.Main("umount", "-a")
	if err != nil {
		fmt.Printf("Error unmounting all drives: %s\n", err)
		return errors.New("To force, use umount -a -f and try again")
	}

	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)

	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}
