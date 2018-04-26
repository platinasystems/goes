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

	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootd"
)

func register(mip string, mac string, ip string) (r int, n string, err error) {
	regReq.Mac = mac
	regReq.IP = ip
	jsonInfo, _ := json.Marshal(regReq)
	s := ""

	if s, err = sendReq(mip, bootd.Register+" "+string(jsonInfo)); err != nil {
		return bootd.BootStateNotRegistered, "", fmt.Errorf("Error contacting Master")
	}

	fmt.Println("JSON:", s)
	err = json.Unmarshal([]byte(s), &regReq)
	if err != nil {
		fmt.Println("There was an error:", err)
	}
	reply := regReply.Reply
	name := regReply.TorName
	err = regReply.Error

	return reply, name, err
}

func executeScript() error {
	return nil
}

func dashboard(mip string) (err error) {
	s := ""
	if s, err = sendReq(mip, "dashboard"); err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func numclients(mip string) (err error) {
	s := ""
	if s, err = sendReq(mip, "numclients"); err != nil {
		return err
	}
	fmt.Println(s)
	return nil
}

func clientdata(mip string, unit int) (err error) {
	s := ""
	if s, err = sendReq(mip, "clientdata "+strconv.Itoa(unit)); err != nil {
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
	return "192.168.101.142" // TODO call function
}

func getMAC() string {
	return "01:02:03:04:05:06" // TODO call function
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
