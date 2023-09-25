// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

func (nif *Netif) Destroy() error {
	inet, err := af.Open[af.Inet]()
	if err != nil {
		return err
	}
	defer af.Close(inet)
	req := NewIfreq[Nothing](nif.Name)
	return IOCTL(inet, syscall.SIOCIFDESTROY, req)
}
