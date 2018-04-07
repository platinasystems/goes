// Copyright © 2015-2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bootd

import (
	"fmt"
	"time"
)

//TODO CONFORM TO MSG TYPE
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
	s = "script"
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
	s += "STALL-TYPE     REGISTERED          INSTALLED       #INST\n"
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
