// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"context"
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/af"
	"github.com/platinasystems/goes/v2/pkg/net/netioctl"
	"golang.org/x/sys/unix"
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

var ConfigFlag = map[string]net.Flags{
	"up":   net.FlagUp,
	"down": net.FlagUp,
}

var ConfigIFF = map[string]int{
	"up":     unix.IFF_UP,
	"-arp":   unix.IFF_NOARP,
	"no-arp": unix.IFF_NOARP,
	"down":   unix.IFF_UP,
	"arp":    unix.IFF_NOARP,
}

var ConfigSIOC = map[string]uintptr{
	"mtu": unix.SIOCSIFMTU,
}

func (nif *NetIf) Config(ctx context.Context, args []string) error {
	inet, err := af.OpenInet()
	if err != nil {
		return egress.Mark(err)
	}
	defer af.Close(inet)
	for err == nil && len(args) > 0 {
		switch ConfigParameter[args[0]] {
		case UnknownParameter:
			return egress.Markf("%q %w", args[0], ErrInvalid)
		case TrueFlagParameter:
			err = netioctl.Admin(nif.Name, ConfigIFF[args[0]], 0)
			if err != nil {
				return err
			}
			if flag, ok := ConfigFlag[args[0]]; ok {
				nif.Flags |= flag
			}
			args = args[1:]
		case FalseFlagParameter:
			err = netioctl.Admin(nif.Name, 0, ConfigIFF[args[0]])
			if err != nil {
				return err
			}
			if flag, ok := ConfigFlag[args[0]]; ok {
				nif.Flags &^= flag
			}
			args = args[1:]
		case Uint32Parameter:
			if len(args) < 2 {
				return egress.Markf("%q %w", args[0],
					ErrIncomplete)
			}
			req := netioctl.NewIfReqUint32(nif.Name)
			_, err = fmt.Sscan(args[1], &req.Value)
			if err != nil {
				return egress.Markf("%q %w", args[1], err)
			}
			sioc := ConfigSIOC[args[0]]
			err = egress.Mark(netioctl.Inet(sioc, req))
			args = args[2:]
		default:
			return egress.Markf("%q %w", args[0], ErrNotFound)
		}
	}
	return err
}
