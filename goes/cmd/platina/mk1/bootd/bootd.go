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

func startHandler() error {
	http.HandleFunc("/", reply)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Print("HTTP Server failed.")
	}
	return nil
}

var i = 0

var res = []string{"111", "222", "333", "444", "555"} //FIXME REPLACE WITH JSON, STRUCT

func reply(w http.ResponseWriter, r *http.Request) {
	var b = ""
	var err error

	r.ParseForm()
	t := strings.Replace(r.URL.Path, "/", "", -1) //FIXME process multiple args

	switch t {
	case "register":
		b, err = register()
		if err != nil {
			b = "error registering\n"
		}
	case "dumpvars":
		b, err = dumpVars()
		b += r.URL.Path + "\n"
		b += t + "\n"
		if err != nil {
			b = "error dumping server variables\n"
		}
	case "dashboard":
		b, err = dashboard()
		if err != nil {
			b = "error getting dashboard\n"
		}
	default:
		b = "404\n"
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

func register() (s string, err error) {
	s = "register\n"
	return s, nil
}

func dumpVars() (s string, err error) {
	s = ""
	return s, nil
}

func dashboard() (s string, err error) {
	s = "\n\n"
	s += "PLATINA MASTER ToR BOOT MANAGER DASHBOARD\n"
	s += "\n"
	s += " UNIT        MAC          DB? CERT?    NAME            IP            Install State            Last Boot              Current State\n"
	s += " ====  =================  === =====  =========  ===============  ====================  ======================  ========================\n"
	s += "   1:  00:00:00:00:00:00   Y    N    Invader22  192.168.101.142  Debian-not-installed  2018-03-20:10:10:11.25  Coreboot payload running\n"
	return s, nil
}
