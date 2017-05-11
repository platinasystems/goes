// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/internal/sriovs"
)

func delSriovs() error { return sriovs.Del(vfs) }

func newSriovs(ver int) error {
	if ver > 0 {
		sriovs.VfName = func(port, subport uint) string {
			return fmt.Sprintf("eth-%d-%d", port+1, subport+1)
		}
	}
	eth0, err := net.InterfaceByName("eth0")
	if err != nil {
		return err
	}
	mac := sriovs.Mac(eth0.HardwareAddr)
	// skip over eth0, eth1, and eth2
	mac.Plus(3)
	sriovs.VfMac = mac.VfMac
	return sriovs.New(vfs)
}
