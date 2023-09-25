// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package kexec

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/internal/fit"
	"github.com/platinasystems/goes/internal/kexec"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "kexec" }

func (*Command) Usage() string { return "kexec [OPTIONS]..." }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "load a new kernel for later execution",
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	flag, args := flags.New(args, "-e", "-f")
	parm, args := parms.New(args, "-c", "-i", "-k", "-l", "-x")

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	kc, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		fmt.Printf("Warning: unable to read kernel command line: %v\n",
			err)
	}

	kcstr := strings.TrimSpace(string(kc))

	cmdline := parm.ByName["-c"]
	if cmdline != "" {
		if cmdline[0] == '+' {
			cmdline = kcstr + " " + cmdline[1:]
		}
	} else {
		cmdline = kcstr
	}

	if image := parm.ByName["-l"]; len(image) > 0 {
		err = loadFit(image, parm.ByName["-x"], cmdline)
		if err != nil {
			return err
		}
	}

	if kernel := parm.ByName["-k"]; len(kernel) > 0 {
		err = loadKernel(kernel, parm.ByName["-i"], cmdline)
		if err != nil {
			return err
		}
	}

	if flag.ByName["-e"] || flag.ByName["-f"] {
		if !flag.ByName["-f"] {
			err = c.g.Main("reboot")
		} else {
			err = c.g.Main("reboot", "-f")
		}
	}
	return err
}

func loadFit(image, x, cmdline string) error {
	b, err := ioutil.ReadFile(image)
	if err != nil {
		return err
	}

	fit := fit.Parse(b)

	if len(x) == 0 {
		x = fit.DefaultConfig
		if len(x) == 0 {
			return fmt.Errorf("No default image, use the -x option")
		}
	}

	config := fit.Configs[x]

	if config == nil {
		return fmt.Errorf("Configuration %s not found", x)
	}
	return fit.KexecLoadConfig(config, cmdline)
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
