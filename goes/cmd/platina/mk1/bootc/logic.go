// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// ZTP Modules
//
// bootc.go Module - GOES-BOOT and GOES
//   provides Client REST messages to server
//   provides state machine for executing wipe/re-install in goesboot
//   executes Register(), receives Client struct, script
//
// bootd.go Daemon - GOES
//   provides Server REST messages to respond to master
//
// pushd.go Daemon - GOES
//   executes wipe request from bootd
//   pushes boot state, install state, etc. to master
//
// clntd.go Daemon - GOES-BOOT
//   launches Bootc() in goes-boot
//   hopefully will beat grub launch daemon
//
// script run infrastructure TBD
// postInstall0 TBD
// postIntalll1 TBD

// FIXME FIX WIPE from WEB
// FIXME REVERIFY REGISTER COMMAND
// FIXME add check for sda6 before updating strings, add check for sda6 before trying to boot sda6 in goes-boot
// FIXME add support for 2 ISOs
// FIXME add config from server
// FIXME add status updates msgs to server
// FIXME CONFIG PORT NUMBER, remove hardcodes, auto update bootc.cfg if out of date, for new fields
// FIXME LAUNCH SCRIPT AFTER EVERY BOOT - special case right after wipe, script is custom
// FIXME BOOTD check works with real master
// FIXME COMMIT NEW SERVER SIDE CODE
// FIXME CONTACT SERVER w/o IP address
// FIXME build standalone

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
	goesBootCfg = "/mountd/sda1/bootc.cfg"
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
	goesBoot    = "goesboot"
	cbSda1      = "/mountd/sda1/"
	cbSda6      = "/mountd/sda6/"
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
	if err := initCfg(); err != nil {
		fmt.Println("Error: bootc.cfg error, drop into grub...")
		return []string{""}
	}
	if err := readCfg(); err != nil {
		fmt.Println("Error: can't read bootc.cfg, drop into grub...")
		return []string{""}
	}
	if err := fixPaths(); err != nil {
		fmt.Println("Error: can't fix paths, drop into grub...")
		return []string{""}
	}

	if !serverAvail() {
		fmt.Println("Info: server is not available, using local bootc.cfg")
	}

	fmt.Printf("Info: Install = %v, BootSda1 = %v, BootSda6Cnt = %v\n",
		Cfg.Install, Cfg.BootSda1, Cfg.BootSda6Cnt)
	if !Cfg.Install && Cfg.BootSda6Cnt == 0 && !Cfg.BootSda1 {
		fmt.Println("Error: bootc has nothing to boot, drop into grub...")
		return []string{""}
	}

	if Cfg.BootSda1 {
		if err := formKexec1(); err != nil {
			fmt.Println("Error: can't form kexec string, drop into grub...")
			return []string{""}
		}
		if err := clrSda1Flag(); err != nil {
			fmt.Println("Error: can't clear sda1 flag, drop into grub...")
			return []string{""}
		}
		return []string{"kexec", "-k", Cfg.Sda1K,
			"-i", Cfg.Sda1I, "-c", kexec1, "-e"}
	}

	if Cfg.Install {
		if err := formKexec1(); err != nil {
			fmt.Println("Error: can't form install kexec, drop into grub...")
			return []string{""}
		}
		if err := clrInstall(); err != nil {
			fmt.Println("Error: can't clear install bit, drop into grub...")
			return []string{""}
		}
		if err := setPostInstall(); err != nil {
			fmt.Println("Error: can't set postinstall bit, drop into grub...")
			return []string{""}
		}
		return []string{"kexec", "-k", Cfg.ReInstallK, "-i",
			Cfg.ReInstallI, "-c", kexec0, "-e"}
	}

	if Cfg.PostInstall {
		if err := clrPostInstall(); err != nil {
			fmt.Println("Error: post install copy failed 1, drop into grub...")
			return []string{""}
		}
		if err := Copy(cbSda1+tarFile, cbSda6+tarFile); err != nil {
			fmt.Println("Error: post install copy failed 2, drop into grub...")
			return []string{""}
		}
		if err := Copy(cbSda1+scriptFile, cbSda6+"etc/"+scriptFile); err != nil {
			fmt.Println("Error: post install copy failed 3, drop into grub...")
			return []string{""}
		}
	}

	if Cfg.BootSda6Cnt > 0 {
		if err := formKexec6(); err != nil {
			fmt.Println("Error: can't form sda6 kexec, drop into grub...")
			return []string{""}
		}
		if err := decBootSda6Cnt(); err != nil {
			fmt.Println("Error: can't decrement sda6cnt, drop into grub...")
			return []string{""}
		}
		return []string{"kexec", "-k", Cfg.Sda6K,
			"-i", Cfg.Sda6I, "-c", kexec6, "-e"}
	}
	fmt.Println("Error: bootc can't boot, drop into grub...")
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
		ReInstallK:      "/mountd/sda1/boot/vmlinuz",
		ReInstallI:      "/mountd/sda1/boot/initrd.gz",
		ReInstallC:      `netcfg/get_hostname=platina netcfg/get_domain=platinasystems.com interface=auto auto locale=en_US preseed/file=/hd-media/preseed.cfg`,
		Sda1K:           "/mountd/sda1/boot/vmlinuz-3.16.0-4-amd64",
		Sda1I:           "/mountd/sda1/boot/initrd.img-3.16.0-4-amd64",
		Sda1C:           "::eth0:none",
		Sda6K:           "/mountd/sda6/boot/vmlinuz-3.16.0-4-amd64",
		Sda6I:           "/mountd/sda6/boot/initrd.img-3.16.0-4-amd64",
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
	if mach == goesBoot {
		return goesBoot, nil
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
	case goesBoot:
		BootcCfgFile = goesBootCfg
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
	kexec6 = "root=UUID=" + uuid6 + " console=ttyS0,115200"
	kexec6 += Cfg.Sda6C
	return nil
}

