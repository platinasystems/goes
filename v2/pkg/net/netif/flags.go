// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"fmt"
	"net"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
)

type NamedFlags uint

func Flags(itf *net.Interface) (NamedFlags, error) {
	sock, err := NewInet()
	if err != nil {
		return 0, err
	}
	defer sock.Close()
	var ifr Ifreq
	ifr.Rename(itf.Name)
	if err = sock.ioctl(syscall.SIOCGIFFLAGS, ifr.Pointer()); err != nil {
		return 0, err
	}
	return NamedFlags((*host.Uint16)(ifr.Ifru[:]).Value()), nil
}

func (f NamedFlags) Format(w fmt.State, verb rune) {
	var sep string
	for bit, s := range FlagNames {
		if len(s) > 0 && (f&(1<<NamedFlags(bit))) != 0 {
			fmt.Fprint(w, sep, s)
			sep = ","
		}
	}
}
