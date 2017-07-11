// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"net"
	"os/exec"
	"strings"

	"github.com/platinasystems/go/goes/cmd/vnetd"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/platforms/mk1"
)

func init() { vnetd.CloseHook = defaultMk1.stopHook }

func (p *mk1Main) stopHook(i *vnetd.Info, v *vnet.Vnet) error {
	if p.SriovMode {
		return mk1.PlatformExit(v, &p.Platform)
	} else {
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
