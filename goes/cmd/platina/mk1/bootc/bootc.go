// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

import (
	"fmt"
	"strconv"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
)

func New() *Command { return new(Command) }

type Command struct {
	// Machines may use Hook to run something before redisd and other
	// daemons.
	Hook func() error

	// Machines may use ConfHook to run something after all daemons start
	// and before source of start command script.
	ConfHook func() error

	// GPIO init hook for machines than need it
	ConfGpioHook func() error

	g *goes.Goes
}

func (Command) String() string { return "bootc" }

func (Command) Usage() string { return "bootc" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "boot client hook to communicate with tor master",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
description
	the bootc command is for debugging bootc client.`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

// FIXME change args to text
func (c *Command) Main(args ...string) (err error) {
	if len(args) == 0 {
		fmt.Println("enter 1 for sda1 install, 6 for normal sda6 boot")
		return fmt.Errorf("args: missing")
	}

	cm, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}
	mip := getMasterIP()
	name := ""

	switch cm {
	case 1:
		mac := getMAC()
		ip := getIP()
		if _, name, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(name)
	case 3:
		Bootc() // if returns, fall through to grub
		return nil
	case 4:
		if err = dumpvars(mip); err != nil {
			return err
		}
	case 5:
		if err = dashboard(mip); err != nil {
			return err
		}
	case 6:
		writeCfg()
	case 7:
		if err = getnumclients(mip); err != nil {
			return err
		}
	case 8:
		if err = getclientdata(mip, 3); err != nil {
			return err
		}
	case 9:
		if err = getscript(mip, "testscript"); err != nil {
			return err
		}
	case 10:
		if err = getbinary(mip, "test.bin"); err != nil {
			return err
		}
	case 11:
		if err = runScript("testscript"); err != nil {
			return err
		}
	case 12:
		if err = test404(mip); err != nil {
			return err
		}
	case 13:
		if err = dashboard("192.168.101.129"); err != nil {
			return err
		}
	case 20:
		setGrubBit()
	case 21:
		clrGrubBit()
	case 22:
		setReInstallBit()
	case 23:
		clrReInstallBit()
	case 24:
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		fmt.Println(Cfg)
	default:
	}
	return nil
}
