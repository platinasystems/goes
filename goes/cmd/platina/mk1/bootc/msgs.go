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
	if s, err = sendReq(mip, bootd.NumClients); err != nil {
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
	if s, err = sendReq(mip, bootd.ClientData+" "+strconv.Itoa(unit)); err != nil {
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
	if s, err = sendReq(mip, bootd.ClientBootData+" "+strconv.Itoa(unit)); err != nil {
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
	if s, err = sendReq(mip, bootd.Script+" "+name); err != nil {
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
	if s, err = sendReq(mip, bootd.Binary+" "+name); err != nil {
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
