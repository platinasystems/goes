// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package uptime returns the current system uptime.
package uptime

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "uptime" }

func (Command) Usage() string { return "uptime" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "report system uptime",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Display the system uptime.`,
	}
}

func (c Command) Main(...string) (err error) {
	s, err := ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return
	}
	var uptimeSec, uptimeMs, cpuidleSec, cpuidleMs time.Duration

	n, err := fmt.Sscanf(string(s), "%d.%d %d.%d", &uptimeSec, &uptimeMs,
		&cpuidleSec, &cpuidleMs)
	if err != nil {
		return
	}
	if n != 4 {
		return fmt.Errorf("Scanning uptime string %s returns %d items",
			string(s), n)
	}
	uptime := uptimeSec*time.Second + uptimeMs*time.Millisecond
	cpuidle := cpuidleSec*time.Second + cpuidleMs*time.Millisecond

	fmt.Printf("Uptime is %s, idle time %s\n", uptime, cpuidle)

	return nil
}
