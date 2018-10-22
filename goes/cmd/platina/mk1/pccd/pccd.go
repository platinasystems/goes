// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// http server daemon replies to master server requests
// This daemon runs on every ToR under linux, does not run in goes-boot

package pccd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/platina/mk1/bootc"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/log"
)

type Command chan struct{}

func (Command) String() string { return "pccd" }

func (Command) Usage() string { return "pccd" }

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

func startHandler() (err error) {
	bootc.ResetSda6Count()
	bootc.ClientCfg = make(map[string]*bootc.Client)
	bootc.ClientBootCfg = make(map[string]*bootc.BootcConfig)
	if err = readClientCfgDB(); err != nil {
		return
	}

	http.HandleFunc("/", serve)
	if err = http.ListenAndServe(":9090", nil); err != nil {
		log.Print("HTTP Server failed.")
		return
	}
	return
}

func serve(w http.ResponseWriter, r *http.Request) {
	var b = ""
	var err error

	r.ParseForm()
	t := strings.Replace(r.URL.Path, "/", "", -1)
	u := strings.Split(t, " ")

	switch u[0] {
	case "wipe":
		go wipe()
		b = "wipe in progress\n"
	case "rebootlinux":
		go rebootlinux()
		b = "reboot in progress\n"
	case "client":
		if b, err = getClientData(1); err != nil {
			b = "error getting client data\n"
		}
	case "boot":
		if b, err = getClientBootData(1); err != nil {
			b = "error getting client boot data\n"
		}
	case "getconfig":
		if b, err = getConfig(); err != nil {
			b = "error getting config\n"
		}
	case "putconfig":
		if b, err = putConfig(); err != nil {
			b = "error putting config\n"
		}
	case "putclientinfo":
		if b, err = putClientInfo(); err != nil {
			b = "error putting config\n"
		}
	case "putiso":
		if b, err = putIso(); err != nil {
			b = "error storing iso\n"
		}
	case "putpreseed":
		if b, err = putPreseed(); err != nil {
			b = "error storing preseed\n"
		}
	case "puttar":
		if b, err = putTar(); err != nil {
			b = "error storing tar file\n"
		}
	case "putrclocal":
		if b, err = putRcLocal(); err != nil {
			b = "error storing rc.local\n"
		}
	case "reboot":
		if b, err = reBoot(); err != nil {
			b = "error doing reboot\n"
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
	log.Print("serve exit ", b)
	fmt.Fprintf(w, b)
}
