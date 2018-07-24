// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

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

// FIXME VERIFY REGISTER COMMAND
// FIXME add support for 2 ISOs
// FIXME add config from server
// FIXME add status updates msgs to server
// FIXME CONFIG PORT NUMBER, remove hardcodes, auto update bootc.cfg if out of date, for new fields
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
	"time"

	"github.com/platinasystems/go/internal/machine"
)

const (
	verNum       = "1.06"
	goesBootCfg  = "/mountd/sda1/bootc.cfg"
	sda1Cfg      = "/bootc.cfg"
	sda6Cfg      = "/mnt/bootc.cfg"
	goesBootPath = "/mountd/sda1/"
	sda1Path     = "/"
	sda6Path     = "/mnt/"
	mount        = "/mnt"
	devSda1      = "/dev/sda1"
	devSda6      = "/dev/sda6"
	tmpFile      = "/tmp/EEOF"
	mntEtc       = "/mnt/etc"
	fstype       = "ext4"
	zero         = uintptr(0)
	sda1         = "sda1"
	sda6         = "sda6"
	goesBoot     = "coreboot"
	cbSda1       = "/mountd/sda1/"
	cbSda6       = "/mountd/sda6/"
	tarFile      = "postinstall.tar.gz"
	scriptFile   = "rc.local"
	waitMs       = 100
	timeoutCnt   = 50
)

var BootcCfgFile string
var uuid1 string
var uuid6 string
var kexec0 string
var kexec1 string
var kexec6 string

var fileList = [...]string{
	"debian-8.10.0-amd64-DVD-1.iso",
	"preseed.cfg",
	"hd-media/preseed.cfg",
	"boot/vmlinuz",
	"boot/initrd.gz",
	"usr/local/sbin/flashrom",
	"usr/local/share/flashrom/layouts",
	"rc.local",
	"postinstall.tar.gz",
}

func Bootc() []string {
	for i := 0; ; i++ {
		time.Sleep(time.Millisecond * time.Duration(waitMs))
		if i > timeoutCnt {
			return []string{""}
		}
		p, err := ioutil.ReadFile("/proc/partitions")
		if err != nil {
			continue
		}
		parts := string(p)
		m, err := ioutil.ReadFile("/proc/mounts")
		if err != nil {
			continue
		}
		mounts := string(m)
		if err := readCfg(); err != nil {
			continue
		}
		if Cfg.Disable {
			return []string{""}
		}

		if err := fixNewroot(); err != nil {
			fmt.Println("Error: can't fix newroot, drop into grub...")
			return []string{""}
		}

		// sda1 utility mode
		if Cfg.BootSda1 && strings.Contains(mounts, "sda1") {
			if err := formKexec1(); err != nil {
				fmt.Println("Error: can't form kexec string, drop into grub...")
				return []string{""}
			}
			if err := clrSda1Flag(); err != nil {
				fmt.Println("Error: can't clear sda1 flag, drop into grub...")
				return []string{""}
			}
			// boot sda1 for recovery
			return []string{"kexec", "-k", Cfg.Sda1K,
				"-i", Cfg.Sda1I, "-c", kexec1, "-e"}
		}

		// install
		if Cfg.Install && strings.Contains(mounts, "sda1") && !strings.Contains(parts, "sda6") {
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
			// boot installer from sda1
			return []string{"kexec", "-k", Cfg.ReInstallK, "-i",
				Cfg.ReInstallI, "-c", kexec0, "-e"}
		}

		// postinstall
		if Cfg.PostInstall && strings.Contains(mounts, sda6) && strings.Contains(mounts, sda1) {
			if err := clrPostInstall(); err != nil {
				fmt.Println("Error: post install copy failed, drop into grub...")
				return []string{""}
			}
			// FIXME update this per new architecture
			if err := Copy(cbSda1+tarFile, cbSda6+tarFile); err != nil {
				fmt.Println("Error: post install copy failed, drop into grub...")
				return []string{""}
			}
			if err := Copy(cbSda1+scriptFile, cbSda6+"etc/"+scriptFile); err != nil {
				fmt.Println("Error: post install copy failed, drop into grub...")
				return []string{""}
			}
		}

		// DHCP, ZTP, PCC
		if err = dhcpZtpPcc(); err != nil {
			return []string{""}
		}

		// sda6 normal
		if Cfg.BootSda6Cnt > 0 && strings.Contains(parts, sda6) && strings.Contains(mounts, sda6) {
			if err := fixPaths(); err != nil {
				fmt.Println("Error: can't fix paths, drop into grub...")
				return []string{""}
			}
			if err := formKexec6(); err != nil {
				fmt.Println("Error: can't form sda6 kexec, drop into grub...")
				return []string{""}
			}
			if err := decBootSda6Cnt(); err != nil {
				fmt.Println("Error: can't decrement sda6cnt, drop into grub...")
				return []string{""}
			}
			// boot sda6
			return []string{"kexec", "-k", Cfg.Sda6K,
				"-i", Cfg.Sda6I, "-c", kexec6, "-e"}
		}

		// non-partitioned
		if !strings.Contains(parts, sda6) && strings.Contains(mounts, sda1) {
			if err := formKexec1(); err != nil {
				fmt.Println("Error: can't form kexec string, drop into grub...")
				return []string{""}
			}
			if err := clrSda1Flag(); err != nil {
				fmt.Println("Error: can't clear sda1 flag, drop into grub...")
				return []string{""}
			}
			// boot sda1 if we are not partitioned
			return []string{"kexec", "-k", Cfg.Sda1K,
				"-i", Cfg.Sda1I, "-c", kexec1, "-e"}
		}
	}

	fmt.Println("Error: bootc can't boot, drop into grub...")
	return []string{""}
}

