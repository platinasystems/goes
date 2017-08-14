// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/platinasystems/go/goes/cmd/vnetd"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/platforms/mk1"
)

func init() {
	vnetd.CloseHook = defaultMk1.stopHook
}

func (p *mk1Main) stopHook(i *vnetd.Info, v *vnet.Vnet) error {
	if p.SriovMode {
		return mk1.PlatformExit(v, &p.Platform)
	} else {
		interfaces, err := net.Interfaces()
		if err != nil {
			return err
		}
		cmd := exec.Command("ip", "-batch", "-")
		w, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		if err = cmd.Start(); err != nil {
			return err
		}
		defer func() {
			w.Close()
			xerr := cmd.Wait()
			if ee, found := xerr.(*exec.ExitError); found {
				err = fmt.Errorf("ip -batch: %s",
					string(ee.Stderr))
				fmt.Println("wait err", err)
			}
		}()
		for _, dev := range interfaces {
			if strings.HasPrefix(dev.Name, "eth-") {
				_, err = fmt.Fprintln(w, "link", "delete",
					dev.Name)
				if err != nil {
					fmt.Println("write err", err)
					return err
				}
			}
		}
		return nil
	}
}
