// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// http server daemon to service 'boot' requests from muliple client devices
// typically this will run on a single "master" ToR device
// contains boot state machine for each client
// accesses database of configurations stored either locally or in the cloud
//
// this is a work in progress, this will change significantly before release
// this package must be manually added to the mk1 goes.go to be included ATM

package bootd

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

//TODO ADD AS A DAEMON
//TODO reply via JSON TFTP
//TODO progress bar, x units, unit names, unit mac, uint cert
//TODO tester
//TODO CONVERT TO ARRAY OR SLICE OF STRUCTS
//TODO LOAD STRUCTS FROM DATAbasE

var i = 0
var req = []string{"aaa", "bbb", "ccc", "ddd", "eee"}
var res = []string{"111", "222", "333", "444", "555"}
var b = ""

func bootcReply(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	fmt.Println(r.Form)
	fmt.Println("path", r.URL.Path)
	if r.URL.Path == "/"+"aaa" {
		b = res[0]
	}
	if r.URL.Path == "/"+"bbb" {
		b = res[1]
	}
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])

	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, b) // send data to client side
}

func main() {
	http.HandleFunc("/", bootcReply)
	err := http.ListenAndServe(":9091", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
