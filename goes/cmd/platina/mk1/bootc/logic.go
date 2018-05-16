// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootd"
)

//const BootcCfgFile = "/newroot/sda1/bootc.cfg" FIXME
const BootcCfgFile = "/tmp/dat12"

var uuid1 string
var uuid6 string
var kexec1 string
var kexec6 string

func Bootc() {
	if err := readCfg(); err != nil {
		fmt.Println("1. ERROR reading bootc.cfg, run grub")
		return
	} else {
		fmt.Println("1. boot.cfg read successfully")
	}
	if serverAvail() == true {
		fmt.Println("2. Server available")
	} else {
		fmt.Println("2. Server is not available, using local bootc.cfg")
	}
	if Cfg.Grub == true {
		fmt.Println("3. GrubBit == TRUE, run grub")
		return
	} else {
		fmt.Println("3. GrubBit == FALSE, set GrubBit, run bootc")
		if err := setGrubBit(); err != nil { // avoid bootc loop
			return
		}
	}
	if err := formStrings(); err != nil {
		fmt.Println("4. ERROR forming kexec strings, run grub")
		return
	} else {
		fmt.Println("4. kexec strings formed correctly")
		fmt.Println("  KEXEC1: ", kexec1)
		fmt.Println("  KEXEC6: ", kexec6)
	}
	if Cfg.ReInstall == true {
		fmt.Println("5. Re-install == TRUE, clear ReInstallBit, run installer")
		if err := clrReInstallBit(); err != nil {
			return
		}
		doKexec(kexec1)
	} else {
		fmt.Println("5. Re-install == FALSE, boot sda6")
		doKexec(kexec6)
	}
}

func serverAvail() bool { //FIXME
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

func writeCfg() error {
	Cfg.Grub = false
	Cfg.ReInstall = false
	Cfg.IAmMaster = false
	Cfg.MyIpAddr = "192.168.101.129"
	Cfg.MyIpGWay = "192.168.101.1"
	Cfg.MyIpMask = "255.255.255.0"
	Cfg.MasterAddresses = []string{"198.168.101.142"}
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
		return err
	}
	dat, err := ioutil.ReadFile(BootcCfgFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(dat), &Cfg)
	if err != nil {
		return err
	}
	fmt.Println(Cfg)
	return nil
}

func formStrings() (err error) {
	uuid1, err = readUUID("sda1")
	if err != nil {
		return err
	}
	uuid6, err = readUUID("sda1")
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

func doKexec(k string) error {
	fmt.Println(k)
	d1 := []byte(k)
	err := ioutil.WriteFile("kexe", d1, 0644)
	if err != nil {
		fmt.Println("error writing kexe")
	}
	//FIXME FOR NOW DON"T KICK OFF --  err = c.g.Main("source", "kexe")
	//if err != nil {
	//	return err
	//}
	return nil // should never get here
}

func setGrubBit() error {
	Cfg.Grub = true
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

func clrGrubBit() error {
	Cfg.Grub = false
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

func setReInstallBit() error {
	Cfg.ReInstall = true
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

func clrReInstallBit() error {
	Cfg.ReInstall = false
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

func runScript(name string) (err error) {
	return nil
}
