// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

const ConfigParameters = `
  up | down
	Enable/Disable named interface.
  arp | -arp
	Enable/Disable Address Resolution Protocol.
  mtu <number>
	Set maximum transmission unit (octets).
`

var ConfigParameter = map[string]Parameter{
	"up":     TrueFlagParameter,
	"-arp":   TrueFlagParameter,
	"no-arp": TrueFlagParameter,
	"down":   FalseFlagParameter,
	"arp":    FalseFlagParameter,
	"mtu":    Uint32Parameter,
}

var ConfigFlag = map[string]IFF{
	"up":     IFF_UP,
	"-arp":   IFF_NOARP,
	"no-arp": IFF_NOARP,
	"down":   IFF_UP,
	"arp":    IFF_NOARP,
}

var ConfigSIOC = map[string]uintptr{
	"mtu": SIOCSIFMTU,
}

func (nif *Netif) Config(args []string) error {
	inet, err := af.OpenInet()
	if err != nil {
		return egress.Marked(err)
	}
	defer af.Close(inet)
	for err == nil && len(args) > 0 {
		switch ConfigParameter[args[0]] {
		case UnknownParameter:
			return egress.Markf("%q %w", args[0], ErrInvalid)
		case TrueFlagParameter:
			req := NewIfreq[IFF](nif.Name)
			err = egress.Marked(IOCTL(inet, SIOCGIFFLAGS, req))
			if err != nil {
				return err
			}
			req.Value |= ConfigFlag[args[0]]
			err = egress.Marked(IOCTL(inet, SIOCSIFFLAGS, req))
			args = args[1:]
		case FalseFlagParameter:
			req := NewIfreq[IFF](nif.Name)
			err = egress.Marked(IOCTL(inet, SIOCGIFFLAGS, req))
			if err != nil {
				return err
			}
			req.Value &^= ConfigFlag[args[0]]
			err = egress.Marked(IOCTL(inet, SIOCSIFFLAGS, req))
			args = args[1:]
		case Uint32Parameter:
			if len(args) < 2 {
				return egress.Markf("%q %w", args[0],
					ErrIncomplete)
			}
			req := NewIfreq[uint32](nif.Name)
			_, err = fmt.Sscan(args[1], &req.Value)
			if err != nil {
				return egress.Markf("%q %w", args[1], err)
			}
			err = egress.Marked(IOCTL(inet, ConfigSIOC[args[0]],
				req))
			args = args[2:]
		default:
			return egress.Markf("%q %w", args[0], ErrNotFound)
		}
	}
	return err
}
