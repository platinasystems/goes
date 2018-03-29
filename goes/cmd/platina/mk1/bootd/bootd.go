// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'boot controller daemon' to service 'boot' requests from muliple clients
// typically this daemon will run on a single "master" ToR instance
// the daemon contains boot state machine for each client
// the daemon reads the config database stored either locally or in the cloud

// DISCLAIMER
// this is a work in progress, this will change significantly before release
// this package must be manually added to the mk1 goes.go to be included ATM

package bootd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/log"
)

// testing infrastructure
// /*
type Command chan struct{}

func (Command) String() string { return "bootd" }

func (Command) Usage() string { return "bootd" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "http boot controller daemon",
	}
}

func (c Command) Close() error {
	close(c)
	return nil
}

func (Command) Kind() cmd.Kind { return cmd.Daemon }

func (c Command) Main(...string) error {
	if err := startHandler(); err != nil {
		return err
	}
	return nil
}

// */

var i = 0
var req = []string{"aaa", "bbb", "ccc", "ddd", "eee"}
var res = []string{"111", "222", "333", "444", "555"}
var b = ""

func reply(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println(r.Form)
	fmt.Println("path", r.URL.Path)

	if r.URL.Path == "/"+"aaa" {
		b = res[0]
	}

	if r.URL.Path == "/"+"dump" {
		b = "\nVARIABLES"
		b += "\n  req: " + r.URL.Path
		b += "\n  res: " + res[1]
		b += "\n  end"
	}

	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	log.Print("reply exit ", b)
	fmt.Fprintf(w, b)
}

func startHandler() error {
	http.HandleFunc("/", reply)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Print("HTTP Server failed.")
	}
	return nil
}
