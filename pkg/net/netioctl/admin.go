// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netioctl

import (
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

func Admin(ifname string, with, without IFF) error {
	inet, err := af.Open[af.Inet]()
	if err != nil {
		return err
	}
	defer af.Close(inet)
	req := NewIfReqUint16(ifname)
	if ioctl(inet, syscall.SIOCGIFFLAGS, req); err != nil {
		return err
	}
	req.Value |= uint16(with)
	req.Value &^= uint16(without)
	return ioctl(inet, syscall.SIOCSIFFLAGS, req)
}
