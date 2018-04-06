// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// DESCRIPTION
// 'boot controller daemon' to service muliple client installs
// this daemon will run on a "master" ToR and backup ToR
// this daemon reads the config database stored locally or in the cloud
// this daemon contains the boot state machine for each client

//CONST
//UTIL
//SCRIPTS

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

//
//
// FIXME MOVE TO ANOTHER FILE
//
//

// BOOT STATES
const (
	BootStateUnknown = iota
	BootStateMachineOff
	BootStateCoreboot
	BootStateCBLinux
	BootStateCBGoes
	BootStateRegistrationStart
	BootStateRegistrationDone
	BootStateScriptStart
	BootStateScriptExecuting
	BootStateScriptDone
	BootStateBooting
	BootStateUp
)

func bootText(i int) string {
	var bootStates = []string{
		"Unknown",
		"Off",
		"Coreboot",
		"CB-linux",
		"CB-goes",
		"Reg-start",
		"Reg-done",
		"Script-start",
		"Script-exec",
		"Script-done",
		"Booting-linux",
		"Up",
	}
	return bootStates[i]
}

// INSTALL STATES
const (
	InstallStateFactory = iota
	InstallStateInProgess
	InstallStateCompleted
	InstallStateFail
	InstallStateFactoryRestoreStart
	InstallStateFactoryRestoreDone
	InstallStateFactoryRestoreFail
)

func installText(i int) string {
	var installStates = []string{
		"Factory",
		"In-progress",
		"Completed",
		"Install-fail",
		"Restore-start",
		"Restore-done",
		"Restore-fail",
	}
	return installStates[i]
}

// INSTALL TYPES
const (
	Debian = iota
)

func installTypeText(i int) string {
	var installTypes = []string{
		"Debian",
	}
	return installTypes[i]
}

// SCRIPTS
const (
	ScriptBootLatest = iota
	ScriptBootKnownGood
	ScriptInstallDebian
)

func scriptText(i int) string {
	var scripts = []string{
		"Boot-latest",
		"Boot-known-good",
		"Debian-install",
	}
	return scripts[i]
}

// CLIENT MESSAGE TYPES
const (
	BootRequestNormal = iota
	BootRequestUnknownReply
	BootRequestKernelNotFound
	BootRequestRebootLoop
)

type BootRequest struct {
	Request int
}

// SERVER MESSAGE TYPES
const (
	BootReplyNormal = iota
	BootReplyRunGoesScript
	BootReplyExecUsermode
	BootReplyExecKernel
	BootReplyReflashAndReboot
)

type BootReply struct {
	Reply   int
	Binary  []byte
	Payload []byte
}

//
//
// MOVE TO OTHER FILE END
//
//

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

//
//
// MOVE TO OTHER FILE
//
//
func register(u *[]string) (s string, err error) {
	if len(*u) < 3 {
		return "", err
	}
	mac := (*u)[1]
	ClientCfg[mac].ipAddr = (*u)[2]
	ClientCfg[mac].bootState = BootStateRegistrationDone
	ClientCfg[mac].installState = InstallStateInProgess
	t := time.Now()
	ClientCfg[mac].timeRegistered = fmt.Sprintf("%10s",
		t.Format("2006-01-02 15:04:05"))
	ClientCfg[mac].timeInstalled = fmt.Sprintf("%10s",
		t.Format("2006-01-02 15:04:05"))
	ClientCfg[mac].installCounter++
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
	s += "UNIT NAME         MACHINE    MAC-ADDRESS       IP-ADDR"
	s += "ESS        BOOT-STATE   INSTALL-STATE  AUTO  CERT  IN"
	s += "STALL-TYPE   REGISTERED          INSTALLED         #INST\n"
	s += "==== ============ ========== ================= ======="
	s += "======== ============== ============== ===== ===== =="
	s += "========== =================== =================== =====\n"
	siz := len(ClientCfg)
	for j := 1; j <= siz; j++ {
		for i, _ := range ClientCfg {
			if ClientCfg[i].unit == j {
				s += fmt.Sprintf("%-4d ", ClientCfg[i].unit)
				s += fmt.Sprintf("%-12s ", ClientCfg[i].name)
				s += fmt.Sprintf("%-10s ", ClientCfg[i].machine)
				s += fmt.Sprintf("%-17s ", ClientCfg[i].macAddr)
				s += fmt.Sprintf("%-15s ", ClientCfg[i].ipAddr)
				s += fmt.Sprintf("%-14s ",
					bootText(ClientCfg[i].bootState))
				s += fmt.Sprintf("%-14s ",
					installText(ClientCfg[i].installState))
				s += fmt.Sprintf("%-5t ", ClientCfg[i].autoInstall)
				s += fmt.Sprintf("%-5t ", ClientCfg[i].certPresent)
				s += fmt.Sprintf("%-12s ",
					installTypeText(ClientCfg[i].installType))
				s += fmt.Sprintf("%-19s ", ClientCfg[i].timeRegistered)
				s += fmt.Sprintf("%-19s ", ClientCfg[i].timeInstalled)
				s += fmt.Sprintf("%-5d ", ClientCfg[i].installCounter)
				s += "\n"
			}
		}
	}
	return s, nil
}

// FIXME FILE OR CLOUD OR LITERAL
func readClientCfg() (err error) {
	ClientCfg["01:02:03:04:05:06"] = &Client{
		unit:           1,
		name:           "Invader10",
		machine:        "ToR MK1",
		macAddr:        "01:02:03:04:05:06",
		ipAddr:         "0.0.0.0",
		bootState:      BootStateUnknown,
		installState:   InstallStateFactory,
		autoInstall:    true,
		certPresent:    false,
		installType:    Debian,
		timeRegistered: "0000-00-00:00:00:00",
		timeInstalled:  "0000-00-00:00:00:00",
		installCounter: 0,
	}
	ClientCfg["01:02:03:04:05:07"] = &Client{
		unit:           2,
		name:           "Invader11",
		machine:        "ToR MK1",
		macAddr:        "01:02:03:04:05:07",
		ipAddr:         "0.0.0.0",
		bootState:      BootStateUnknown,
		installState:   InstallStateFactory,
		autoInstall:    true,
		certPresent:    false,
		installType:    Debian,
		timeRegistered: "0000-00-00:00:00:00",
		timeInstalled:  "0000-00-00:00:00:00",
		installCounter: 0,
	}
	ClientCfg["01:02:03:04:05:08"] = &Client{
		unit:           3,
		name:           "Invader12",
		machine:        "ToR MK1",
		macAddr:        "01:02:03:04:05:08",
		ipAddr:         "0.0.0.0",
		bootState:      BootStateUnknown,
		installState:   InstallStateFactory,
		autoInstall:    true,
		certPresent:    false,
		installType:    Debian,
		timeRegistered: "0000-00-00:00:00:00",
		timeInstalled:  "0000-00-00:00:00:00",
		installCounter: 0,
	}
	return nil
}

//
//
// MOVE TO OTHER FILE END
//
//
