// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netstat

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

const NetstatUsage = `
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

{{flags .}}`

var (
	Netstat_F, Netstat_p, Netstat_w int

	Netstat_I, Netstat_f string

	Netstat_i, Netstat_inet, Netstat_inet6, Netstat_m, Netstat_mm,
	Netstat_n, Netstat_r, Netstat_s, Netstat_ss bool
)

var NetstatFlags = xflag.Labels{
	{"F", "FIB number, -1 for current.", &Netstat_F},
	{"I", "Interface name.", &Netstat_I},
	{"f", "Address Family: inet, inet6, link.", &Netstat_f},
	{"i", "Show interface info.", &Netstat_i},
	{"inet", "Address filter.", &Netstat_inet},
	{"inet6", "Address filter.", &Netstat_inet6},
	{"m", "Show memory stats.", &Netstat_m},
	{"mm", "Show detailed memory stats.", &Netstat_mm},
	{"n", "Show numeric address instead of lookup.", &Netstat_n},
	{"p", "Protocol number.", &Netstat_p},
	{"r", "Show routing table.", &Netstat_r},
	{"s", "Show per-protocol stats.", &Netstat_s},
	{"ss", "Show per-protocol, non-zero stats.", &Netstat_ss},
	{"w", "Wait interval.", &Netstat_w},
}

func Netstat(ctx context.Context, args []string) error {
	xflag.TemplateUsage(NetstatUsage)
	err := NetstatFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	args = flag.Args()

	switch {
	case Netstat_i:
		return netstati(ctx)
	case Netstat_r:
		return netstatr(ctx)
	default:
		return xerrors.FIXME("show active sockets")
	}
	return nil
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
	if s := Netstat_I; len(s) > 0 {
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
	if Netstat_inet {
		family = xnet.AF_INET
	} else if Netstat_inet6 {
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
