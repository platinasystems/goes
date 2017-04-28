// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"net"
	"os/exec"
	"strings"

	"github.com/platinasystems/go/goes/cmd/stop"
	"github.com/platinasystems/go/internal/sriovs"
)

func init() {
	stop.Hook = func() error {
		fns, err := sriovs.NumvfsFns()
		if err == nil && len(fns) > 0 {
			return delSriovs()
		}
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}
		for _, dev := range interfaces {
			if strings.HasPrefix(dev.Name, "eth-") ||
				strings.HasPrefix(dev.Name, "ixge") ||
				strings.HasPrefix(dev.Name, "meth-") {
				exec.Command("/bin/ip", "link", "delete",
					dev.Name).Run()
			}
		}
		return nil
	}
}
