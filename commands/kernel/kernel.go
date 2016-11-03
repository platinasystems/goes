// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package kernel provides kernel admin commands.
package kernel

import (
	"github.com/platinasystems/go/commands/kernel/cmdline"
	"github.com/platinasystems/go/commands/kernel/dmesg"
	"github.com/platinasystems/go/commands/kernel/iminfo"
	"github.com/platinasystems/go/commands/kernel/insmod"
	"github.com/platinasystems/go/commands/kernel/kexec"
	"github.com/platinasystems/go/commands/kernel/lsmod"
	"github.com/platinasystems/go/commands/kernel/reboot"
	"github.com/platinasystems/go/commands/kernel/rmmod"
	"github.com/platinasystems/go/commands/kernel/watchdog"
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
		watchdog.New(),
	}
}
