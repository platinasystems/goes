// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'boot' requestor, this will be run in parallel on muliple client devices
// this will be run automatically from the kernel+initrd(goes) image
// kernel + initrd will be loaded by Coreboot or PXE boot

// DISCLAIMER
// this is a work in progress, this will change significantly before release
// this package must be manually added to the mk1 goes.go to be included ATM

// TO DO LIST
//TODO get our ip and mac address, replace with getting master server address
//TODO add test case of 100 units simultaneously asking for updates
//TODO define normal boot up message exchange sequence
//TODO define priming boot up message exchange sequence

// LIST OF KEY PIECES
// piece (a) kernel+initrd+boot(/init) blob, (w/utilities)
// piece (b) x509 cert
// piece (c) boot(/init) to contact server and run script
// piece (d) boot-controller on master tor
// piece (e) debian isos (etc.)
// piece (f) preseed file to answer debian install questions
// piece (g) database holding configurations of each unit indexed by mac/cert
// piece (h) modify debian installer to install Coreboot (ToR only?)
// ...
// piece (i) ansible et al
// piece (j) cloud based dashboard

// HTTP MESSAGE FORMATS
// msg format: command, client mac (xlat to numb), data
// format:id(mac),cmd,data  cmd=progress,data=x | return:name
// cmd register with master, "i just booted"
// cmd "what is my next step, i am on step x"

// MASTER SHOULD
// grab database from local disk or cloud
// wait for request for install instructions - and respond with blob
// run the 'install state machine' on per unit basis

// CLIENT SHOULD
// pxe boot
// k&i install one-of-us borgify (or--always pxe boot k&i)
// k&i install/partition
// deb install w/preseed

package bootc

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/log"
)

/////
/////
/////

// debug infrastructure
///*
type Command struct{}

func (Command) String() string { return "bootc" }

func (Command) Usage() string { return "bootc" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "boot client hook to communicate with tor master",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
description
	the bootc command is for debugging bootc client.`,
	}
}

func (Command) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("args: missing")
	}

	c, err := strconv.ParseUint(args[0], 10, 32)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}
	switch c {
	case 1: //simulate auto-boot case fixme
		sendreq(1, args[1])
	case 2: //manual case fixme
		sendreq(1, args[1])
	default:
		fmt.Println("no command...")
		log.Print("no command")
	}

	return nil
}

//*/

/////
/////
/////

func sendreq(c int, s string) {
	if c == 1 {
		resp, err := http.Get("http://192.168.101.142:9090/" + s)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
	}

	if 1 == 2 { //post case
		resp, err := http.PostForm("http://duckduckgo.com",
			url.Values{"q": {"github"}})
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("post:\n", minLines(string(body), 3))
	}
}

func minLines(s string, n int) string {
	result := strings.Join(strings.Split(s, "\n")[:n], "\n")
	return strings.Replace(result, "\r", "", -1)
}
