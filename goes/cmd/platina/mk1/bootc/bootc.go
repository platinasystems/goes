// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootd"
	"github.com/platinasystems/go/goes/lang"
)

const CfgFile = "/newroot/sda1/bootc.cfg"

///* for testing
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
	case 0:
		mac := getMAC()
		ip := getIP()
		if _, name, err = register(mip, mac, ip); err != nil {
			return err
		}
		fmt.Println(name)
	case 1:
		uuid, err = readUUID("sda1")
		if err != nil {
			return err
		}

		kexc := "kexec -k /newroot/sda1/boot/vmlinuz -i /newroot/sda1/boot/initrd.gz -c 'root=UUID=" + uuid + " console=ttyS0,115200 netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto' -e"
		fmt.Println(kexc)

		d1 := []byte(kexc)
		err := ioutil.WriteFile("kexec1", d1, 0644)
		if err != nil {
			fmt.Println("error writing kexec1")
		}

		err = c.g.Main("source", "kexec1")
		if err != nil {
			return err
		}
	case 2:
		uuid, err = readUUID("sda1")
		if err != nil {
			return err
		}

		//read vmlinuz name

		//read initrd name

		kexc := "kexec -k /newroot/sda1/boot/vmlinuz-3.16.0-4-amd64 -i /newroot/sda1/boot/initrd.img-3.16.0-4-amd64 -c 'root=UUID=" + uuid + " console=ttyS0,115200 netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto' -e"
		fmt.Println(kexc)

		d1 := []byte(kexc)
		err := ioutil.WriteFile("kexec1", d1, 0644)
		if err != nil {
			fmt.Println("error writing kexec1")
		}

		err = c.g.Main("source", "kexec1")
		if err != nil {
			return err
		}
		/*	mac := getMAC2()
			ip := getIP2()
			if _, name, err = register(mip, mac, ip); err != nil {
				return err
			}

			fmt.Println(name)
		*/
	case 3:
		Bootc()
		return // fall through to grub
		/*
			mac := getMAC3()
			ip := getIP3()
			if _, name, err = register(mip, mac, ip); err != nil {
				return err
			}
			fmt.Println(name)
		*/
	case 4:
		if err = dumpvars(mip); err != nil {
			return err
		}
	case 5:
		if err = dashboard(mip); err != nil {
			return err
		}
	case 6:
		uuid, err = readUUID("sda6")
		if err != nil {
			return err
		}

		kexc := "kexec -k /newroot/sda6/boot/vmlinuz-3.16.0-4-amd64 -i /newroot/sda6/boot/initrd.img-3.16.0-4-amd64 -c 'root=UUID=" + uuid + " console=ttyS0,115200 netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto' -e"
		fmt.Println(kexc)

		d1 := []byte(kexc)
		err := ioutil.WriteFile("kexec1", d1, 0644)
		if err != nil {
			fmt.Println("error writing kexec1")
		}

		err = c.g.Main("source", "kexec1")
		if err != nil {
			return err
		}
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
	case 11: //run script
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
	case 14:
		fmt.Println("kexec -k /newroot/sda1/boot/vmlinuz -i /newroot/sda1/boot/initrd.gz -c 'console=ttyS0,115200 keymap=us debian-installer/locale=en_US netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto' -e")
	case 15:
		fmt.Println("kexec -k /newroot/sda1/boot/vmlinuz-3.16.0-4-amd64 -i /newroot/sda1/boot/initrd.img-3.16.0-4-amd64 -c 'console=ttyS0,115200 keymap=us debian-installer/locale=en_US netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto' -e")
	default:
		fmt.Println("enter 1 for sda1 install, 6 for normal sda6 boot")
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
	// TODO [2] REGISTER TIMEOUT
	// TODO READ the /boot directory into slice, bootd store last known good booted image
	// TODO [3] boot grub(GRUB TO TELL WHAT ITS BOOTING), give me your images/BOOT THIS IMAGE, ASK SCRIPT TO RUN/RUN IT
	// TODO BOOTC, BOOTD /etc/MASTER logic , bootc runs if no /etc/MASTER file, bootd runsi if /etc/MASTER (filesystem is not up btw)
	// TODO bootd state machines
	// TODO add test infra, with 100 units
	// TODO master to trigger client reset

	return nil
}

func runScript(name string) (err error) {
	// TODO check if script exists

	// TODO run script

	return nil
}

var uuid1 string
var uuid6 string
var kexec1 string
var kexec6 string

func Bootc() {
	if err := readcfg(); err != nil {
		fmt.Println("boot.cfg - error reading configuration, run grub")
		return
	}
	if contactServer() == true {
	}
	if cfg.grub == "grub" {
		fmt.Println("Grub Enable == TRUE, run grub")
		return
	}
	err := formStrings()
	if err != nil {
		return
	}
	if cfg.reinstall == "reinstall" {
		fmt.Println("Re-install == TRUE, run installer")
		kexec(kexec1)
	} else {
		fmt.Println("Re-install == FALSE, boot sda6")
		kexec(kexec6)
	}
}

func contactServer() bool {
	return false
}

func writecfg() error {
	Cfg.GrubEnabled = false
	Cfg.ReInstallEnabled = false
	Cfg.IAmMaster = false
	Cfg.MyIpAddr = "192.168.101.129"
	Cfg.MyIpGWay = "192.168.101.1"
	Cfg.MyIpMask = "255.255.255.0"
	Cfg.MasterAddresses = "slice"
	Cfg.ReInstallK = "/newroot/sda1/boot/vmlinuz-3.16.0-4-amd64"
	Cfg.ReInstallI = "/newroot/sda1/boot/initrd.img-3.16.0-4-amd64"
	Cfg.ReInstallC = "ip=192.168.101.129::192.168.101.1:255.255.255.192::eth0:none"
	Cfg.Sda6K = "/newroot/sda1/boot/vmlinuz-3.16.0-4-amd64"
	Cfg.Sda6I = "/newroot/sda1/boot/initrd.img-3.16.0-4-amd64"
	Cfg.Sda6C = "ip=192.168.101.129::192.168.101.1:255.255.255.192::eth0:none"

	// MARSHALL
	// WRITE
}

func readcfg() error {
	if _, err := os.Stat(CfgFile); os.IsNotExist(err) {
		fmt.Println("bootc.cfg does not exist, run grub")
		return
	}
	dat, err := ioutil.ReadFile(Bootccfg)
	if err != nil {
		fmt.Println("error reading bootc.cfg.")
		return "", err
	}
	dat1 := strings.Split(string(dat), "\n")
	for i, j := range dat1 {
		fmt.Println(i, j)
	}
	// UNMARSHALL
}

func formStrings() error {
	uuid1, err = readUUID("sda1")
	if err != nil {
		return err
	}
	uuid6, err = readUUID("sda6")
	if err != nil {
		return err
	}
	// form kexec strings
	// set grub bit, write file
}

func readUUID(string partition) (string uuid, error err) {
	dat, err := ioutil.ReadFile("/newroot/" + partition + "/etc/fstab")
	if err != nil {
		return "", err
	}
	dat1 := strings.Split(string(dat), "UUID=")
	dat2 := strings.Split(dat1[2], "/")
	dat3 := []byte(dat2[0])
	len3 := len(dat3) - 1
	dat4 := string(dat3[0:len3])
	uuid := string(dat4)
	return uuid, nil
}

/*
files, err := ioutil.ReadDir("./")
    if err != nil {
        log.Fatal(err)
    }

    for _, f := range files {
            fmt.Println(f.Name())
    }
*/
