// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/platinasystems/goes/internal/prog"
)

// Assert that named interface has carrier w/in 3sec.
func Carrier(netns, ifname string) error {
	const period = 250 * time.Millisecond
	fn := filepath.Join("/sys/class/net", ifname, "carrier")
	for t := 3 * (time.Second / period); t != 0; t-- {
		cmd := exec.Command(prog.Name(), "-test.main",
			"ip", "netns", "exec", netns,
			prog.Name(), "-test.main",
			"cat", fn)
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		if bytes.Equal(output, []byte("1\n")) {
			return nil
		}
		time.Sleep(period)
	}
	return fmt.Errorf("%s no carrier", ifname)
}
