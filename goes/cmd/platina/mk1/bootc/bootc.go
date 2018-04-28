// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'bootc' client used for auto-install
// runs in goes-coreboot

package bootc

import (
	"fmt"
	"strconv"

	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootd"
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
	mip := getMasterIP()
	name := ""

	switch c {
	case 1:
		mac := getMAC()
		ip := getIP()
		if _, name, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(name)
	case 2:
		mac := getMAC2()
		ip := getIP2()
		if _, name, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(name)
	case 3:
		mac := getMAC3()
		ip := getIP3()
		if _, name, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(name)
	case 4:
		if err = dumpvars(mip); err != nil {
			return err
		}
	case 5:
		if err = dashboard(mip); err != nil {
			return err
		}
	case 6:
		if err = getnumclients(mip); err != nil {
			return err
		}
	case 7:
		if err = getclientdata(mip, 3); err != nil {
			return err
		}
	case 8:
		if err = getscript(mip, "testscript"); err != nil {
			return err
		}
	case 9:
		if err = getbinary(mip, "test.bin"); err != nil {
			return err
		}
	case 10: //run script
		if err = runScript("testscript"); err != nil {
			return err
		}
	case 11:
		if err = test404(mip); err != nil {
			return err
		}
	case 12:
		if err = dashboard("192.168.101.129"); err != nil {
			return err
		}
	default:
		fmt.Println("no command...")
	}
	return nil
}

//*/

func boot() (err error) { // Coreboot "init"
	mip := getMasterIP()
	mac := getMAC()
	ip := getIP()
	reply := 0
	//TODO [2] ADD FASTER TIMEOUT
	reply, _, err = register(mip, mac, ip)
	if err != nil || reply != bootd.RegReplyRegistered {
		reply, _, err = register(mip, mac, ip)
		if err != nil || reply != bootd.RegReplyRegistered {
			return err // fall into grub
		}
	}

	// TODO TRY REAL REGISTRATION TO SERVER BOLT IN OF BOOTC TO GOES INIT
	// TODO run install script (format, install debian, etc. OR just boot)
	// TODO if debian install fails ==> try again
	// TODO [2] REGISTER TIMEOUT
	// TODO READ the /boot directory into slice, bootd store last known good booted image
	// TODO [3] boot grub(GRUB TO TELL WHAT ITS BOOTING), give me your images/BOOT THIS IMAGE, ASK SCRIPT TO RUN/RUN IT
	// TODO BOOTC, BOOTD /etc/MASTER logic , bootc runs if no /etc/MASTER file, bootd runsi if /etc/MASTER (filesystem is not up btw)
	// TODO bootd state machines
	// TODO add test infra, with 100 units
	// TODO master to trigger client reset
	// TODO CB to boot new goes payload
	// TODO goes formats SDA2, installs debian use INSTALL/PRESEED
	// TODO ADD LOCATION OF ToR -- how?

	return nil
}

func runScript(name string) (err error) {
	// TODO check if script exists

	// TODO run script

	return nil
}
