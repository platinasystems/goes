// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tool

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xslices"
)

const ifconfigUsage = `
usage: {{.Name}} [flags]... [parameters]...
This emulates the Unix net-tools command to configure and display network
interface parameters.

  {{.Name}} [modifiers] [filter]
	Display parameters of matching interfaces.

  {{.Name}} <ifname> <prefix> [command] [parameter]...
	Add or delete a network prefix.

  {{.Name}} <ifname> <address> <destination> [command] [parameters]
	Add or delete a point-to-point address.

  {{.Name}} <filter> [<parameter>]...
	Configure link parameters of matching interfaces.

  {{.Name}} <device|ifname> create [parameter]...
	Create the specified network pseudo-device with given name or
	auto-named with cloneable device prefix.

  {{.Name}} <ifname> destroy
	Destroy the named pseudo-device.

  {{.Name}} -C
	List cloneable devices.

  {{.Name}} -l <filter>
	List matching interfaces.

Flags:

{{flags .}}
Filters: { -a, -d, -u, -X <pattern, <family>, <ifname> }

Modifiers: { -L, -m, -r, -v }

Commands:
` + netif.AddressCommands + `
Parameters:
` + netif.AddressParameters + netif.ConfigParameters + netif.CreateParameters

const (
	IfConfig_C_Flag xflag.KeyUsage[bool] = "C " +
		"List cloneable devices."
	IfConfig_L_Flag xflag.KeyUsage[bool] = "L " +
		"Display IPv6 address lifetime as offset."
	IfConfig_X_Flag xflag.KeyUsage[string] = "X " +
		"Pattern match interface name."
	IfConfig_a_Flag xflag.KeyUsage[bool] = "a " +
		"Display all interfaces (implied unless -d, -u, -X)."
	IfConfig_d_Flag xflag.KeyUsage[bool] = "d " +
		"Only display down interfaces."
	IfConfig_l_Flag xflag.KeyUsage[bool] = "l " +
		"List available interfaces."
	IfConfig_m_Flag xflag.KeyUsage[bool] = "m " +
		"Display all supported media."
	IfConfig_r_Flag xflag.KeyUsage[bool] = "r " +
		"Display route references."
	IfConfig_u_Flag xflag.KeyUsage[bool] = "u " +
		"Only display up interfaces."
	IfConfig_v_Flag xflag.KeyUsage[bool] = "v " +
		"Verbose display."
)

var inets = []string{"inet", "inet6"}

