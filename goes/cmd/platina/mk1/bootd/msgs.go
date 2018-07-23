// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootc"
	"github.com/platinasystems/go/goes/cmd/platina/mk1/upgrade"
	"github.com/platinasystems/go/internal/kexec"
)

func wipe() (s string, err error) {
	if err := bootc.Wipe(); err != nil {
		return "", err
	}
	reboot()
	return "", nil
}

func rebootlinux() (s string, err error) {
	var i = 0
	for start := time.Now(); time.Since(start) < 10*time.Second; {
		i++
	}
	reboot()
	return "", nil
}

func getConfig() (s string, err error) {
	return "", nil
}

func putConfig() (s string, err error) {
	return "", nil
}

func getClientData(j int) (s string, err error) {
	readClientCfgDB()
	i := "01:02:03:04:05:06"
	ip := getIpAddr()
	bootc.ClientCfg[i].IpAddr = ip
	bootc.ClientCfg[i].Name = "i" + bootc.ClientCfg[i].IpAddr + "sjc01"
	vg, vk, vc, err := upgrade.GetVersionInfo()
	if err != nil {
		return "", nil
	}
	bootc.ClientCfg[i].GoesVersion = vg
	bootc.ClientCfg[i].KernelVersion = vk
	bootc.ClientCfg[i].GoesBootVersion = vc

	jsonInfo, err := json.Marshal(bootc.ClientCfg[i])
	if err != nil {
		return "", err
	}
	return string(jsonInfo), nil
}

func getClientBootData(j int) (s string, err error) {
	readClientCfgDB()
	for i, _ := range bootc.ClientBootCfg {
		if bootc.ClientCfg[i].Unit == j {
			jsonInfo, err := json.Marshal(bootc.ClientBootCfg[i])
			if err != nil {
				return "", err
			}
			return string(jsonInfo), nil
		}
	}
	err = fmt.Errorf("client number not found: %v", err)
	return "", nil
}

func putClientInfo() (s string, err error) {
	return "", nil
}

func putIso() (s string, err error) {
	return "", nil
}

func putPreseed() (s string, err error) {
	return "", nil
}

func putTar() (s string, err error) {
	return "", nil
}

func putRcLocal() (s string, err error) {
	return "", nil
}

func reBoot() (s string, err error) {
	return "", nil
}

func readClientCfgDB() (err error) {
	// TODO try reading from cloud DB

	// TODO try reading from local DB

	// default to literal for testing
	bootc.ClientCfg["01:02:03:04:05:06"] = &bootc.Client{
		Unit:           1,
		Name:           "Invader10",
		Machine:        "PS-3001",
		MacAddr:        "01:02:03:04:05:06",
		IpAddr:         "0.0.0.0",
		BootState:      bootc.BootStateNotRegistered,
		InstallState:   bootc.InstallStateFactory,
		AutoInstall:    true,
		CertPresent:    false,
		DistroType:     bootc.Debian,
		TimeRegistered: "0000-00-00:00:00:00",
		TimeInstalled:  "0000-00-00:00:00:00",
		InstallCounter: 0,
	}
	bootc.ClientCfg["01:02:03:04:05:07"] = &bootc.Client{
		Unit:           2,
		Name:           "Invader11",
		Machine:        "PS-3001",
		MacAddr:        "01:02:03:04:05:07",
		IpAddr:         "0.0.0.0",
		BootState:      bootc.BootStateNotRegistered,
		InstallState:   bootc.InstallStateFactory,
		AutoInstall:    true,
		CertPresent:    false,
		DistroType:     bootc.Debian,
		TimeRegistered: "0000-00-00:00:00:00",
		TimeInstalled:  "0000-00-00:00:00:00",
		InstallCounter: 0,
	}
	bootc.ClientCfg["01:02:03:04:05:08"] = &bootc.Client{
		Unit:           3,
		Name:           "Invader12",
		Machine:        "PS-3001",
		MacAddr:        "01:02:03:04:05:08",
		IpAddr:         "0.0.0.0",
		BootState:      bootc.BootStateNotRegistered,
		InstallState:   bootc.InstallStateFactory,
		AutoInstall:    true,
		CertPresent:    false,
		DistroType:     bootc.Debian,
		TimeRegistered: "0000-00-00:00:00:00",
		TimeInstalled:  "0000-00-00:00:00:00",
		InstallCounter: 0,
	}
	bootc.ClientBootCfg["01:02:03:04:05:06"] = &bootc.BootcConfig{
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
		ISO2Name:        " ",
		ISO2Desc:        " ",
		ISOlastUsed:     1,
	}
	bootc.ClientBootCfg["01:02:03:04:05:07"] = &bootc.BootcConfig{
		Install:         false,
		BootSda1:        false,
		BootSda6Cnt:     3,
		EraseSda6:       false,
		IAmMaster:       false,
		MyIpAddr:        "192.168.101.130",
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
	}
	bootc.ClientBootCfg["01:02:03:04:05:07"] = &bootc.BootcConfig{
		Install:         false,
		BootSda1:        false,
		BootSda6Cnt:     3,
		EraseSda6:       false,
		IAmMaster:       false,
		MyIpAddr:        "192.168.101.131",
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
	}

	// temp
	mac := "01:02:03:04:05:06"
	bootc.ClientCfg[mac].BootState = bootc.BootStateRegistered
	bootc.ClientCfg[mac].InstallState = bootc.InstallStateInProgress
	t := time.Now()
	bootc.ClientCfg[mac].TimeRegistered = fmt.Sprintf("%10s",
		t.Format("2006-01-02 15:04:05"))
	bootc.ClientCfg[mac].TimeInstalled = fmt.Sprintf("%10s",
		t.Format("2006-01-02 15:04:05"))
	bootc.ClientCfg[mac].InstallCounter++

	return nil
}

func getIpAddr() string {
	d1 := []byte("#!/bin/bash\necho -e " + `| /sbin/ip addr show dev eth0 primary`)
	if err := ioutil.WriteFile("/tmp/tmp1", d1, 0755); err != nil {
		return "0.0.0.0"
	}
	cmd := exec.Command("/tmp/tmp1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "0.0.0.0"
	}
	outs := strings.Split(string(out), "\n")
	ip := "0.0.0.0"
	for _, m := range outs {
		if strings.Contains(m, "scope global eth0") {
			mm := strings.Split(m, " ")
			for n, mmm := range mm {
				if mmm == "inet" {
					mmmm := strings.Split(mm[n+1], "/")
					ip = mmmm[0]
				}
			}
		}
	}
	return ip
}

func reboot() error {
	kexec.Prepare()

	_ = syscall.Reboot(syscall.LINUX_REBOOT_CMD_KEXEC)

	return syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
}
