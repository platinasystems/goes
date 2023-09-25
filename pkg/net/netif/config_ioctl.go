// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

const ConfigParameters = `
Config Parameters
  FIXME
`

var ConfigParameter = map[string]Parameter{
	"up":     TrueFlagParameter,
	"-arp":   TrueFlagParameter,
	"no-arp": TrueFlagParameter,
	"down":   FalseFlagParameter,
	"arp":    FalseFlagParameter,
	"mtu":    Uint32Parameter,
	"add":    PrefixParameter,
	"alias":  PrefixParameter,
	"del":    AddrParameter,
	"-alias": AddrParameter,
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

func (nif *Netif) Config(args ...string) error {
	inet, err := af.OpenInet()
	if err != nil {
		return err
	}
	defer af.Close(inet)
	inet6, err := af.OpenInet6()
	if err != nil {
		return err
	}
	defer af.Close(inet6)
	for err == nil && len(args) > 0 {
		change := ConfigFlag[args[0]]
		sioc := ConfigSIOC[args[0]]
		switch parameter := ConfigParameter[args[0]]; parameter {
		case UnknownParameter:
			err = fmt.Errorf("%q %w", args[0], ErrInvalid)
		case TrueFlagParameter:
			args = args[1:]
			req := NewIfreq[IFF](nif.Name)
			err = egress.Marked(IOCTL(inet, SIOCGIFFLAGS, req))
			if err == nil {
				req.Value |= change
				err = egress.
					Marked(IOCTL(inet, SIOCSIFFLAGS, req))
			}
		case FalseFlagParameter:
			args = args[1:]
			req := NewIfreq[IFF](nif.Name)
			err = egress.Marked(IOCTL(inet, SIOCGIFFLAGS, req))
			if err == nil {
				req.Value &^= change
				err = egress.
					Marked(IOCTL(inet, SIOCSIFFLAGS, req))
			}
		case Uint32Parameter:
			req := NewIfreq[uint32](nif.Name)
			if len(args) < 2 {
				err = ErrIncomplete
			} else if _, err = fmt.
				Sscan(args[1], &req.Value); err != nil {
				err = fmt.Errorf("%q %w", args[1], err)
			} else {
				args = args[2:]
				err = egress.Marked(IOCTL(inet, sioc, req))
			}
		case PrefixParameter:
			if len(args) < 2 {
				err = ErrIncomplete
			} else if ip, ipnet, cidrerr := net.
				ParseCIDR(args[1]); cidrerr != nil {
				err = fmt.Errorf("%q %w", args[1], cidrerr)
			} else {
				var dest net.IP
				if len(args) > 2 && args[2] == "dest" {
					if len(args) < 4 {
						err = ErrIncomplete
						break
					}
					dest = net.ParseIP(args[3])
					if dest == nil {
						err = fmt.Errorf("%q %w",
							args[3], ErrInvalid)
						break
					}
					args = args[4:]
				} else {
					args = args[2:]
				}
				if ip4 := ip.To4(); ip4 != nil {
					req := NewInAliasreq(nif.Name,
						ip4, dest, ipnet.Mask)
					err = egress.Marked(IOCTL(inet,
						SIOCAIFADDR, req))
				} else if ip6 := ip.To16(); ip6 != nil {
					req := NewIn6Aliasreq(nif.Name,
						ip6, dest, ipnet.Mask)
					req.Flags |= IN6_IFF_NODAD
					err = egress.Marked(IOCTL(inet6,
						SIOCAIFADDR_IN6, req))
				} else {
					err = fmt.Errorf("%q %w",
						args[0], ErrUnsupported)
				}
			}
		case AddrParameter:
			if len(args) < 2 {
				err = ErrIncomplete
			} else if args[0] != "del" && args[0] != "-alias" {
				err = ErrInvalid
			} else if ip := net.ParseIP(args[1]); ip == nil {
				err = fmt.Errorf("%q %w", args[1], ErrInvalid)
			} else if ip4 := ip.To4(); ip4 != nil {
				args = args[2:]
				req := NewIfreq[SockaddrIn](nif.Name)
				SockaddrInInit(&req.Value, ip4)
				err = egress.Marked(IOCTL(inet,
					SIOCDIFADDR, req))
			} else if ip6 := ip.To16(); ip6 != nil {
				args = args[2:]
				req := NewIfreq[SockaddrIn6](nif.Name)
				SockaddrIn6Init(&req.Value, ip6)
				err = egress.Marked(IOCTL(inet6,
					SIOCDIFADDR_IN6, req))
			} else {
				err = fmt.Errorf("%v: %w",
					ip, ErrUnsupported)
			}
		default:
			err = fmt.Errorf("%q %w", args[0], ErrNotFound)
		}
	}
	return err
}
