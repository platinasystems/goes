// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/platinasystems/go/internal/prog"
)

// Assert ping response to given address w/in 3sec.
func Ping(netns, addr string) error {
	const period = 250 * time.Millisecond
	for t := 3 * (time.Second / period); t != 0; t-- {
		cmd := exec.Command(prog.Name(), "-test.main",
			"ip", "netns", "exec", netns,
			"ping", "-q", "-c", "1", addr)
		if cmd.Run() == nil {
			return nil
		}
		time.Sleep(period)
	}
	return fmt.Errorf("%s no response", addr)
}
