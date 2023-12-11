// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
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
  • {{.Path}} -r [-Aaln] [-4|-6]
  • {{.Path}} -rs [-s]
  • {{.Path}} -B [-I interface]
{{.Flag}}`

func NetstatUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: strings.Join(ctxparm.Strings.In(ctx), " "),
		Flag: ctxparm.SprintFlagsIn(ctx),
	}
}

func Netstat(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	flags := usage.NewFlags("netstat")
	ctx = ctxparm.Flags.With(ctx, flags)
	iFlag := flags.Bool("i", false, "Show interface info.")
	rFlag := flags.Bool("r", false, "Show routing table.")
	afinet := flags.Bool("4", false, "Address filter.")
	flags.BoolVar(afinet, "inet", false, "aka -4.")
	afinet6 := flags.Bool("6", false, "Address filter.")
	flags.BoolVar(afinet6, "inet6", false, "aka -6.")
	_ = flags.String("f", "", "Address Family: inet, inet6, link.")
	_ = flags.Int("F", -1, "FIB number, -1 for current.")
	_ = flags.String("I", "", "Interface name.")
	_ = flags.Bool("m", false, "Show memory stats.")
	_ = flags.Bool("mm", false, "Show detailed memory stats.")
	_ = flags.Bool("n", false, "Show numeric address instead of lookup.")
	_ = flags.Int("p", 0, "Protocol number.")
	_ = flags.Bool("s", false, "Show per-protocol stats.")
	_ = flags.Bool("ss", false, "Show per-protocol, non-zero stats.")
	_ = flags.Duration("w", 0, "Wait interval.")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *usage.Help {
		return usage.Error(NetstatUsageTemplate[1:],
			NetstatUsageData(ctx))
	}
	args = flags.Args()
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
	w := ctxparm.Writer.In(ctx)
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
	if ifname := ctxparm.SearchFlagsIn[string](ctx, "I"); len(ifname) > 0 {
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
	w := ctxparm.Writer.In(ctx)
	streamer, err := netrt.NewList(ctx)
	if err != nil {
		return err
	}
	defer streamer.Close()
	dstbuf := new(strings.Builder)
	gwbuf := new(strings.Builder)
	flagbuf := new(strings.Builder)
	var dsts, gws, flags, ifnames []string
	for {
		nrt, err := streamer.Next()
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
