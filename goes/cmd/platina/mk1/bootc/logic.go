// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootd"
	"github.com/platinasystems/go/internal/machine"
)

const (
	CorebootCfg = "/newroot/sda1/bootc.cfg"
	Sda1Cfg     = "/bootc.cfg"
	Sda6Cfg     = "/mnt/bootc.cfg"
	Mount       = "/mnt"
	Sda1        = "/mnt/sda1"
	Fstype      = "ext4"
	Zero        = uintptr(0)
)

var BootcCfgFile string
var uuid1 string
var uuid6 string
var kexec0 string
var kexec1 string
var kexec6 string

func Bootc() []string {
	if err := readCfg(); err != nil {
		fmt.Println("ERROR: couldn't read bootc.cfg => run grub")
		return []string{""}
	}

	if !serverAvail() {
		fmt.Println("INFO: server is not available, using local bootc.cfg")
	}

	fmt.Printf("INFO: Install = %v, BootSda1 = %v, BootSda6Cnt = %v\n",
		Cfg.Install, Cfg.BootSda1, Cfg.BootSda6Cnt)

	if !Cfg.Install && Cfg.BootSda6Cnt == 0 && !Cfg.BootSda1 {
		fmt.Println("INFO: !Install, !BootSda1, BootSda6Cnt==0 => run grub")
		return []string{""}
	}

	if Cfg.BootSda1 {
		if err := formString1(); err != nil {
			fmt.Println("ERROR: couldn't form kexec1 string => run grub")
			return []string{""}
		}
		if err := clrSda1Flag(); err != nil {
			fmt.Println("ERROR: could not clear BootSda1")
			return []string{""}
		}
		return []string{"kexec", "-k", Cfg.Sda1K,
			"-i", Cfg.Sda1I, "-c", kexec1, "-e"}
	}

	if Cfg.Install {
		if err := formString1(); err != nil {
			fmt.Println("ERROR: couldn't form kexec1 string => run grub")
			return []string{""}
		}
		if err := clrInstall(); err != nil {
			fmt.Println("ERROR: could not clear Install")
			return []string{""}
		}
		return []string{"kexec", "-k", Cfg.ReInstallK, "-i",
			Cfg.ReInstallI, "-c", kexec0, "-e"}
	}

	if Cfg.BootSda6Cnt > 0 {
		if err := formString6(); err != nil {
			fmt.Println("ERROR: couldn't form kexec6 string => run grub")
			return []string{""}
		}
		if err := decBootSda6Cnt(); err != nil {
			fmt.Println("ERROR: could not decrement BootSda6Cnt")
			return []string{""}
		}

		// FIXME copy script, goes, modify rclocal, script does "goes upgrade -k, -g"

		return []string{"kexec", "-k", Cfg.Sda6K,
			"-i", Cfg.Sda6I, "-c", kexec6, "-e"}
	}
	return []string{""}
}

func (c *Command) bootc() {
	if kexec := Bootc(); len(kexec) > 1 {
		err := c.Main(kexec...)
		fmt.Println(err)
	}
	return
}

func serverAvail() bool {
	return false

	mip := getMasterIP()
	mac := getMAC()
	ip := getIP()
	reply := 0
	reply, _, err := register(mip, mac, ip)
	if err != nil || reply != bootd.RegReplyRegistered {
		reply, _, err = register(mip, mac, ip)
		if err != nil || reply != bootd.RegReplyRegistered {
			return false
		}
	}
	return true
}

func initCfg() error {
	Cfg.Install = false
	Cfg.BootSda1 = false
	Cfg.BootSda6Cnt = 3
	Cfg.EraseSda6 = false
	Cfg.IAmMaster = false
	Cfg.MyIpAddr = "192.168.101.129"
	Cfg.MyGateway = "192.168.101.1"
	Cfg.MyNetmask = "255.255.255.0"
	Cfg.MasterAddresses = []string{"198.168.101.142"}
	Cfg.ReInstallK = "/newroot/sda1/boot/vmlinuz"
	Cfg.ReInstallI = "/newroot/sda1/boot/initrd.gz"
	Cfg.ReInstallC = `netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto locale=en_US preseed/file=/hd-media/preseed.cfg`
	Cfg.Sda1K = "/newroot/sda1/boot/vmlinuz-3.16.0-4-amd64"
	Cfg.Sda1I = "/newroot/sda1/boot/initrd.img-3.16.0-4-amd64"
	Cfg.Sda1C = "::eth0:none"
	Cfg.Sda6K = "/newroot/sda6/boot/vmlinuz-3.16.0-4-amd64"
	Cfg.Sda6I = "/newroot/sda6/boot/initrd.img-3.16.0-4-amd64"
	Cfg.Sda6C = "::eth0:none"
	Cfg.InitScript = false
	Cfg.InitScriptName = "sda6-init.sh"
	Cfg.ISO1Name = "debian-8.10.0-amd64-DVD-1.iso"
	Cfg.ISO1Desc = "Jessie debian-8.10.0"
	Cfg.ISO2Name = ""
	Cfg.ISO2Desc = ""
	Cfg.ISOlastUsed = 1
	err := writeCfg()
	if err != nil {
		return err
	}
	return nil
}

