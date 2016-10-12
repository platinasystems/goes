// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package kexec

import (
	"fmt"
	"io/ioutil"

	"github.com/platinasystems/go/fit"
	"github.com/platinasystems/go/flags"
	"github.com/platinasystems/go/parms"
	"github.com/platinasystems/goes"
)

type kexec struct{}

func New() kexec { return kexec{} }

func (kexec) String() string { return "kexec" }
func (kexec) Usage() string  { return "kexec -l IMAGE" }

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

func (kexec kexec) Main(args ...string) error {
	flag, args := flags.New(args, "-e", "-l")
	parm, args := parms.New(args, "-c", "-l")

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}
	if image := parm["-l"]; len(image) > 0 {
		err := kexec.load(image, parm["-c"])
		if err != nil {
			return err
		}
	}

	if flag["-e"] {
		err := fit.KexecRebootSyscall()
		if err != nil {
			return err
		}
	}
	return nil
}

func (kexec) load(image, configName string) error {
	b, err := ioutil.ReadFile(image)
	if err != nil {
		return err
	}

	fit := fit.Parse(b)

	if len(configName) == 0 {
		configName = fit.DefaultConfig
	}

	config := fit.Configs[configName]
	config.BaseAddr = 0x60008000

	return fit.KexecLoadConfig(config, 0x0)
}
