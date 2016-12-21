// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package kernel provides kernel admin commands.
package kernel

import (
	"github.com/platinasystems/go/goes/kernel/cmdline"
	"github.com/platinasystems/go/goes/kernel/dmesg"
	"github.com/platinasystems/go/goes/kernel/iminfo"
	"github.com/platinasystems/go/goes/kernel/insmod"
	"github.com/platinasystems/go/goes/kernel/kexec"
	"github.com/platinasystems/go/goes/kernel/lsmod"
	"github.com/platinasystems/go/goes/kernel/reboot"
	"github.com/platinasystems/go/goes/kernel/rmmod"
)

func New() []interface{} {
	return []interface{}{
		cmdline.New(),
		dmesg.New(),
		iminfo.New(),
		insmod.New(),
		kexec.New(),
		lsmod.New(),
		reboot.New(),
		rmmod.New(),
	}
}
