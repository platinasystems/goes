// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/flagctx"
	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

const NetstatUsageTemplate = `
usage: {{.Path}} [<option>]...
Show network status.
		
  • {{.Path}} [-AaLlnW] [-f <family> | -p <protocol>]
  • {{.Path}} [-gilns] [-v] [-f <family>] [-I <interface>]
  • {{.Path}} -i | -I <interface> [-w <period>] [-c <queue>] [-abdgqRtS]\n",
  • {{.Path}} -s [-s] [-f <family> | -p <protocol>] [-w <period>]
  • {{.Path}} -i | -I <interface> -s [-f <family> | -p <protocol>]
  • {{.Path}} -m [-m]
  • {{.Path}} -r [-Aaln] [-f <family>]
  • {{.Path}} -rs [-s]
  • {{.Path}} -B [-I interface]
{{.Flag}}`

func NetstatUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: pathctx.StringIn(ctx),
		Flag: flagctx.StringIn(ctx),
	}
}

func Netstat(ctx context.Context, args ...string) error {
	fs := flag.NewSilentFlagSet("netstat")
	ctx = flagctx.Parameter.With(ctx, fs)
	iFlag := fs.Bool("i", false, "Show interface info.")
	rFlag := fs.Bool("r", false, "Show routing table.")
	_ = fs.String("f", "", "Address Family: inet, inet6, link.")
	_ = fs.Int("F", -1, "FIB number, -1 for current.")
	_ = fs.String("I", "", "Interface name.")
	_ = fs.Bool("m", false, "Show memory stats.")
	_ = fs.Bool("mm", false, "Show detailed memory stats.")
	_ = fs.Bool("n", false, "Show numeric address instead of lookup.")
	_ = fs.Int("p", 0, "Protocol number.")
	_ = fs.Bool("s", false, "Show per-protocol stats.")
	_ = fs.Bool("ss", false, "Show per-protocol, non-zero stats.")
	_ = fs.Duration("w", 0, "Wait interval.")
	ctx = flagctx.Parameter.With(ctx, fs)
	if flag.Search[bool]("complete") {
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return usage.Error(NetstatUsageTemplate[1:],
			NetstatUsageData(ctx))
	}
	args = fs.Args()
	switch {
	case *iFlag:
		return netstati(ctx)
	case *rFlag:
		return netstatr(ctx)
	default:
		return errors.New("FIXME")
	}
	return nil
}

func netstati(ctx context.Context) error {
	fs := flagctx.Parameter.In(ctx)
	w := wctx.Parameter.In(ctx)
	fmt.Fprintf(w, "%-15s", "Name")
	fmt.Fprintf(w, " %5s", "MTU")
	fmt.Fprintf(w, " %11s", "Ipkts")
	fmt.Fprintf(w, " %11s", "Ibytes")
	fmt.Fprintf(w, " %11s", "Idrops")
	fmt.Fprintf(w, " %11s", "Ierrs")
	fmt.Fprintf(w, " %11s", "Opkts")
	fmt.Fprintf(w, " %11s", "Obytes")
	fmt.Fprintf(w, " %11s", "Odrops")
	fmt.Fprintf(w, " %11s", "Oerrs")
	fmt.Fprintf(w, " %11s", "Coll")
	fmt.Fprintln(w)
	show := func(nif *netif.NetIf) {
		fmt.Fprintf(w, "%-15s", nif.Name)
		fmt.Fprintf(w, " %5d", nif.MTU)
		fmt.Fprintf(w, " %11d", nif.Rx.Packets)
		fmt.Fprintf(w, " %11d", nif.Rx.Bytes)
		fmt.Fprintf(w, " %11d", nif.Rx.Drops)
		fmt.Fprintf(w, " %11d", nif.Rx.Errors)
		fmt.Fprintf(w, " %11d", nif.Tx.Packets)
		fmt.Fprintf(w, " %11d", nif.Tx.Bytes)
		fmt.Fprintf(w, " %11d", nif.Tx.Drops)
		fmt.Fprintf(w, " %11d", nif.Tx.Errors)
		fmt.Fprintf(w, " %11d", nif.Collisions)
		fmt.Fprintln(w)
	}
	if ifname := flag.Search[string]("I", fs); len(ifname) > 0 {
		if nif := netif.Named(ifname); nif == nil {
			return fmt.Errorf("%q %w", ifname, ErrNotFound)
		} else {
			show(nif)
		}
	} else {
		for _, nif := range netif.Interfaces() {
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
	var family uint
	fs := flagctx.Parameter.In(ctx)
	w := wctx.Parameter.In(ctx)
	switch s := flag.Search[string]("f", fs); s {
	case "":
		family = af.UNSPEC
	case "inet":
		family = af.INET
	case "inet6":
		family = af.INET6
	default:
		return fmt.Errorf("%q %w", s, ErrInvalid)
	}
	nrts, err := netrt.NewList()
	if err != nil {
		return err
	}
	defer nrts.Close()
	dstbuf := new(strings.Builder)
	gwbuf := new(strings.Builder)
	flagbuf := new(strings.Builder)
	var dsts, gws, flags, ifnames []string
	for {
		nrt, err := nrts.Next()
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
		if family == af.INET && dstip.Is6() {
			continue
		}
		if family == af.INET6 && dstip.Is4() {
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
				if nif := netif.Indexed(line); nif != nil {
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
		} else if nif := netif.Indexed(line); nif != nil {
			gwbuf.WriteString(nif.Name)
		} else {
			fmt.Fprint(gwbuf, "line#", line)
		}
		for _, fc := range netrt.NetstatFlagCodes {
			if nrt.Flags()&fc.Flag != 0 {
				flagbuf.WriteRune(fc.Code)
			}
		}
		dsts = append(dsts, dstbuf.String())
		gws = append(gws, gwbuf.String())
		flags = append(flags, flagbuf.String())
		i := nrt.Index()
		if nif := netif.Indexed(i); nif != nil {
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
			fmt.Fprintf(w, "%-*s %-*s %-*s %s\n",
				wdst, dsts[i],
				wgw, gws[i],
				wflags, flags[i],
				ifnames[i])
		}
	}
	return nil
}
