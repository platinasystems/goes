// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'boot controller daemon' to service muliple client installs
// this daemon will run on a "master" ToR and backup ToR
// this daemon reads the config database stored locally or in the cloud
// this daemon contains the boot state machine for each client
// run bootd on all ToR's so they can be activated as a backup master
// normally they would be dormant, can be activated as backup, or
// receive a reboot or other command from the master asyncronously

//TODO TRY DIFF MACHINES INCL CLOUD
//TODO REAL DB, WRITE DB
//TODO CYCLE THROUGH OPTIONS
//TODO MAKE BOOTC GO LIVE IN GOES
//TODO JSON REPLY
//TODO DB read from FILE CLOUD or LITERAL

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

type Client struct {
	unit           int
	name           string
	machine        string
	macAddr        string
	ipAddr         string
	bootState      int
	installState   int
	autoInstall    bool
	certPresent    bool
	installType    int
	timeRegistered string
	timeInstalled  string
	installCounter int
}

var ClientCfg map[string]*Client

func startHandler() (err error) {
	ClientCfg = make(map[string]*Client)
	if err = readClientCfg(); err != nil {
		return
	}

	http.HandleFunc("/", reply)
	if err = http.ListenAndServe(":9090", nil); err != nil {
		log.Print("HTTP Server failed.")
		return
	}
	return
}

//TODO CONFORM TO MSG TYPES
func reply(w http.ResponseWriter, r *http.Request) {
	var b = ""
	var err error

	r.ParseForm()
	t := strings.Replace(r.URL.Path, "/", "", -1)
	u := strings.Split(t, " ")

	switch u[0] {
	case "register":
		if b, err = register(&u); err != nil {
			b = "error registering\n"
		}
	case "dumpvars":
		if b, err = dumpVars(); err != nil {
			b = "error dumping server variables\n"
		}
		b += r.URL.Path + "\n"
		b += t + "\n"
	case "dashboard":
		if b, err = dashboard(); err != nil {
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
