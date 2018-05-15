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

const BootcCfgFile = "/newroot/sda1/bootc.cfg"

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
	default:
	}
	return nil
}

var uuid1 string
var uuid6 string
var kexec1 string
var kexec6 string

func Bootc() {
	if err := readCfg(); err != nil {
		fmt.Println("boot.cfg - error reading configuration, run grub")
		return
	}
	if contactServer() == true {
		//FIXME
	}
	if cfg.grub == "grub" {
		fmt.Println("Grub Enable == TRUE, run grub")
		return
	}
	if err := formStrings(); err != nil {
		fmt.Println("error forming kexec strings, run grub")
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

	mip := getMasterIP()
	mac := getMAC()
	ip := getIP()
	reply := 0
	reply, _, err = register(mip, mac, ip)
	if err != nil || reply != bootd.RegReplyRegistered {
		reply, _, err = register(mip, mac, ip)
		if err != nil || reply != bootd.RegReplyRegistered {
			return false
		}
	}
	return true
}

func writeCfg() error {
	Cfg.GrubEnabled = false
	Cfg.ReInstallEnabled = false
	Cfg.IAmMaster = false
	Cfg.MyIpAddr = "192.168.101.129"
	Cfg.MyIpGWay = "192.168.101.1"
	Cfg.MyIpMask = "255.255.255.0"
	Cfg.MasterAddresses = []string{"a", "b", "c"}
	Cfg.ReInstallK = "/newroot/sda1/boot/vmlinuz-3.16.0-4-amd64"
	Cfg.ReInstallI = "/newroot/sda1/boot/initrd.img-3.16.0-4-amd64"
	Cfg.ReInstallC = "netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto"
	Cfg.Sda6K = "/newroot/sda1/boot/vmlinuz-3.16.0-4-amd64"
	Cfg.Sda6I = "/newroot/sda1/boot/initrd.img-3.16.0-4-amd64"
	Cfg.Sda6C = "::eth0:none"

	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
}

func readCfg() error {
	if _, err := os.Stat(BootcCfgFile); os.IsNotExist(err) {
		fmt.Println(BootcCfgFile + " does not exist, run grub")
		return err
	}
	dat, err := ioutil.ReadFile(BootcCfgFile)
	if err != nil {
		fmt.Println("error reading " + BootcCfgFile)
		return err
	}
	err = json.Unmarshal([]byte(dat), &Cfg2)
	if err != nil {
		fmt.Println("There was an error:", err)
		return err
	}
	fmt.Println(Cfg2)
	return nil
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

	kexec1 = "kexec -k " + Cfg.ReInstallK + " -i " + Cfg.ReInstallI + " -c "
	kexec1 += "'root=UUID=" + uuid1 + " console=ttyS0,115200 "
	kexec1 += Cfg.ReInstallC
	kexec1 += "' -e"
	kexec6 = "kexec -k " + Cfg.Sda6K + " -i " + Cfg.Sda6I + " -c "
	kexec6 += "'root=UUID=" + uuid6 + " console=ttyS0,115200 "
	kexec6 += "ip=" + Cfg.MyIpAddr + "::" + Cfg.MyIpGWay + ":" + Cfg.MyIpMask + Cfg.Sda6C
	kexec6 += "' -e"

	Cfg.GrubEnabled = true // set grub flag to avoid bootc loop
	jsonInfo, err := json.Marshal(Cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(BootcCfgFile, jsonInfo, 0644)
	if err != nil {
		return err
	}
	return nil
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

func runScript(name string) (err error) {
	return nil
}
