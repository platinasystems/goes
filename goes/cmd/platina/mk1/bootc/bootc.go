// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'bootc' client used for auto-install
// runs as coreboot payload (kernel + initrd + goes)

// TODO add test infra, with 100 units
// TODO installing apt-gets support
// TODO master to trigger client reset
// TODO CB to boot new goes payload
// TODO goes formats SDA2, installs debian use INSTALL/PRESEED

package bootc

import (
	"fmt"
	"strconv"

	"github.com/platinasystems/go/goes/lang"
)

///* for testing
type Command struct{}

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

func (Command) Main(args ...string) (err error) {
	if len(args) == 0 {
		return fmt.Errorf("args: missing")
	}

	c, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}
	s := ""
	mip := getMasterIP()

	switch c {
	case 1:
		mac := getMAC()
		ip := getIP()
		if s, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(s)
	case 2:
		mac := getMAC2()
		ip := getIP2()
		if s, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(s)
	case 3:
		mac := getMAC3()
		ip := getIP3()
		if s, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(s)
	case 4:
		if err = dumpVars(mip); err != nil {
			return err
		}
	case 5:
		if err = dashboard(mip); err != nil {
			return err
		}
	case 6:
		if err = test404(mip); err != nil {
			return err
		}
	default:
		fmt.Println("no command...")
	}
	return nil
}

//*/

func boot() (err error) { // Coreboot goes "init"
	mip := getMasterIP()
	mac := getMAC()
	ip := getIP()
	// TODO [3] try register 3 times with delay in between
	if _, err = register(mip, mac, ip); err != nil {
		// register failed: forget master, fall into grub-equivalent
		return err
	}
	// TODO run script (format, install debian, etc. OR just boot)

	// TODO if debian install fails ==> try again, PXE boot?

	return nil
}
