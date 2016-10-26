// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package kutils provides kernel admin commands.
package kutils

import (
	"github.com/platinasystems/go/kutils/cmdline"
	"github.com/platinasystems/go/kutils/dmesg"
	"github.com/platinasystems/go/kutils/iminfo"
	"github.com/platinasystems/go/kutils/insmod"
	"github.com/platinasystems/go/kutils/kexec"
	"github.com/platinasystems/go/kutils/lsmod"
	"github.com/platinasystems/go/kutils/reboot"
	"github.com/platinasystems/go/kutils/rmmod"
	"github.com/platinasystems/go/kutils/watchdog"
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
