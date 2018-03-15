// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// 'boot' requestor, this will be run in parallel on muliple client devices
// this will be run automatically from the kernel+initrd(goes) image
// kernel + initrd will be loaded by Coreboot or PXE boot
//
// for testing a CLI command will be added to goes temporarily to call this
//
// this is a work in progress, this will change significantly before release
// this package must be manually added to the mk1 goes.go to be included ATM

package bootc

import (
	"net/http"
	"net/url"
)
import "io/ioutil"
import "fmt"
import "strings"

func minLines(s string, n int) string {
	result := strings.Join(strings.Split(s, "\n")[:n], "\n")
	return strings.Replace(result, "\r", "", -1)
}

//LIST OF KEY PIECES
//==================
//PIECE (A) KERNEL+INITRD+BOOT(/INIT) BLOB, (w/UTILITIES)
//PIECE (B) X509 CERT
//PIECE (C) BOOT(/INIT) TO CONTACT SERVER AND RUN SCRIPT
//PIECE (D) BOOT-CONTROLLER ON MASTER ToR
//PIECE (E) DEBIAN ISOs (etc.)
//PIECE (F) PRESEED FILE TO ANSWER DEBIAN INSTALL QUESTIONS
//PIECE (G) DATABASE HOLDING CONFIGURATIONS OF EACH UNIT INDEXED BY MAC/CERT
//...
//PIECE (H) ANSIBLE ET AL
//PIECE (I) CLOUD BASED DASHBOARD

//HTTP MESSAGE FORMATS
//====================
//format:ID(Mac),CMD,DATA  CMD=PROGRESS,DATA=x
//return:NAME

//THINGS MASTER SHOULD DO
//=======================
//GRAB DATABASE FROM LOCAL DISK OR CLOUD
//WAIT FOR REQUEST FOR INSTALL INSTRUCTIONS - AND RESPOND WITH BLOB
//  RUN INSTALL STATE MACHINE ON PER UNIT BASIS

//THINGS 3RD PARTY BOX SHOULD DO
//==============================
//PXE BOOT
//K&I INSTALL ONE-OF-US BORGIFY (OR--ALWAYS PXE BOOT K&I)
//  K&I INSTALL/PARTITION
//DEB INSTALL W/PRESEED

//NORMAL BOOTUP
//=============
//REGISTER WITH MASTER SERVER

func main() {
	resp, err := http.Get("http://172.16.2.21:9091/bbb")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))

	if 1 == 2 {
		resp, err = http.PostForm("http://duckduckgo.com",
			url.Values{"q": {"github"}})
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		body, err = ioutil.ReadAll(resp.Body)
		fmt.Println("post:\n", minLines(string(body), 3))
	}
}
