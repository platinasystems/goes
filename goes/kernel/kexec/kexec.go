// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package kexec

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"syscall"

	"github.com/platinasystems/go/goes/internal/flags"
	"github.com/platinasystems/go/goes/internal/parms"
	"github.com/platinasystems/go/goes/kernel/internal/fit"
	"github.com/platinasystems/go/kexec"
)

const Name = "kexec"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [OPTIONS]..." }

func (cmd) Main(args ...string) error {
	var err error
	flag, args := flags.New(args, "-e", "-f")
	parm, args := parms.New(args, "-c", "-i", "-k", "-l", "-x")

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	if image := parm["-l"]; len(image) > 0 {
		err = loadFit(image, parm["-x"])
		if err != nil {
			return err
		}
	}

	if kernel := parm["-k"]; len(kernel) > 0 {
		err = loadKernel(kernel, parm["-i"], parm["-c"])
		if err != nil {
			return err
		}
	}

	if flag["-e"] || flag["-f"] {
		if !flag["-f"] {
			kexec.Prepare()
		}
		err = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)
	}
	return err
}

func loadFit(image, x string) error {
	b, err := ioutil.ReadFile(image)
	if err != nil {
		return err
	}

	fit := fit.Parse(b)

	if len(x) == 0 {
		x = fit.DefaultConfig
	}
	config := fit.Configs[x]
	config.BaseAddr = 0x60008000

	return fit.KexecLoadConfig(config, 0x0)
}

func loadKernel(kernel, initramfs, cmdline string) error {
	k, err := os.Open(kernel)
	if err != nil {
		return err
	}
	defer k.Close()

	if len(initramfs) == 0 {
		return errors.New("Initramfs (-i) must be specified")
	}

	i, err := os.Open(initramfs)
	if err != nil {
		return err
	}
	defer i.Close()

	return kexec.FileLoad(k, i, cmdline, 0)
}

// FIXME add Apropos() and Man()
