// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'boot controller daemon' to service client installs
// this daemon will run on a "master" ToR
// this daemon reads the config database stored in the cloud or locally

//TODO always build-in bootc and bootd.  MASTER=BOOTD CLIENT=BOOTC only 1.TRY-IT-MASTERCHECK /etc/MASTER 2.DRIVECODE run on 2 invaders
//TODO TRY RUNNING BOOTD FROM DIFF MACHINES INCL NON-TOR
//TODO MAKE BOOTC, BOOTD GO LIVE IN GOES WITH HARDCODED ADDRESSES - DONT COMMIT UNTIL ADDRESSING WORKED OUT
//TODO ADD LOCATION
//TODO REAL DB, WRITE DB, (LOCAL, CLOUD, OR LITERAL)
//TODO CYCLE THROUGH OPTIONS FOR CONTACTING CLOUD
//TODO ADD STATE MACHINES LOOP X TIMES

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
	distroType     int
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

var blah = "OKAY"

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
		b += blah + "\n"
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
