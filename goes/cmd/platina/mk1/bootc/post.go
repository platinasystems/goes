// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootc

//TODO FIX SERVER COMM
//TODO ADD REGISTRATION TO REAL GOES IF ENB
//TODO FIX IP COMMAND IN GOESBOOT
//TODO START CODING GOES-BOOT ZTP, DHCP, PCC

//TODO go test
//TODO run goes-build
//TODO pull build server image
//TODO build on laptop
//TODO update ZTP document
//TODO latest images on i28

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	CLASS    string = "register"
	BSTAT    string = "bootstatus"
	RDCFG    string = "invadercfg"
	REGIS    string = "invader"
	UNREG    string = "unregister"
	TESTENB  bool   = true
	TESTIP   string = "192.168.101.142"
	TESTPORT string = "8081"
	TESTSN   string = "12345678"
)

type PCC struct {
	enb  bool
	ip   string
	port string
	sn   string
}

var pcc = PCC{
	enb:  false,
	ip:   "",
	port: "",
	sn:   "",
}

func pccInitFile() error {
	if err := readCfg(); err != nil {
		return err
	}
	Cfg.PccEnb = TESTENB
	Cfg.PccIP = TESTIP
	Cfg.PccPort = TESTPORT
	Cfg.PccSN = TESTSN
	if err := writeCfg(); err != nil {
		pcc.enb = false
		return err
	}
	return nil
}

func pccInit() error {
	if err := readCfg(); err != nil {
		pcc.enb = false
		return err
	}
	pcc.enb = Cfg.PccEnb
	pcc.ip = Cfg.PccIP
	pcc.port = Cfg.PccPort
	pcc.sn = Cfg.PccSN
	return nil
}

func doPost(cmd string, msg string) (res string, err error) {
	pccURL := "http://" + pcc.ip + ":" + pcc.port + "/" + CLASS + "/" + cmd + "/" + pcc.sn
	if msg == "" {
		msg = "/" + CLASS + "/" + cmd + "/" + pcc.sn
	}

	v := url.Values{}
	v.Set("msg", msg)
	s := v.Encode()
	fmt.Printf("v.Encode(): %v\n", s)

	req, err := http.NewRequest("POST", pccURL, strings.NewReader(s))
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return "", err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return "", err
	}
	res = string(data)
	return
}
