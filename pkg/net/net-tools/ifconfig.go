// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"fmt"
	"io"
	"net"
	"regexp"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Ifconfig(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<modifier(s)>] [<filter>] [<parameter>]...
Configure and display network interface parameters.

  • {{$path}} [<modifier(s)>] [<filter>]
    Display parameters of matching interfaces.
  • {{$path}} <name> <prefix> <dest> [[add | alias] | [del | -alias]]
    Add (default) or delete <prefix> from named interface with optional
    poinit-to-point <dest> address.
  • {{$path}} <filter> [<parameter>]...
    Configure parameters of matching interfaces. 
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

Options{{print .Flags}}` +
		netif.ConfigParameters +
		netif.CreateParameters
	var (
		pat *regexp.Regexp
		sel []*netif.Netif
	)
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
	FFlag := fs.String("F", "", "Address Family {inet, inet6, link}.")
	XFlag := fs.String("X", "", "Pattern match interface name.")
	_ = *mFlag || *LFlag || *aFlag || *vFlag || *rFlag
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
	args = fs.Args()
	if len(*XFlag) > 0 {
		if pat, err = regexp.Compile(*XFlag); err != nil {
			return err
		}
	}
	family := *FFlag
	if len(args) > 0 &&
		(args[0] == "inet" || args[0] == "inet6" || args[0] == "link") {
		family = args[0]
		args = args[1:]
	}
	if len(args) > 1 && args[1] == "create" {
		nif, err := netif.Create(args[0], args[1:]...)
		if err != nil {
			return err
		}
		fmt.Fprintln(w, nif.Name)
		return nil
	}
	nifs, err := netif.List()
	if err != nil {
		return err
	}
	named := make(map[string]*netif.Netif)
	for _, nif := range nifs {
		named[nif.Name] = nif
	}
	if *CFlag {
		var sep string
		for _, dev := range netif.Cloneable {
			fmt.Fprint(w, sep, dev)
			sep = " "
		}
		if len(sep) > 0 {
			fmt.Fprintln(w)
		}
		return nil
	}
	if *lFlag {
		var sep string
		for _, nif := range nifs {
			isup := (nif.Flags & net.FlagUp) == net.FlagUp
			if (*dFlag && !isup) || (*uFlag && isup) ||
				(!*dFlag && !*uFlag) {
				// FIXME family filter
				_ = family
				fmt.Fprint(w, sep, nif.Name)
				sep = " "
			}
		}
		if len(sep) > 0 {
			fmt.Fprintln(w)
		}
		return nil
	}
	if len(args) == 0 {
		if *dFlag {
			for _, nif := range nifs {
				if nif.Flags&net.FlagUp == 0 {
					fmt.Fprint(w, nif)
				}
			}
		} else if *uFlag {
			for _, nif := range nifs {
				if nif.Flags&net.FlagUp == net.FlagUp {
					fmt.Fprint(w, nif)
				}
			}
		} else if pat != nil {
			for _, nif := range nifs {
				if pat.MatchString(nif.Name) {
					fmt.Fprint(w, nif)
				}
			}
		} else {
			for _, nif := range nifs {
				fmt.Fprint(w, nif)
			}
		}
		return nil
	}
	if len(args) == 1 {
		if nif, ok := named[args[0]]; !ok {
			err = fmt.Errorf("%q not found", args[0])
		} else {
			fmt.Fprint(w, nif)
		}
		return err
	}
	if args[1] == "destroy" {
		if nif, ok := named[args[0]]; !ok {
			err = fmt.Errorf("%q not found", args[0])
		} else {
			err = nif.Destroy()
		}
		return err
	}
	if *dFlag {
		for _, nif := range nifs {
			if nif.Flags&net.FlagUp == 0 {
				sel = append(sel, nif)
			}
		}
	} else if *uFlag {
		for _, nif := range nifs {
			if nif.Flags&net.FlagUp == net.FlagUp {
				sel = append(sel, nif)
			}
		}
	} else if pat != nil {
		for _, nif := range nifs {
			if pat.MatchString(nif.Name) {
				sel = append(sel, nif)
			}
		}
	} else if nif, ok := named[args[0]]; ok {
		sel = append(sel, nif)
		// rearrange args
		if _, _, terr := net.ParseCIDR(args[1]); terr == nil {
			if len(args) > 2 {
				if net.ParseIP(args[2]) != nil {
					if len(args) > 3 &&
						(args[3] == "add" ||
							args[3] == "alias") {
						// <prefix> <dest> {add | alias}
						args = append([]string{
							"add", args[1],
							"dest", args[2],
						}, args[4:]...)
					} else {
						// <prefix> <dest>
						args = append([]string{
							"add", args[1],
							"dest", args[2],
						}, args[3:]...)
					}
				} else if args[2] == "add" || args[2] == "alias" {
					// <prefix> {add | alias}
					args = append([]string{"add", args[1]},
						args[3:]...)
				} else {
					// <prefix> [parameter]...
					args = append([]string{"add", args[1]},
						args[2:]...)
				}
			} else {
				// <prefix>
				args = []string{"add", args[1]}
			}
		} else if net.ParseIP(args[1]) != nil {
			if len(args) > 2 {
				if args[2] == "del" || args[2] == "-alias" {
					// <address> {del || -alias}
					args = []string{"del", args[1]}
				} else {
					return fmt.Errorf("%q invalid", args[2])
				}
			} else {
				return fmt.Errorf("incomplete")
			}
		}
	} else {
		return fmt.Errorf("%q not found", args[0])
	}
	for _, nif := range sel {
		if err = nif.Config(args...); err != nil {
			return err
		}
	}
	return nil
}
