// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'boot controller daemon' to service muliple client installs
// this daemon will run on a "master" ToR and backup ToR
// this daemon reads the config database stored locally or in the cloud
// this daemon contains the boot state machine for each client

package bootd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

var cfg map[string]*Client

// FIXME boot states const with string

// FIXME install states const with string

type Client struct {
	name             string
	machine          string
	macAddr          string
	ipAddr           string
	bootState        int
	bootStateStr     string
	installState     int
	installStateStr  string
	autoInstall      bool
	certPresent      bool
	installScript    int
	installScriptStr string
	timeRegistered   string
	timeInstalled    string
	installCounter   int
}

func startHandler() (err error) {
	cfg = make(map[string]*Client)
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

func register(u *[]string) (s string, err error) {
	if len(*u) < 3 {
		return "", err
	}
	mac := (*u)[1]
	cfg[mac].ipAddr = (*u)[2]
	t := time.Now()
	cfg[mac].timeRegistered = fmt.Sprintf("%10s",
		t.Format("2006-01-02 15:04:05"))
	s = "script" // FIXME JSON
	return s, nil
}

func dumpVars() (s string, err error) {
	s = ""
	return s, nil
}

func dashboard() (s string, err error) {
	s = "\n\n"
	s += "PLATINA MASTER ToR - BOOT MANAGER DASHBOARD\n"
	s += "\n"
	s += "CLIENT STATUS\n"
	s += "UNIT  NAME          MACHINE     MAC-ADDRESS        IP-ADDR"
	s += "ESS         BOOT-STATE    INSTALL-STATE   AUTO   CERT   IN"
	s += "STALL-TYPE    REGISTERED           INSTALLED          #INST\n"
	s += "====  ============  ==========  =================  ======="
	s += "========  ==============  ==============  =====  =====  =="
	s += "==========  ===================  ===================  =====\n"
	n := 1
	for i, _ := range cfg {
		s += fmt.Sprintf("%-4d  ", n)
		s += fmt.Sprintf("%-12s  ", cfg[i].name)
		s += fmt.Sprintf("%-10s  ", cfg[i].machine)
		s += fmt.Sprintf("%-17s  ", cfg[i].macAddr)
		s += fmt.Sprintf("%-15s  ", cfg[i].ipAddr)
		s += fmt.Sprintf("%-14s  ", cfg[i].bootStateStr)
		s += fmt.Sprintf("%-14s  ", cfg[i].installStateStr)
		s += fmt.Sprintf("%-5t  ", cfg[i].autoInstall)
		s += fmt.Sprintf("%-5t  ", cfg[i].certPresent)
		s += fmt.Sprintf("%-12s  ", cfg[i].installScriptStr)
		s += fmt.Sprintf("%-19s  ", cfg[i].timeRegistered)
		s += fmt.Sprintf("%-19s  ", cfg[i].timeInstalled)
		s += fmt.Sprintf("%-5d  ", cfg[i].installCounter)
		s += "\n"
		n++
	}
	return s, nil
}

func readClientCfg() (err error) {
	// FIXME READ FILE OR CLOUD
	cfg["01:02:03:04:05:06"] = &Client{
		name:             "Invader10",
		machine:          "ToR MK1",
		macAddr:          "01:02:03:04:05:06",
		ipAddr:           "0.0.0.0",
		bootState:        0,
		bootStateStr:     "OFF",
		installState:     0,
		installStateStr:  "Factory",
		autoInstall:      true,
		certPresent:      false,
		installScript:    0,
		installScriptStr: "Debian",
		timeRegistered:   "0000-00-00:00:00:00",
		timeInstalled:    "0000-00-00:00:00:00",
		installCounter:   0,
	}
	cfg["01:02:03:04:05:07"] = &Client{
		name:             "Invader11",
		machine:          "ToR MK1",
		macAddr:          "01:02:03:04:05:07",
		ipAddr:           "0.0.0.0",
		bootState:        0,
		bootStateStr:     "OFF",
		installState:     0,
		installStateStr:  "Factory",
		autoInstall:      true,
		certPresent:      false,
		installScript:    0,
		installScriptStr: "Debian",
		timeRegistered:   "0000-00-00:00:00:00",
		timeInstalled:    "0000-00-00:00:00:00",
		installCounter:   0,
	}
	cfg["01:02:03:04:05:08"] = &Client{
		name:             "Invader12",
		machine:          "ToR MK1",
		macAddr:          "01:02:03:04:05:08",
		ipAddr:           "0.0.0.0",
		bootState:        0,
		bootStateStr:     "OFF",
		installState:     0,
		installStateStr:  "Factory",
		autoInstall:      true,
		certPresent:      false,
		installScript:    0,
		installScriptStr: "Debian",
		timeRegistered:   "0000-00-00:00:00:00",
		timeInstalled:    "0000-00-00:00:00:00",
		installCounter:   0,
	}
	return nil
}
