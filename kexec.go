// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package kexec

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/fit"
	"github.com/platinasystems/goes/coreutils/reboot"
	"github.com/platinasystems/oops"
	"io/ioutil"
	"syscall"
)

type kexec struct {
	oops.Id
	flags map[string]bool
	parms map[string]string
}

var Kexec = &kexec{Id: "kexec"}

func (*kexec) Usage() string {
	return "kexec -l IMAGE"
}

func (*kexec) Flags() goes.Flags {
	return goes.Flags{
		"-e": goes.Flag{},
		"-f": goes.Flag{},
	}
}

func (*kexec) Parms() goes.Parms {
	return goes.Parms{
		"-c": goes.Parm{"IMAGE", nil, ""},
		"-l": goes.Parm{"IMAGE", goes.Complete.File, ""},
	}
}

func (p *kexec) Main(args ...string) {
	var err error
	p.flags, args = p.Flags().Parse(args)
	p.parms, args, err = p.Parms().Parse(args)
	if err != nil {
		p.Panic(err)
	}

	if image := p.parms["-l"]; len(image) > 0 {
		switch len(args) {
		case 0:
			p.load(image)
		default:
			p.Panic(args[0:], ": unexpected")
		}
	}

	if p.flags["-e"] || p.flags["-f"] {
		if !p.flags["-f"] {
			reboot.Prepare()
		}
		err = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
		if err != nil {
			p.Panic(err)
		}
	}
}

func (p *kexec) load(image string) {
	b, err := ioutil.ReadFile(image)
	if err != nil {
		p.Panic(err)
	}

	fit := fit.Parse(b)

	configName := p.parms["-c"]
	if len(configName) == 0 {
		configName = fit.DefaultConfig
	}

	config := fit.Configs[configName]
	config.BaseAddr = 0x60008000;

	err = fit.KexecLoadConfig(config, 0x0)

	if err != nil {
		p.Panic(err)
	}
}
