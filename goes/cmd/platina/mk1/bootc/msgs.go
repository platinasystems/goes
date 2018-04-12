// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func register(mip string, mac string, ip string) (s string, err error) {

	regInfo.Mac = mac
	regInfo.IP = ip
	jsonInfo, _ := json.Marshal(regInfo)

	// TODO [1] READ the /boot directory into slice

	// TODO [0] REPLACE register WITH CONSTANT
	if s, err = sendReq(mip, "register "+string(jsonInfo)); err != nil {
		return "", fmt.Errorf("Error contacting Master")
	}

	// TODO [2] UNMARSHALL JSON BLOB, name, script, err - return these

	return s, nil
}

func execute() error { // TODO
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

func dumpVars(mip string) (err error) {
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
	return "192.168.101.142" // FIXME
}

func getMAC() string {
	return "01:02:03:04:05:06" // FIXME
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
