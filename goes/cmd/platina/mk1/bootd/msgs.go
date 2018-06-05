// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//TODO need to mount /sda1 on a per sda1 or sda6 basis - no coreboot

package bootd

import ()

func wipe() (s string, err error) {
	return "", nil
}

func getConfig() (s string, err error) {
	return "", nil
}

func putConfig() (s string, err error) {
	return "", nil
}

func getClientInfo() (s string, err error) {
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
	ClientCfg["01:02:03:04:05:06"] = &Client{
		Unit:           1,
		Name:           "Invader10",
		Machine:        "ToR MK1",
		MacAddr:        "01:02:03:04:05:06",
		IpAddr:         "0.0.0.0",
		BootState:      BootStateNotRegistered,
		InstallState:   InstallStateFactory,
		AutoInstall:    true,
		CertPresent:    false,
		DistroType:     Debian,
		TimeRegistered: "0000-00-00:00:00:00",
		TimeInstalled:  "0000-00-00:00:00:00",
		InstallCounter: 0,
	}
	ClientCfg["01:02:03:04:05:07"] = &Client{
		Unit:           2,
		Name:           "Invader11",
		Machine:        "ToR MK1",
		MacAddr:        "01:02:03:04:05:07",
		IpAddr:         "0.0.0.0",
		BootState:      BootStateNotRegistered,
		InstallState:   InstallStateFactory,
		AutoInstall:    true,
		CertPresent:    false,
		DistroType:     Debian,
		TimeRegistered: "0000-00-00:00:00:00",
		TimeInstalled:  "0000-00-00:00:00:00",
		InstallCounter: 0,
	}
	ClientCfg["01:02:03:04:05:08"] = &Client{
		Unit:           3,
		Name:           "Invader12",
		Machine:        "ToR MK1",
		MacAddr:        "01:02:03:04:05:08",
		IpAddr:         "0.0.0.0",
		BootState:      BootStateNotRegistered,
		InstallState:   InstallStateFactory,
		AutoInstall:    true,
		CertPresent:    false,
		DistroType:     Debian,
		TimeRegistered: "0000-00-00:00:00:00",
		TimeInstalled:  "0000-00-00:00:00:00",
		InstallCounter: 0,
	}
	ClientBootCfg["01:02:03:04:05:06"] = &BootcConfig{
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
		ISO2Name:        " ",
		ISO2Desc:        " ",
		ISOlastUsed:     1,
	}
	ClientBootCfg["01:02:03:04:05:07"] = &BootcConfig{
		Install:         false,
		BootSda1:        false,
		BootSda6Cnt:     3,
		EraseSda6:       false,
		IAmMaster:       false,
		MyIpAddr:        "192.168.101.130",
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
	}
	ClientBootCfg["01:02:03:04:05:07"] = &BootcConfig{
		Install:         false,
		BootSda1:        false,
		BootSda6Cnt:     3,
		EraseSda6:       false,
		IAmMaster:       false,
		MyIpAddr:        "192.168.101.131",
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
	}
	return nil
}
