// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

func Netstat(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<option>]...
Show network status.

  • {{$path}} [-AaLlnW] [-f <family> | -p <protocol>]
  • {{$path}} [-gilns] [-v] [-f <family>] [-I <interface>]
  • {{$path}} -i | -I <interface> [-w <period>] [-c <queue>] [-abdgqRtS]
  • {{$path}} -s [-s] [-f <family> | -p <protocol>] [-w <period>]
  • {{$path}} -i | -I <interface> -s [-f <family> | -p <protocol>]
  • {{$path}} -m [-m]
  • {{$path}} -r [-Aaln] [-f <family>]
  • {{$path}} -rs [-s]
  • {{$path}} -B [-I interface]

Options{{SprintDefault .Flags}}`
	fs := flag.NewSilentFlagSet("netstat")
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
	if flag.Search[bool]("complete") {
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
	args = fs.Args()
	switch {
	case *iFlag:
		return netstati(ctx, w, fs)
	case *rFlag:
		return netstatr(ctx, w, fs)
	default:
		return errors.New("FIXME")
	}
	return nil
}

func netstati(ctx context.Context, w io.Writer, fs *flag.FlagSet) error {
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

func netstatr(ctx context.Context, w io.Writer, fs *flag.FlagSet) error {
	var family uint
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
