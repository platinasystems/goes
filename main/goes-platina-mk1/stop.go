// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"os/exec"

	"github.com/platinasystems/go/internal/goes/cmd/stop"
)

func init() {
	// FIXME remove this after integrating sriov branch
	if true {
		stop.Hook = stopHook
	}
}

func stopHook() error {
	// Alpha: 0:32
	// Beta: 1:33
	// So, cover all with 0..33
	for port := 0; port < 33; port++ {
		for subport := 0; subport < 4; subport++ {
			exec.Command("/bin/ip", "link", "delete",
				fmt.Sprintf("eth-%d-%d", port, subport),
			).Run()
		}
	}
	for port := 0; port < 2; port++ {
		exec.Command("/bin/ip", "link", "delete",
			fmt.Sprintf("ixge2-0-%d", port),
		).Run()
	}
	for port := 0; port < 2; port++ {
		exec.Command("/bin/ip", "link", "delete",
			fmt.Sprintf("meth-%d", port),
		).Run()
	}
	return nil
}