func Ifconfig(ctx context.Context, complete bool, args []string) error {
	var pat *regexp.Regexp

	xflag.TemplateUsage(ifconfigUsage)

	CFlag := IfConfig_C_Flag.Define(false)
	IfConfig_L_Flag.Define(false)
	XFlag := IfConfig_X_Flag.Define("")
	aFlag := IfConfig_a_Flag.Define(false)
	dFlag := IfConfig_d_Flag.Define(false)
	lFlag := IfConfig_l_Flag.Define(false)
	IfConfig_m_Flag.Define(false)
	IfConfig_r_Flag.Define(false)
	uFlag := IfConfig_u_Flag.Define(false)
	IfConfig_v_Flag.Define(false)

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	nifs, err := netif.List(ctx)
	if err != nil {
		return err
	}

	if args = flag.Args(); complete {
		if len(args) < 2 {
			var s string
			if len(args) == 1 {
				s = args[0]
			}
			namecap := len(nifs) + len(netif.Cloneable)
			names := make([]string, len(nifs), namecap)
			for i, nif := range nifs {
				names[i] = nif.Name
			}
			names = append(names, netif.Cloneable...)
			c := xslices.Match(names, strings.HasPrefix, s)
			sort.Strings(c)
			for _, s = range c {
				fmt.Println(s)
			}
		}
		return nil
	}

	if len(*XFlag) > 0 {
		if pat, err = regexp.Compile(*XFlag); err != nil {
			return err
		}
	}

	switch {
	case *CFlag:
		var sep string
		for _, dev := range netif.Cloneable {
			fmt.Print(sep, dev)
			sep = " "
		}
		if len(sep) > 0 {
			fmt.Println()
		}
		return nil
	case *lFlag:
		var sep string
		for _, nif := range nifs {
			isup := (nif.Flags & net.FlagUp) == net.FlagUp
			if (*dFlag && !isup) || (*uFlag && isup) ||
				(!*dFlag && !*uFlag) {
				// FIXME family filter
				fmt.Print(sep, nif.Name)
				sep = " "
			}
		}
		if len(sep) > 0 {
			fmt.Println()
		}
		return nil
	case *aFlag || len(args) == 0:
		switch {
		case *dFlag:
			for _, nif := range nifs {
				if pat == nil || pat.MatchString(nif.Name) {
					if nif.Flags&net.FlagUp == 0 {
						fmt.Print(nif)
					}
				}
			}
		case *uFlag:
			for _, nif := range nifs {
				if pat == nil || pat.MatchString(nif.Name) {
					if nif.Flags&net.FlagUp == net.FlagUp {
						fmt.Print(nif)
					}
				}
			}
		case pat != nil:
			for _, nif := range nifs {
				if pat.MatchString(nif.Name) {
					fmt.Print(nif)
				}
			}
		default:
			for _, nif := range nifs {
				fmt.Print(nif)
			}
		}
		return nil
	case pat != nil:
		for _, nif := range nifs {
			if pat.MatchString(nif.Name) {
				err = nif.Config(ctx, args[1:]...)
				if err != nil {
					return err
				}
			}
		}
		return err
	case len(args) > 1 && args[1] == "create":
		nif, err := netif.Create(ctx, args[0], args[2:]...)
		if err == nil {
			fmt.Println(nif.Name)
		}
		return err
	}

	nif := new(netif.NetIf)
	if _, err = fmt.Sscan(args[0], &nif.Index); err != nil {
		nif.Name = args[0]
	}
	if err = nif.Refresh(ctx); err != nil {
		return err
	}

	args = args[1:]
	if len(args) == 0 {
		fmt.Print(nif)
		return nil
	}
	if args[0] == "destroy" {
		return nif.Destroy(ctx)
	}
	if slices.Index(inets, args[0]) >= 0 ||
		unicode.IsNumber([]rune(args[0])[0]) {
		args, err = ifconfigAddr(ctx, nif, args)
	}
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args...)
	}
	return err
}

func ifconfigAddr(ctx context.Context, nif *netif.NetIf, args []string) (
	[]string, error,
) {
	var (
		prefix netip.Prefix
		addr,
		dest netip.Addr
		bits  int
		parms []string
		err   error
	)
	if slices.Index(inets, args[0]) >= 0 {
		args = args[1:]
	}
	if len(args) == 0 {
		return args, xerrors.Incomplete("address")
	}
	slash := strings.IndexRune(args[0], '/')
	percent := strings.IndexRune(args[0], '%')
	switch {
	case 0 < slash && slash < percent:
		prefix, err = netip.ParsePrefix(args[0][:percent])
		if err != nil {
			return args, err
		}
		addr = prefix.Addr().WithZone(args[0][percent+1:])
		bits = prefix.Bits()
	case 0 < percent && percent < slash:
		addr, err = netip.ParseAddr(args[0][:slash])
		if err != nil {
			return args, err
		}
		_, err = fmt.Sscan(args[0][slash+1:], &bits)
		if err != nil {
			return args, err
		}
	case slash < 0:
		addr, err = netip.ParseAddr(args[0][:slash])
		if err != nil {
			return args, err
		}
		if addr.Is4() {
			bits = 32
		} else if addr.Is6() {
			bits = 128
		} else {
			return args, xerrors.Invalid("address", args[0])
		}
	}
	if args = args[1:]; len(args) > 0 {
		if (nif.Flags & net.FlagPointToPoint) != 0 {
			dest, err = netip.ParseAddr(args[0])
			if err == nil {
				args = args[1:]
			}
		}
	}
	for len(args) > 0 {
		switch args[0] {
		case "add", "alias":
			return args[1:],
				nif.Add(ctx, addr, dest, bits, parms...)
		case "del", "delete", "-alias":
			return args[1:],
				nif.Del(ctx, addr, dest, bits, parms...)
		case "change":
			return args[1:],
				nif.Change(ctx, addr, dest, bits, parms...)
		case "replace":
			return args[1:],
				nif.Replace(ctx, addr, dest, bits, parms...)
		default:
			parms = append(parms, args[0])
			args = args[1:]
		}
	}
	return args, nif.Add(ctx, addr, dest, bits, parms...)
}
