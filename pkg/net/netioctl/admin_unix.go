// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netioctl

import (
	"github.com/platinasystems/goes/v2/pkg/net/af"
	"golang.org/x/sys/unix"
)

func Admin[T ~int | ~uint](ifname string, with, without T) error {
	inet, err := af.Open[af.Inet]()
	if err != nil {
		return err
	}
	defer af.Close(inet)
	req := NewIfReqUint16(ifname)
	if ioctl(inet, unix.SIOCGIFFLAGS, req); err != nil {
		return err
	}
	req.Value |= uint16(with)
	req.Value &^= uint16(without)
	return ioctl(inet, unix.SIOCSIFFLAGS, req)
}