func readUUID(partition string) (uuid string, err error) {
	dat, err := ioutil.ReadFile("/mountd/" + partition + "/etc/fstab")
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

func ResetSda6Count() error {
	c, err := getContext()
	if err != nil {
		return err
	}
	if c != sda6 {
		return nil
	}
	if err = readCfg(); err != nil {
		return err
	}
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

func UpdateBootcCfg(k, i string) error {
	k = cbSda6 + "boot/" + k
	err := setKernel(k)
	if err != nil {
		return err
	}
	i = cbSda6 + "boot/" + i
	err = setInitrd(i)
	if err != nil {
		return err
	}
	return nil
}

func Wipe() error {
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
	return nil
}

func runScript(name string) (err error) {
	return nil
}

func fixPaths() error {
	if err := fixNewroot(); err != nil { // FIXME Temporary remove by 7/31/2018
		fmt.Println("Error: can't fix newroot, drop into grub...")
		return err
	}

	files, err := ioutil.ReadDir(cbSda6 + "boot")
	if err != nil {
		return err
	}
	k := ""
	kn := ""
	for _, f := range files {
		if strings.Contains(f.Name(), "vmlinuz") {
			fn := strings.Replace(f.Name(), ".", "", 1)
			if fn > kn {
				k = f.Name()
				kn = fn
			}
		}
	}
	i := strings.Replace(k, "vmlinuz", "initrd.img", 1)
	err = UpdateBootcCfg(k, i)
	if err != nil {
		return err
	}
	return nil
}

func fixNewroot() error { // FIXME Temporary remove by 7/31/2018
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.ReInstallK = strings.Replace(Cfg.ReInstallK, "newroot", "mountd", 1)
	Cfg.ReInstallI = strings.Replace(Cfg.ReInstallI, "newroot", "mountd", 1)
	Cfg.Sda1K = strings.Replace(Cfg.Sda1K, "newroot", "mountd", 1)
	Cfg.Sda1I = strings.Replace(Cfg.Sda1I, "newroot", "mountd", 1)
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
