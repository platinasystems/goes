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

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
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
  • {{$path}} <name> <prefix> [<command>] [<parameter>]...
    Add or delete a network prefix.
  • {{$path}} <name> <address> <destination> [<command>] [<parameter>]...
    Add or delete a point-to-point address.
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

Options{{SprintDefault .Flags}}
Commands` + netif.AddressCommands + `
Parameters` + netif.AddressParameters +
		netif.ConfigParameters +
		netif.CreateParameters
	var pat *regexp.Regexp
	fs := flag.NewSilentFlagSet("ifconfig")
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
	if flag.Search[bool]("complete") {
		if len(args) < 2 {
			nifs := netif.Interfaces()
			namecap := len(nifs) + len(netif.Cloneable)
			names := make([]string, len(nifs), namecap)
			for i, nif := range nifs {
				names[i] = nif.Name
			}
			names = append(names, netif.Cloneable...)
			style.Completions(args, names)
		}
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return style.Usage(usage, struct {
			Path  []string
			Flags *flag.FlagSet
		}{path, fs})
	}
	if len(*XFlag) > 0 {
		if pat, err = regexp.Compile(*XFlag); err != nil {
			return err
		}
	}
	args = fs.Args()
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
		for _, nif := range netif.Interfaces() {
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
			for _, nif := range netif.Interfaces() {
				if pat == nil || pat.MatchString(nif.Name) {
					if nif.Flags&net.FlagUp == 0 {
						fmt.Fprint(w, nif)
					}
				}
			}
		case *uFlag:
			for _, nif := range netif.Interfaces() {
				if pat == nil || pat.MatchString(nif.Name) {
					if nif.Flags&net.FlagUp == net.FlagUp {
						fmt.Fprint(w, nif)
					}
				}
			}
		case pat != nil:
			for _, nif := range netif.Interfaces() {
				if pat.MatchString(nif.Name) {
					fmt.Fprint(w, nif)
				}
			}
		default:
			for _, nif := range netif.Interfaces() {
				fmt.Fprint(w, nif)
			}
		}
		return nil
	case pat != nil:
		for _, nif := range netif.Interfaces() {
			if pat.MatchString(nif.Name) {
				if err = nif.Config(args[1:]); err != nil {
					return err
				}
			}
		}
		return err
	}
	if len(args) > 1 && args[1] == "create" {
		nif, err := netif.Create(args[0], args[2:]...)
		if err == nil {
			fmt.Fprintln(w, nif.Name)
		}
		return err
	}
	nif := netif.Named(args[0])
	if nif == nil {
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

func ifconfigAddr(nif *netif.NetIf, args []string) error {
	var (
		prefix netip.Prefix
		dest   netip.Addr
		err    error
	)
	if slices.Index(inets, args[0]) >= 0 {
		args = args[1:]
	}
	if (nif.Flags & net.FlagPointToPoint) != 0 {
		if len(args) == 0 {
			return egress.Markf("%w, no address", ErrIncomplete)
		}
		if addr, err := netip.ParseAddr(args[0]); err != nil {
			return egress.Markf("%q %w", args[0], err)
		} else if addr.Is4() {
			prefix = netip.PrefixFrom(addr, 32)
		} else if addr.Is6() {
			prefix = netip.PrefixFrom(addr, 128)
		} else {
			return egress.Markf("%q %w", args[0], ErrInvalid)
		}
		if args = args[1:]; len(args) == 0 {
			return egress.
				Markf("%w, no destination", ErrIncomplete)
		}
		if dest, err = netip.ParseAddr(args[0]); err != nil {
			return egress.Markf("%q %w", args[0], err)
		}
		args = args[1:]
	} else if len(args) == 0 {
		return egress.Markf("%w, no prefix", ErrIncomplete)
	} else if prefix, err = netip.ParsePrefix(args[0]); err != nil {
		return egress.Markf("%q %w", args[0], err)
	} else {
		args = args[1:]
	}
	if len(args) > 0 {
		switch args[0] {
		case "add", "alias":
			return nif.Add(prefix, dest, args[1:])
		case "del", "delete", "-alias":
			return nif.Del(prefix, dest, args[1:])
		case "change":
			return nif.Change(prefix, dest, args[1:])
		case "replace":
			return nif.Replace(prefix, dest, args[1:])
		}
	}
	return nif.Add(prefix, dest, args)
}
