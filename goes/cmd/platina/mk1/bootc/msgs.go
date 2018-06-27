// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

func register(mip string, mac string, ip string) (r int, n string, err error) {
	regReq.Mac = mac
	regReq.IP = ip
	jsonInfo, _ := json.Marshal(regReq)
	s := ""

	if s, err = sendReq(mip, Register+" "+string(jsonInfo)); err != nil {
		return BootStateNotRegistered, "", fmt.Errorf("Error contacting Master")
	}

	err = json.Unmarshal([]byte(s), &regReply)
	if err != nil {
		fmt.Println("There was an error:", err)
	}
	reply := regReply.Reply
	name := regReply.TorName
	err = regReply.Error

	return reply, name, err
}

func getnumclients(mip string) (err error) {
	s := ""
	if s, err = sendReq(mip, NumClients); err != nil {
		return err
	}

	err = json.Unmarshal([]byte(s), &numReply)
	if err != nil {
		fmt.Println("There was an error:", err)
	}
	num := numReply.Clients
	err = numReply.Error
	fmt.Println(num, err)

	return err
}

func getclientdata(mip string, unit int) (err error) {
	s := ""
	if s, err = sendReq(mip, ClientData+" "+strconv.Itoa(unit)); err != nil {
		return err
	}

	err = json.Unmarshal([]byte(s), &dataReply)
	if err != nil {
		fmt.Println("There was an error:", err)
	}
	dat := dataReply.Client
	err = dataReply.Error
	fmt.Println(dat, err)

	return err
}
func getclientdata2(mip string, unit int) (err error) {
	return nil
}
func getclientdata3(mip string, unit int) (err error) {
	return nil
}

func getclientbootdata(mip string, unit int) (err error) {
	s := ""
	if s, err = sendReq(mip, ClientBootData+" "+strconv.Itoa(unit)); err != nil {
		return err
	}

	err = json.Unmarshal([]byte(s), &dataReply)
	if err != nil {
		fmt.Println("There was an error:", err)
	}
	dat := dataReply.Client
	err = dataReply.Error
	fmt.Println(dat, err)

	return err
}

func getscript(mip string, name string) (err error) {
	s := ""
	if s, err = sendReq(mip, Script+" "+name); err != nil {
		return err
	}

	err = json.Unmarshal([]byte(s), &scriptReply)
	if err != nil {
		fmt.Println("There was an error:", err)
	}
	script := scriptReply.Script
	err = scriptReply.Error
	fmt.Println(script, err)

	return err
}

func getbinary(mip string, name string) (err error) {
	s := ""
	if s, err = sendReq(mip, Binary+" "+name); err != nil {
		return err
	}

	err = json.Unmarshal([]byte(s), &binaryReply)
	if err != nil {
		fmt.Println("There was an error:", err)
	}
	bin := binaryReply.Binary
	err = numReply.Error
	fmt.Println(bin, err)

	return err
}

func dashboard(mip string) (err error) {
	s := ""
	if s, err = sendReq(mip, "dashboard"); err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func dumpvars(mip string) (err error) {
	s := ""
	if s, err = sendReq(mip, "dumpvars"); err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func test404(mip string) (err error) {
	s := ""
	if s, err = sendReq(mip, "xxx"); err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func sendReq(mip string, s string) (res string, err error) {
	resp, err := http.Get("http://" + mip + ":9090/" + s)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getMasterIP() string {
	// TODO cycle through DNS, .1, URL, hardcode list
	return "192.168.101.142"
}

func getIP() string {
	return "192.168.101.142" // TODO call getIP function
}

func getMAC() string {
	return "01:02:03:04:05:06" // TODO call getMAC function
}

func getIP2() string {
	return "192.168.101.143"
}

func getIP3() string {
	return "192.168.101.144"
}

func getMAC2() string {
	return "01:02:03:04:05:07"
}

func getMAC3() string {
	return "01:02:03:04:05:08"
}

/*
func dumpJson() {
	ClientCfg = Client{
		Unit:            1,
		Name:            "Invader10",
		Machine:         "ToR MK1",
		MacAddr:         "01:02:03:04:05:06",
		IpAddr:          "198.168.101.129",
		IpGWay:          "192.168.101.1",
		IpMask:          "255.255.255.0",
		BootState:       BootStateNotRegistered,
		InstallState:    InstallStateFactory,
		AutoInstall:     true,
		CertPresent:     false,
		DistroType:      Debian,
		TimeRegistered:  "0000-00-00:00:00:00",
		TimeInstalled:   "0000-00-00:00:00:00",
		InstallCounter:  0,
		LastISOname:     "debian-8.10.0-amd64-DVD-1.iso",
		LastISOdesc:     "Jessie debian-8.10.0",
		GoesVersion:     "v0.41-49-gdd4de12",
		KernelVersion:   "4.13.0-platina-mk1-amd64",
		GoesBootVersion: "v0.4-2-ge517297-dirty",
	}
	BootcCfg = BootcConfig{
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
		PostInstall:     false,
	}
	jsonClient, _ := json.Marshal(ClientCfg)
	fmt.Println(string(jsonClient))
	jsonBoot, _ := json.Marshal(BootcCfg)
	fmt.Println(string(jsonBoot))
}
*/
