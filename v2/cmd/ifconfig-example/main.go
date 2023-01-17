// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"os"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const Usage = `
usage: {{.}} [<ifname> [<command> [<args>]]]

Commands
  up
  down
  add <prefix> [<broadcast>]
  del <prefix>
`

var (
	ErrNoPrefix = errors.New("missing <prefix>")
	ErrUnknown  = errors.New("unknown")
)

func main() {
	var p netip.Prefix
	nargs := len(os.Args)
	if nargs == 2 && os.Args[1] == "-h" {
		template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(os.Stdout, program.Base())
		return
	}
	if nargs == 1 {
		itfs, err := net.Interfaces()
		if err != nil {
			fmt.Println(err)
			return
		}
		for i := range itfs {
			show(&itfs[i])
		}
		return
	}
	itf, err := net.InterfaceByName(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	if nargs == 2 {
		show(itf)
		return
	}
	cmd := os.Args[2]
	switch cmd {
	case "up":
		err = netif.Up(itf.Name)
	case "down":
		err = netif.Down(itf.Name)
	case "add":
		if nargs < 4 {
			err = ErrNoPrefix
		} else if p, err = netip.ParsePrefix(os.Args[3]); err == nil {
			bc := netip.IPv4Unspecified()
			if p.Addr().Is4() && nargs > 4 {
				bc, err = netip.ParseAddr(os.Args[4])
				if err == nil {
					err = netif.Add(itf.Name, p, bc)
				}
			} else {
				err = netif.Add(itf.Name, p, bc)
			}
		}
	case "del":
		if nargs < 4 {
			err = ErrNoPrefix
		} else if p, err = netip.ParsePrefix(os.Args[3]); err == nil {
			err = netif.Del(itf.Name, p)
		}
	default:
		err = ErrUnknown
	}
	if err != nil {
		fmt.Print(itf.Name, ":", cmd, ": ", err, "\n")
	}
}

func show(itf *net.Interface) {
	defer fmt.Println()
	fmt.Printf("%s[%d]:", itf.Name, itf.Index)
	if flags, err := netif.Flags(itf); err == nil {
		fmt.Printf(" flags=%04x<%v>", uint(flags), flags)
	} else {
		fmt.Printf(" flags=%04x<%v>", uint(itf.Flags), flags)
	}
	fmt.Print(" mtu ", itf.MTU)
	if len(itf.HardwareAddr) > 0 {
		fmt.Print("\n\tether ", itf.HardwareAddr)
	}
	if addrs, err := itf.Addrs(); err == nil {
		for i, a := range addrs {
			if ipn, ok := a.(*net.IPNet); ok {
				if ipn.IP.To4() != nil {
					fmt.Print("\n\tinet ", ipn)
				} else if ipn.IP.To16() != nil {
					fmt.Print("\n\tinet6 ", ipn)
				}
			} else {
				fmt.Printf("\n\taddress[%d]: %T", i, a)
			}
		}
	}
}