func setCfgPath() error {
	context := machine.Name
	if context == "platina-mk1" {
		//FIXME sda1 or sda6?
		context = "sda6"
	}
	fmt.Printf("context = %s\n", context)

	switch context {
	case "coreboot":
		BootcCfgFile = CorebootCfg
	case "sda1":
		BootcCfgFile = Sda1Cfg
	case "sda6":
		BootcCfgFile = Sda6Cfg
		if _, err := os.Stat(Mount); os.IsNotExist(err) {
			err := os.Mkdir(Mount, os.FileMode(0755))
			if err != nil {
				fmt.Printf("Error mkdir: %v", err)
				return err
			}
		}
		if err := syscall.Mount(Sda1, Mount, Fstype, Zero, ""); err != nil {
			fmt.Printf("Error mounting: %v", err)
		}
	default:
		return fmt.Errorf("ERROR: unknown machine could not form path")
	}
	return nil
}

func writeCfg() error {
	if err := setCfgPath(); err != nil {
		return err
	}

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
	if err := setCfgPath(); err != nil {
		return err
	}

	if _, err := os.Stat(BootcCfgFile); os.IsNotExist(err) {
		return err
	}
	dat, err := ioutil.ReadFile(BootcCfgFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(dat, &Cfg)
	if err != nil {
		return err
	}
	return nil
}

func formString1() (err error) {
	uuid1, err = readUUID("sda1")
	if err != nil {
		return err
	}
	kexec0 = "root=UUID=" + uuid1 + " console=ttyS0,115200 " + Cfg.ReInstallC
	kexec1 = "root=UUID=" + uuid1 + " console=ttyS0,115200 "
	kexec1 += "ip=" + Cfg.MyIpAddr + "::" + Cfg.MyGateway + ":" + Cfg.MyNetmask
	kexec1 += Cfg.Sda1C
	return nil
}

func formString6() (err error) {
	uuid6, err = readUUID("sda6")
	if err != nil {
		return err
	}
	kexec6 = "root=UUID=" + uuid6 + " console=ttyS0,115200 "
	kexec6 += "ip=" + Cfg.MyIpAddr + "::" + Cfg.MyGateway + ":" + Cfg.MyNetmask
	kexec6 += Cfg.Sda6C
	return nil
}

func readUUID(partition string) (uuid string, err error) {
	dat, err := ioutil.ReadFile("/newroot/" + partition + "/etc/fstab")
	if err != nil {
		return "", err
	}
	dat1 := strings.Split(string(dat), "UUID=")
	dat2 := strings.Split(dat1[2], "/")
	dat3 := []byte(dat2[0])
	len3 := len(dat3) - 1
	dat4 := string(dat3[0:len3])
	uuid = string(dat4)
	return uuid, nil
}

func setSda6Count(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	i, err := strconv.Atoi(x)
	if err != nil {
		return err
	}
	Cfg.BootSda6Cnt = i
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

func clrSda6Count() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.BootSda6Cnt = 0
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

func decBootSda6Cnt() error {
	if err := readCfg(); err != nil {
		return err
	}
	x := Cfg.BootSda6Cnt
	if x > 0 {
		x--
	}
	Cfg.BootSda6Cnt = x
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

func setInstall() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Install = true
	Cfg.BootSda6Cnt = 3
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

func clrInstall() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.BootSda1 = false
	Cfg.EraseSda6 = false
	Cfg.Install = false
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

func setSda1Flag() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.BootSda1 = true
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

func clrSda1Flag() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.BootSda1 = false
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

func setIp(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.MyIpAddr = x
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

func setNetmask(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.MyNetmask = x
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

func setGateway(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.MyGateway = x
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

func setKernel(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Sda6K = x
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

func setInitrd(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Sda6I = x
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

func setInitScriptName(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.InitScriptName = x
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

func setInitScript() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.InitScript = true
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

func clrInitScript() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.InitScript = false
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

func wipe() error {
	d1 := []byte(`#!/bin/bash\necho -e "d\n6\nw\n" | /sbin/fdisk /dev/sda\n`)
	if err := ioutil.WriteFile("/tmp/EEOF", d1, 0655); err != nil {
		return err
	}

	fmt.Println("Please wait...reinstalling debian on sda6")
	if err := setInstall(); err != nil {
		return err
	}
	cmd := exec.Command("/tmp/EEOF")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("cmd.Run() failed with %s\n", err)
	} else {
		fmt.Printf("combined out:\n%s\n", string(out))
	}
	reboot()
	return nil
}

func runScript(name string) (err error) {
	return nil
}

func reboot() error {
	fmt.Print("\nWILL REBOOT in 1 minute... Please login again\n")
	u, err := exec.Command("shutdown", "-r", "+1").Output()
	fmt.Println(u)
	if err != nil {
		return err
	}
	return nil
}
