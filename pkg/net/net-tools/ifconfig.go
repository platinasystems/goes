// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
	"regexp"
	"slices"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

var inets = []string{"inet", "inet6"}

func Ifconfig(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<option>]... [<parameter>]...
Configure and display network interface parameters.

  • {{$path}} [<modifier(s)>] [<filter>]
    Display parameters of matching interfaces.
  • {{$path}} <name> <prefix> [<destination>] [<command>] [<parameter>]...
    Add or delete a network prefix.
    A point-to-point interface also requires the remote destination.
  • {{$path}} <filter> [<parameter>]...
    Configure link parameters of matching interfaces. 
  • {{$path}} <device|name> create [<parameter>]...
    Create the specified network pseudo-device with given name or auto-named
    with cloneable device prefix.
  • {{$path}} <name> destroy
    Destroy the named pseudo-device.
  • {{$path}} -C
    List cloneable devices.
  • {{$path}} -l <filter>
    List matching interfaces.

Filters
	{ -a, -d, -u, -X <pattern, <family>, <name> }

Modifiers
	{ -L, -m, -r, -v }

Options{{print .Flags}}
Commands` + netif.AddressCommands + `
Parameters` + netif.AddressParameters +
		netif.ConfigParameters +
		netif.CreateParameters
	var pat *regexp.Regexp
	fs, h := flag.New()
	mFlag := fs.Bool("m", false, "Display all supported media.")
	LFlag := fs.Bool("L", false, "Display IPv6 address lifetime as offset.")
	aFlag := fs.Bool("a", false,
		"Display all interfaces (implied unless -d, -u, -X).")
	dFlag := fs.Bool("d", false, "Only display down interfaces.")
	uFlag := fs.Bool("u", false, "Only display up interfaces.")
	lFlag := fs.Bool("l", false, "List available interfaces.")
	vFlag := fs.Bool("v", false, "Verbose display.")
	CFlag := fs.Bool("C", false, "List cloneable devices.")
	rFlag := fs.Bool("r", false, "Display route references.")
	XFlag := fs.String("X", "", "Pattern match interface name.")
	// FIXME add these display modifiers
	_ = *mFlag || *LFlag || *vFlag || *rFlag
	if complete.Parameter.Value(ctx) {
		if len(args) < 2 {
			ncloneable := len(netif.Cloneable)
			nifs, _ := netif.List()
			names := make([]string, ncloneable+len(nifs))
			copy(names, netif.Cloneable)
			for i, nif := range nifs {
				names[ncloneable+i] = nif.Name
			}
			style.Completions(args, names)
		}
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if help.Parameter.Value(ctx) || *h {
		return style.Usage(usage, struct {
			Path  []string
			Flags fmt.Formatter
		}{path, fs})
	}
	if len(*XFlag) > 0 {
		if pat, err = regexp.Compile(*XFlag); err != nil {
			return err
		}
	}
	args = fs.Args()
	nifs, err := netif.List()
	if err != nil {
		return err
	}
	named := make(map[string]*netif.Netif)
	for _, nif := range nifs {
		named[nif.Name] = nif
	}
	switch {
	case *CFlag:
		var sep string
		for _, dev := range netif.Cloneable {
			fmt.Fprint(w, sep, dev)
			sep = " "
		}
		if len(sep) > 0 {
			fmt.Fprintln(w)
		}
		return nil
	case *lFlag:
		var sep string
		for _, nif := range nifs {
			isup := (nif.Flags & net.FlagUp) == net.FlagUp
			if (*dFlag && !isup) || (*uFlag && isup) ||
				(!*dFlag && !*uFlag) {
				// FIXME family filter
				fmt.Fprint(w, sep, nif.Name)
				sep = " "
			}
		}
		if len(sep) > 0 {
			fmt.Fprintln(w)
		}
		return nil
	case *aFlag || len(args) == 0:
		switch {
		case *dFlag:
			for _, nif := range nifs {
				if pat == nil || pat.MatchString(nif.Name) {
					if nif.Flags&net.FlagUp == 0 {
						fmt.Fprint(w, nif)
					}
				}
			}
		case *uFlag:
			for _, nif := range nifs {
				if pat == nil || pat.MatchString(nif.Name) {
					if nif.Flags&net.FlagUp == net.FlagUp {
						fmt.Fprint(w, nif)
					}
				}
			}
		case pat != nil:
			for _, nif := range nifs {
				if pat.MatchString(nif.Name) {
					fmt.Fprint(w, nif)
				}
			}
		default:
			for _, nif := range nifs {
				fmt.Fprint(w, nif)
			}
		}
		return nil
	case pat != nil:
		for _, nif := range nifs {
			if pat.MatchString(nif.Name) {
				if err = nif.Config(args[1:]); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if len(args) > 1 && args[1] == "create" {
		nif, err := netif.Create(args[0], args[2:]...)
		if err == nil {
			fmt.Fprintln(w, nif.Name)
		}
		return err
	}
	nif, exists := named[args[0]]
	if !exists {
		return fmt.Errorf("%s %w", args[0], ErrNotFound)
	}
	args = args[1:]
	if len(args) == 0 {
		fmt.Fprint(w, nif)
		return nil
	}
	if args[0] == "destroy" {
		return nif.Destroy()
	}
	if slices.Index(inets, args[0]) >= 0 ||
		unicode.IsNumber([]rune(args[0])[0]) {
		return ifconfigAddr(nif, args)
	}
	return nif.Config(args)
}

func ifconfigAddr(nif *netif.Netif, args []string) error {
	if slices.Index(inets, args[0]) >= 0 {
		args = args[1:]
	}
	prefix, err := netip.ParsePrefix(args[0])
	if err != nil {
		return egress.Markf("%s %w", args[0], err)
	}
	addr, bits := prefix.Addr(), prefix.Bits()
	args = args[1:]
	var dest netip.Addr
	if (nif.Flags & net.FlagPointToPoint) != 0 {
		if len(args) == 0 {
			return egress.
				Markf("%w, missing point-to-point destination",
					ErrIncomplete)
		}
		if dest, err = netip.ParseAddr(args[0]); err != nil {
			return egress.Markf("%v %w", args[0], err)
		}
		args = args[1:]
	}
	if len(args) > 0 {
		switch args[0] {
		case "add", "alias":
			return nif.Add(addr, dest, bits, args[1:])
		case "del", "delete", "-alias":
			return nif.Del(addr, dest, bits, args[1:])
		case "change":
			return nif.Change(addr, dest, bits, args[1:])
		case "replace":
			return nif.Replace(addr, dest, bits, args[1:])
		}
	}
	return nif.Add(addr, dest, bits, args)
}
