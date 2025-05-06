// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tool

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/netrt"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

func Netstat(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Prints network status.

  {{.Name}} [-AaLlnW] [-f <family> | -p <protocol>]
  {{.Name}} [-gilns] [-v] [-f <family>] [-I <interface>]
  {{.Name}} -i | -I <interface> [-w <period>] [-c <queue>] [-abdgqRtS]
  {{.Name}} -s [-s] [-f <family> | -p <protocol>] [-w <period>]
  {{.Name}} -i | -I <interface> -s [-f <family> | -p <protocol>]
  {{.Name}} -m [-m]
  {{.Name}} -r [-Aaln] [-4|-6]
  {{.Name}} -rs [-s]
  {{.Name}} -B [-I interface]

Flags:

{{flags .}}`)

	defineNetstatFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	args = flag.Args()

	switch {
	case netstat_i:
		return netstati(ctx)
	case netstat_r:
		return netstatr(ctx)
	default:
		return xerrors.FIXME("show active sockets")
	}
	return nil
}

var (
	// Netstat Flags.
	netstat_F     = -1
	netstat_I     = ""
	netstat_f     = ""
	netstat_i     = false
	netstat_inet  = false
	netstat_inet6 = false
	netstat_m     = false
	netstat_mm    = false
	netstat_n     = false
	netstat_p     = 0
	netstat_r     = false
	netstat_s     = false
	netstat_ss    = false
	netstat_w     = 0
)

func defineNetstatFlags() {
	xflag.Define(&netstat_F, "F", "FIB number, -1 for current.")
	xflag.Define(&netstat_I, "I", "Interface name.")
	xflag.Define(&netstat_f, "f", "Address Family: inet, inet6, link.")
	xflag.Define(&netstat_i, "i", "Show interface info.")
	xflag.Define(&netstat_inet, "inet", "Address filter.")
	xflag.Define(&netstat_inet6, "inet6", "Address filter.")
	xflag.Define(&netstat_m, "m", "Show memory stats.")
	xflag.Define(&netstat_mm, "mm", "Show detailed memory stats.")
	xflag.Define(&netstat_n, "n", "Show numeric address instead of lookup.")
	xflag.Define(&netstat_p, "p", "Protocol number.")
	xflag.Define(&netstat_r, "r", "Show routing table.")
	xflag.Define(&netstat_s, "s", "Show per-protocol stats.")
	xflag.Define(&netstat_ss, "ss", "Show per-protocol, non-zero stats.")
	xflag.Define(&netstat_w, "w", "Wait interval.")
}

func netstati(ctx context.Context) error {
	fmt.Printf("%-15s", "Name")
	fmt.Printf(" %5s", "MTU")
	fmt.Printf(" %11s", "Ipkts")
	fmt.Printf(" %11s", "Ibytes")
	fmt.Printf(" %11s", "Idrops")
	fmt.Printf(" %11s", "Ierrs")
	fmt.Printf(" %11s", "Opkts")
	fmt.Printf(" %11s", "Obytes")
	fmt.Printf(" %11s", "Odrops")
	fmt.Printf(" %11s", "Oerrs")
	fmt.Printf(" %11s", "Coll")
	fmt.Println()
	show := func(nif *netif.NetIf) {
		fmt.Printf("%-15s", nif.Name)
		fmt.Printf(" %5d", nif.MTU)
		fmt.Printf(" %11d", nif.Rx.Packets)
		fmt.Printf(" %11d", nif.Rx.Bytes)
		fmt.Printf(" %11d", nif.Rx.Drops)
		fmt.Printf(" %11d", nif.Rx.Errors)
		fmt.Printf(" %11d", nif.Tx.Packets)
		fmt.Printf(" %11d", nif.Tx.Bytes)
		fmt.Printf(" %11d", nif.Tx.Drops)
		fmt.Printf(" %11d", nif.Tx.Errors)
		fmt.Printf(" %11d", nif.Collisions)
		fmt.Println()
	}
	if s := netstat_I; len(s) > 0 {
		nif, err := netif.Named(ctx, s)
		if err != nil {
			return err
		} else {
			show(nif)
		}
	} else if nifs, err := netif.List(ctx); err != nil {
		return err
	} else {
		for _, nif := range nifs {
			select {
			case <-ctx.Done():
				return nil
			default:
				show(nif)
			}
		}
	}
	return nil
}

func netstatr(ctx context.Context) error {
	family := xnet.AF_UNSPEC
	if netstat_inet {
		family = xnet.AF_INET
	} else if netstat_inet6 {
		family = xnet.AF_INET6
	}
	rter, err := netrt.Routes(ctx, family)
	if err != nil {
		return err
	}
	defer rter.Close()
	dstbuf := new(strings.Builder)
	gwbuf := new(strings.Builder)
	flagbuf := new(strings.Builder)
	var dsts, gws, flags, ifnames []string
	for {
		nrt, err := rter.NextRt(ctx)
		if err != nil {
			return err
		} else if nrt == nil {
			break
		}
		dstip := nrt.Dst()
		gwip := nrt.GW()
		line := nrt.Line()
		if !dstip.IsValid() {
			continue
		}
		dstbuf.Reset()
		gwbuf.Reset()
		flagbuf.Reset()
		if dstip.IsUnspecified() {
			fmt.Fprint(dstbuf, "default")
		} else {
			fmt.Fprint(dstbuf, dstip)
		}
		if dstip.Is6() && !gwip.IsValid() {
			if line > 0 {
				nif, err := netif.Indexed(ctx, line)
				if err == nil {
					fmt.Fprint(dstbuf, "%", nif.Name)
				} else {
					fmt.Fprint(dstbuf, "%line#", line)
				}
			}
		}
		if n := nrt.Bits(); n > 0 {
			fmt.Fprint(dstbuf, "/", n)
		}
		if gwip.IsValid() {
			fmt.Fprint(gwbuf, gwip)
		} else if ha := nrt.HA(); len(ha) > 0 {
			s := ha.String()
			gwbuf.WriteString(strings.Replace(s, ":", ".", -1))
		} else if nif, err := netif.Indexed(ctx, line); err == nil {
			gwbuf.WriteString(nif.Name)
		} else {
			fmt.Fprint(gwbuf, "line#", line)
		}
		rtf := nrt.Flags()
		for i, c := range rtfmark {
			if (rtf & (1 << i)) != 0 {
				flagbuf.WriteRune(c)
			}
		}
		dsts = append(dsts, dstbuf.String())
		gws = append(gws, gwbuf.String())
		flags = append(flags, flagbuf.String())
		i := nrt.Index()
		if nif, err := netif.Indexed(ctx, i); err == nil {
			ifnames = append(ifnames, nif.Name)
		} else {
			ifnames = append(ifnames, fmt.Sprint(i))
		}
	}
	wdst := 16
	for _, s := range dsts {
		if n := len(s); n > wdst {
			wdst = n
		}
	}
	wgw := 16
	for _, s := range gws {
		if n := len(s); n > wgw {
			wgw = n
		}
	}
	wflags := 4
	for _, s := range flags {
		if n := len(s); n > wflags {
			wflags = n
		}
	}
	for i := range dsts {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fmt.Printf("%-*s %-*s %-*s %s\n",
				wdst, dsts[i],
				wgw, gws[i],
				wflags, flags[i],
				ifnames[i])
		}
	}
	return nil
}
