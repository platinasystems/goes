// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//FIXME refactor set/clears
//FIXME add support for 2 ISOs
//FIXME add config from server
//FIXME add status updates msgs to server

package bootc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/platinasystems/go/internal/machine"
)

const (
	corebootCfg = "/newroot/sda1/bootc.cfg"
	sda1Cfg     = "/bootc.cfg"
	sda6Cfg     = "/mnt/bootc.cfg"
	mount       = "/mnt"
	devSda1     = "/dev/sda1"
	devSda6     = "/dev/sda6"
	tmpFile     = "/tmp/EEOF"
	mntEtc      = "/mnt/etc"
	fstype      = "ext4"
	zero        = uintptr(0)
	sda1        = "sda1"
	sda6        = "sda6"
	coreboot    = "coreboot"
	cbSda1      = "/newroot/sda1/"
	cbSda6      = "/newroot/sda6/"
	tarFile     = "postinstall.tar.gz"
	scriptFile  = "rc.local"
)

var BootcCfgFile string
var uuid1 string
var uuid6 string
var kexec0 string
var kexec1 string
var kexec6 string

func Bootc() []string {
	if err := readCfg(); err != nil {
		return []string{""}
	}

	if !serverAvail() {
		fmt.Println("Info: server is not available, using local bootc.cfg")
	}

	fmt.Printf("Info: Install = %v, BootSda1 = %v, BootSda6Cnt = %v\n",
		Cfg.Install, Cfg.BootSda1, Cfg.BootSda6Cnt)
	if !Cfg.Install && Cfg.BootSda6Cnt == 0 && !Cfg.BootSda1 {
		return []string{""}
	}

	if Cfg.BootSda1 {
		if err := formKexec1(); err != nil {
			return []string{""}
		}
		if err := clrSda1Flag(); err != nil {
			return []string{""}
		}
		return []string{"kexec", "-k", Cfg.Sda1K,
			"-i", Cfg.Sda1I, "-c", kexec1, "-e"}
	}

	if Cfg.Install {
		if err := formKexec1(); err != nil {
			return []string{""}
		}
		if err := clrInstall(); err != nil {
			return []string{""}
		}
		if err := setPostInstall(); err != nil {
			return []string{""}
		}
		return []string{"kexec", "-k", Cfg.ReInstallK, "-i",
			Cfg.ReInstallI, "-c", kexec0, "-e"}
	}

	if Cfg.PostInstall {
		if err := clrPostInstall(); err != nil {
			return []string{""}
		}
		if err := Copy(cbSda1+tarFile, cbSda6+tarFile); err != nil {
			return []string{""}
		}
		if err := Copy(cbSda6+"etc/"+scriptFile, cbSda6+"etc/"+scriptFile+"SAVE"); err != nil {
			return []string{""}
		}
		if err := Copy(cbSda1+scriptFile, cbSda6+"etc/"+scriptFile); err != nil {
			return []string{""}
		}
	}

	if Cfg.BootSda6Cnt > 0 {
		if err := formKexec6(); err != nil {
			return []string{""}
		}
		if err := decBootSda6Cnt(); err != nil {
			return []string{""}
		}
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
	if err != nil || reply != RegReplyRegistered {
		reply, _, err = register(mip, mac, ip)
		if err != nil || reply != RegReplyRegistered {
			return false
		}
	}
	return true
}

func initCfg() error {
	Cfg = BootcConfig{
		Install:         false,
		BootSda1:        false,
		BootSda6Cnt:     3,
		EraseSda6:       false,
		IAmMaster:       false,
		MyIpAddr:        "192.168.101.129",
		MyGateway:       "192.168.101.1",
		MyNetmask:       "255.255.255.0",
		MasterAddresses: []string{"198.168.101.142"},
		ReInstallK:      "/newroot/sda1/boot/vmlinuz",
		ReInstallI:      "/newroot/sda1/boot/initrd.gz",
		ReInstallC:      `netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto locale=en_US preseed/file=/hd-media/preseed.cfg`,
		Sda1K:           "/newroot/sda1/boot/vmlinuz-3.16.0-4-amd64",
		Sda1I:           "/newroot/sda1/boot/initrd.img-3.16.0-4-amd64",
		Sda1C:           "::eth0:none",
		Sda6K:           "/newroot/sda6/boot/vmlinuz-3.16.0-4-amd64",
		Sda6I:           "/newroot/sda6/boot/initrd.img-3.16.0-4-amd64",
		Sda6C:           "::eth0:none",
		ISO1Name:        "debian-8.10.0-amd64-DVD-1.iso",
		ISO1Desc:        "Jessie debian-8.10.0",
		ISO2Name:        "",
		ISO2Desc:        "",
		ISOlastUsed:     1,
		PostInstall:     false,
	}
	if err := writeCfg(); err != nil {
		return err
	}
	return nil
}

func getContext() (context string, err error) {
	mach := machine.Name
	if mach == coreboot {
		return coreboot, nil
	}
	if mach == "platina-mk1" {
		cmd := exec.Command("df")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		outs := strings.Split(string(out), "\n")
		for _, m := range outs {
			if strings.Contains(m, devSda1) {
				return sda1, nil
			}
			if strings.Contains(m, devSda6) {
				return sda6, nil
			}
		}
	}
	return "", fmt.Errorf("Error: root directory not found")
}

func setBootcCfgFile() error {
	context, err := getContext()
	if err != nil {
		return err
	}
	switch context {
	case coreboot:
		BootcCfgFile = corebootCfg
	case sda1:
		BootcCfgFile = sda1Cfg
	case sda6:
		BootcCfgFile = sda6Cfg
		if _, err := os.Stat(mount); os.IsNotExist(err) {
			err := os.Mkdir(mount, os.FileMode(0755))
			if err != nil {
				fmt.Printf("Error mkdir: %v", err)
				return err
			}
		}
		if _, err := os.Stat(mntEtc); os.IsNotExist(err) {
			if err := syscall.Mount(devSda1, mount, fstype, zero, ""); err != nil {
				fmt.Printf("Error mounting: %v", err)
				return err
			}
		}
	default:
		return fmt.Errorf("Error: unknown machine/partition")
	}
	return nil
}

/*func mountSda6() error {
	if _, err := os.Stat(mount); os.IsNotExist(err) {
		err := os.Mkdir(mount, os.FileMode(0755))
		if err != nil {
			fmt.Printf("Error mkdir: %v", err)
			return err
		}
	}
	if _, err := os.Stat(mntEtc); os.IsNotExist(err) {
		if err := syscall.Mount(devSda6, mount, fstype, zero, ""); err != nil {
			fmt.Printf("Error mounting: %v", err)
			return err
		}
	}
	return nil
}*/

func writeCfg() error {
	if err := setBootcCfgFile(); err != nil {
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
	if err := setBootcCfgFile(); err != nil {
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

func formKexec1() (err error) {
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

func formKexec6() (err error) {
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
	Cfg.PostInstall = false
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
	Cfg.Install = false
	Cfg.PostInstall = false
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

func setPostInstall() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PostInstall = true
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

func clrPostInstall() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PostInstall = false
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

func wipe() error {
	context, err := getContext()
	if context != sda6 && context != sda1 {
		return fmt.Errorf("Not booted from sda6 or sda1, can't wipe.")
	}
	if err := clrInstall(); err != nil {
		return err
	}

	// make sure sda6 exists
	d1 := []byte("#!/bin/bash\necho -e " + `"p\nq\n"` + " | /sbin/fdisk /dev/sda\n")
	if err = ioutil.WriteFile(tmpFile, d1, 0755); err != nil {
		return err
	}
	cmd := exec.Command(tmpFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("fdisk: %v, %v\n", out, err)
		return err
	}
	outs := strings.Split(string(out), "\n")
	n := 0
	for _, m := range outs {
		if strings.Contains(m, devSda6) {
			n = 1
		}
	}
	if n == 0 {
		return fmt.Errorf("Error: /dev/sda6 not in partition table, aborting")
	}

	// delete sda6
	d1 = []byte("#!/bin/bash\necho -e " + `"d\n6\nw\n"` + " | /sbin/fdisk /dev/sda\n")
	if err = ioutil.WriteFile(tmpFile, d1, 0755); err != nil {
		return err
	}
	cmd = exec.Command(tmpFile)
	out, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("fdisk: %v, %v\n", out, err)
	}

	// make sure sda6 is gone
	d1 = []byte("#!/bin/bash\necho -e " + `"p\nq\n"` + " | /sbin/fdisk /dev/sda\n")
	if err = ioutil.WriteFile(tmpFile, d1, 0755); err != nil {
		return err
	}
	cmd = exec.Command(tmpFile)
	out, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("fdisk: %v, %v\n", out, err)
		return err
	}
	outs = strings.Split(string(out), "\n")
	for _, m := range outs {
		if strings.Contains(m, devSda6) {
			return fmt.Errorf("Error: /dev/sda6 not deleted, aborting wipe")
		}
	}

	fmt.Println("\nPlease wait...reinstalling linux on sda6\n")
	if err := setInstall(); err != nil {
		return err
	}
	reboot()
	return nil
}

func runScript(name string) (err error) {
	return nil
}

func reboot() error {
	fmt.Print("\nWILL REBOOT NOW!!!\n")
	u, err := exec.Command("shutdown", "-r", "now").Output()
	fmt.Println(u)
	if err != nil {
		return err
	}
	return nil
}

func Copy(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
