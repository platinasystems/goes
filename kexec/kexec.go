// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package kexec

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/fit"
	kexecSyscall "github.com/platinasystems/goes/kexec"
	"github.com/platinasystems/goes/coreutils/reboot"
	"github.com/platinasystems/oops"
	"io/ioutil"
	"syscall"
	"os"
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
		"-c": goes.Parm{"COMMANDLINE", nil, ""},
		"-i": goes.Parm{"INITRAMFS", goes.Complete.File, ""},
		"-k": goes.Parm{"KERNEL", goes.Complete.File, ""},
		"-l": goes.Parm{"IMAGE", goes.Complete.File, ""},
	        "-x": goes.Parm{"CONFIGURATION", nil, ""},
	}
}

func (p *kexec) Main(args ...string) {
	var err error
	p.flags, args = p.Flags().Parse(args)
	p.parms, args, err = p.Parms().Parse(args)
	if err != nil {
		p.Panic(err)
	}

	if len(args) > 0 {
		p.Panic(args[0:], ": unexpected")
	}

	if image := p.parms["-l"]; len(image) > 0 {
		p.loadFit(image)
	}

	if kernel := p.parms["-k"]; len(kernel) > 0 {
		p.loadKernel(kernel)
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

func (p *kexec) loadFit(image string) {
	b, err := ioutil.ReadFile(image)
	if err != nil {
		p.Panic(err)
	}

	fit := fit.Parse(b)

	configName := p.parms["-x"]
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

func (p *kexec) loadKernel(kernel string) {
	k, err := os.Open(kernel)
	if err != nil {
		p.Panic(err)
	}
	defer k.Close()

	initramfs := p.parms["-i"]
	if len(initramfs) == 0 {
		p.Panic("Initramfs (-i) must be specified")
	}

	i, err := os.Open(initramfs)
	if err != nil {
		p.Panic(err)
	}
	defer i.Close()

	err = kexecSyscall.FileLoad(k, i, p.parms["-c"], 0)
	if err != nil {
		p.Panic(err)
	}
}
