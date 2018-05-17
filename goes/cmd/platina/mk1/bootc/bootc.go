// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

import (
	"fmt"

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

func (Command) Usage() string {
	return `bootc [register] [bootc] [dumpvars] [dashboard] [initcfg]\n
	[getnumclients] [getclientdata] [getscript] [getbinary] [testscript]\n
	[test404] [dashboard9] [setgrub] [clrgrub] [setinstall] [clrinstall]\n
	[readcfg] [setip] [setnetmask] [setgateway] [setkernel6] [setinitrd6]`
}

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

func (c *Command) Main(args ...string) (err error) {
	if len(args) == 0 {
		return fmt.Errorf("args: missing")
	}

	mip := getMasterIP()
	name := ""

	switch args[0] {
	case "register":
		mac := getMAC()
		ip := getIP()
		if _, name, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(name)
	case "bootc":
		if Bootc() {
			err = c.g.Main("source", "kexe") //FIXME MOVE BACK TO logic.go
			if err != nil {
				return err
			}
			return
		} else {
			return // fall through to grub
		}
	case "dumpvars":
		if err = dumpvars(mip); err != nil {
			return err
		}
	case "dashboard":
		if err = dashboard(mip); err != nil {
			return err
		}
	case "initcfg":
		initCfg()
	case "getnumclients":
		if err = getnumclients(mip); err != nil {
			return err
		}
	case "getclientdata":
		if err = getclientdata(mip, 3); err != nil {
			return err
		}
	case "getscript":
		if err = getscript(mip, "testscript"); err != nil {
			return err
		}
	case "getbinary":
		if err = getbinary(mip, "test.bin"); err != nil {
			return err
		}
	case "testscript":
		if err = runScript("testscript"); err != nil {
			return err
		}
	case "test404":
		if err = test404(mip); err != nil {
			return err
		}
	case "dashboard9":
		if err = dashboard("192.168.101.129"); err != nil {
			return err
		}
	case "setgrub":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		setGrubBit()
	case "clrgrub":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		clrGrubBit()
	case "setinstall":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		setReInstallBit()
	case "clrinstall":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		clrReInstallBit()
	case "readcfg":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		fmt.Printf("%+v\n", Cfg)
	case "setip":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		Cfg.MyIpAddr = args[1]
		if err := writeCfg(); err != nil {
			fmt.Println("boot.cfg - error writing configuration, run grub")
			return err
		}
	case "setnetmask":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		Cfg.MyNetmask = args[1]
		if err := writeCfg(); err != nil {
			fmt.Println("boot.cfg - error writing configuration, run grub")
			return err
		}
	case "setgateway":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		Cfg.MyGateway = args[1]
		if err := writeCfg(); err != nil {
			fmt.Println("boot.cfg - error writing configuration, run grub")
			return err
		}
	case "setkernel6":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		Cfg.Sda6K = args[1]
		if err := writeCfg(); err != nil {
			fmt.Println("boot.cfg - error writing configuration, run grub")
			return err
		}
	case "setinitrd6":
		if err := readCfg(); err != nil {
			fmt.Println("boot.cfg - error reading configuration, run grub")
			return err
		}
		Cfg.Sda6I = args[1]
		if err := writeCfg(); err != nil {
			fmt.Println("boot.cfg - error writing configuration, run grub")
			return err
		}
	default:
	}
	return nil
}