func dhcpZtpPcc() error {
	// DHCP, ZTP, PCC
	//
	//===========
	// SETUP PHASE - set dhcp bits, static IP
	//
	// D10 3 SEC LOOP wait for console interrupt --> YES, E10 goes-boot shell
	//
	// D20 static IP configured? (bootc.cfg)     --> NO,  goto D60 check if PCC enabled
	//                                                    skip ZTP & DHCP
	//                                                    P10 set DHCP option 55 and option 61
	//
	// D30a Is ZTP Enabled?                      --> YES, P20 set Option 43 (ZTP script)
	//
	//===========
	// DHCP PHASE
	//
	// D40 INFINITE LOOP DHCP server found?      --> NO, check for ESC --> E20a shell
	//
	// ... DO DHCP
	//
	//===========
	// ZTP PHASE
	//
	// D30b Is ZTP Enabled?                      --> NO, goto D60, PCC Enb check
	//
	// D50 INFINITE LOOP ZTP server found?       --> NO, check for ESC --> E20b shell
	//
	// P30 update bootc.cfg, images, run script
	//
	//===========
	// PCC PHASE
	//
	// D60 Is PCC Enabled                        --> NO, goto P50 boot sda6
	//
	// D70
	//
	// P40
	//
	// P50

	return nil
}

func checkFiles() error {
	context, err := getContext()
	if err != nil {
		return err
	}
	path := ""
	switch context {
	case goesBoot:
		fmt.Println("TEMP: goes-boot context.")
		path = goesBootPath
	case sda1:
		fmt.Println("TEMP: goes sda1 context.")
		path = sda1Path
	case sda6:
		fmt.Println("TEMP: goes sda6 context.")
		path = sda6Path
		if err := mountSda1(); err != nil {
			fmt.Println("TEMP: goes sda6 mount-sda1 fail.")
			return err
		}
		fmt.Println("TEMP: goes sda6 mount-sda1 success.")
	default:
		fmt.Println("ERROR: could not determine context.")
		return nil
	}

	good := true
	//FIXME change names of .iso's to original.iso, latest.iso
	for _, f := range fileList {
		fmt.Println("Checking", path+f, "exists...")
		if _, err := os.Stat(path + f); os.IsNotExist(err) {
			fmt.Println("ERROR:	file", path+f, "does not exist")
			good = false
		}
	}
	// 1. check for goes running
	// 2. check for sda6cnt == 3
	// 3. check the sda6 k string, see it matches /boot
	// 4. check the sda6 i string, see it matches /boot
	if good {
		fmt.Println("PASSED: wipe/reinstall is configured properly.")
	} else {
		fmt.Println("FAILED: wipe/reinstall is NOT configured properly.")
	}
	return nil
}

func getFiles() error {
	context, err := getContext()
	if err != nil {
		return err
	}
	path := ""
	switch context {
	case goesBoot:
		fmt.Println("TEMP: goes-boot context.")
		path = goesBootPath
	case sda1:
		fmt.Println("TEMP: goes sda1 context.")
		path = sda1Path
	case sda6:
		fmt.Println("TEMP: goes sda6 context.")
		path = sda6Path
		if err := mountSda1(); err != nil {
			fmt.Println("TEMP: goes sda6 mount-sda1 fail.")
			return err
		}
		fmt.Println("TEMP: goes sda6 mount-sda1 success.")
	default:
		fmt.Println("ERROR: could not determine context.")
		return nil
	}

	fmt.Println("TEMP: sda1 path = ", path)
	return nil
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
		Disable:         false,
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

func mountSda1() error {
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
	return nil
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
		if err := mountSda1(); err != nil {
			return err
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

func setDisable() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Disable = true
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

func clrDisable() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Disable = false
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

func SetSda1K(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Sda1K = x
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

func SetSda1I(x string) error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.Sda1I = x
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

func SetSda6K(x string) error {
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

func SetSda6I(x string) error {
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
	err := SetSda6K(k)
	if err != nil {
		return err
	}
	i = cbSda6 + "boot/" + i
	err = SetSda6I(i)
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

func fixPaths() error { //FIXME Temporary remove by 7/31/2018
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
	Cfg.Sda6K = strings.Replace(Cfg.Sda6K, "newroot", "mountd", 1)
	Cfg.Sda6I = strings.Replace(Cfg.Sda6I, "newroot", "mountd", 1)
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
