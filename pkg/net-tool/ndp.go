// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tool

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/netrt"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

func NDP(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [args]
Control/diagnose IPv6 neighbor discovery protocol
{{flags .}}`)

	defineNDPFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	switch {
	case len(ndp_f) > 0:
		err = script()
	case ndp_a || ndp_c:
		_, err = dump(ctx)
	case len(ndp_d) > 0:
		err = remove()
	case len(ndp_i) > 0:
		err = ifinfo(ctx, ndp_i)
	case len(ndp_I) > 0:
		err = defif()
	case ndp_p:
		err = prefixes()
	case ndp_r:
		err = routers()
	case ndp_s:
		err = set()
	case ndp_H:
		err = harmonize()
	case ndp_P || ndp_R:
		err = flush()
	default:
		if flag.NArg() == 0 {
			flag.Usage()
		} else if ss, err := net.LookupHost(flag.Arg(0)); err != nil {
			return err
		} else {
			var as []netip.Addr
			for _, s := range ss {
				if a, err := netip.ParseAddr(s); err == nil {
					as = append(as, a)
				}
			}
			_, err = dump(ctx, as...)
		}
	}
	return err
}

var (
	// NDP Flags
	ndp_A = 0
	ndp_H = false
	ndp_I = ""
	ndp_P = false
	ndp_R = false
	ndp_a = false
	ndp_c = false
	ndp_d = ""
	ndp_f = ""
	ndp_i = ""
	ndp_l = false
	ndp_n = false
	ndp_p = false
	ndp_r = false
	ndp_s = false
	ndp_t = false
	ndp_x = false
	ndp_w = false
)

func defineNDPFlags() {
	xflag.Define(&ndp_A, "A", "Repeat show interval (seconds).")
	xflag.Define(&ndp_H, "H", "Harmonize routing and neighbor tables.")
	xflag.Define(&ndp_I, "I", "Set, “show” or “delete” default interface.")
	xflag.Define(&ndp_P, "P", "Flush all the entries in the prefix list.")
	xflag.Define(&ndp_R, "R",
		"Flush all the entries in the default router list.")
	xflag.Define(&ndp_a, "a", "Show current entries.")
	xflag.Define(&ndp_c, "c", "Erase all entries.")
	xflag.Define(&ndp_d, "d", "Delete specified entry.")
	xflag.Define(&ndp_f, "f", "Table configuration file.")
	xflag.Define(&ndp_i, "i",
		"View information for the specified interface.")
	xflag.Define(&ndp_l, "l", "Show link-layer reachability information.")
	xflag.Define(&ndp_n, "n",
		"Don't resolve numeric addresses to hostnames.")
	xflag.Define(&ndp_p, "p", "Show prefix list.")
	xflag.Define(&ndp_r, "r", "Show default router list.")
	xflag.Define(&ndp_s, "s", "Register an NDP entry for a node.")
	xflag.Define(&ndp_t, "t", "Show timestamp for each entry.")
	xflag.Define(&ndp_x, "x",
		"Show extended link-layer reachability information.")
	xflag.Define(&ndp_w, "w",
		"Show node's cryptographically generated address.")
}

func dump(ctx context.Context, match ...netip.Addr) (int, error) {
	const heading = `
Neighbor                                Linklayer Address  Netif Expire    St Flgs Prbs`
	ismatch := func(netip.Addr) bool { return true }
	if len(match) > 0 {
		ismatch = func(dst netip.Addr) bool {
			for _, addr := range match {
				if addr.Compare(dst) == 0 {
					return true
				}
			}
			return false
		}
	}
	nifs := make(map[int]*netif.NetIf)
	nder, err := netrt.Neighbors(ctx, xnet.AF_INET6)
	if err != nil {
		return 0, xerrors.Mark(err)
	}
	defer nder.Close()
	if !ndp_t {
		fmt.Println(heading[1:])
	}
	count := 0
	nd, err := nder.NextNd(ctx)
	for ; ctx.Err() == nil && err == nil && nd != nil; nd, err =
		nder.NextNd(ctx) {
		utc := time.Now().UTC()
		dst := nd.Dst()
		if !ismatch(dst) || dst.IsLoopback() {
			continue
		}
		i := nd.Line()
		if i <= 0 {
			i = nd.Index()
		}
		nif, ok := nifs[i]
		if !ok {
			nif, err = xerrors.MarkResult(netif.Indexed(ctx, i))
			if err != nil {
				break
			} else if len(nif.Name) == 0 {
				continue
			}
			nifs[i] = nif
		}
		if dst.IsLinkLocalUnicast() {
			dst = dst.WithZone(nif.Name)
		}
		name := dst.String()
		if !ndp_n {
			l, err := net.LookupAddr(name)
			if err == nil && len(l) > 0 {
				name = l[0]
			}
		}
		if ndp_t {
			fmt.Print(utc.Format("15:04:05.000000"), " ")
		}
		fmt.Printf("%-39s ", name)
		ha := nd.HA()
		if len(ha) > 0 {
			fmt.Printf("%-17v ", ha)
		} else {
			fmt.Printf("%-17v ", "(incomplete)")
		}
		fmt.Printf("%6.6s ", nif.Name)
		if expire := nd.Expire(); expire == 0 {
			fmt.Printf("permanent ")
		} else if epoch := int(utc.Unix()); expire <= epoch {
			fmt.Printf("expired   ")
		} else {
			fmt.Printf("%9d ", expire-epoch)
		}
		fmt.Print(ndpState(nd.State()), "  ")
		fmt.Printf("%4s ", ndpFlags(nd.Flags()))
		if v, ok := nd.(netrt.Probes); ok {
			if prbs := v.Probes(); prbs != 0 {
				fmt.Printf("%4d", prbs)
			}
		}
		fmt.Println()
		count += 1
	}
	return count, err
}

func defif() error {
	return xerrors.ErrUnsupported
}

func flush() error {
	return xerrors.ErrUnsupported
}

func harmonize() error {
	return xerrors.ErrUnsupported
}

func ifinfo(ctx context.Context, ifname string) error {
	fmt.Printf("linkmtu=%d, ", 1500)
	fmt.Printf("curhlim=%d, ", 0)
	fmt.Printf("basereachable=%v, ", 0*time.Microsecond)
	fmt.Printf("reachable=%v, ", 0*time.Microsecond)
	fmt.Printf("retrans=%v\n", 0*time.Microsecond)
	fmt.Printf("Flags: %#x %v\n", 0, "FIXME")
	return nil
}

func prefixes() error {
	return xerrors.ErrUnsupported
}

func remove() error {
	return xerrors.ErrUnsupported
}

func routers() error {
	return xerrors.ErrUnsupported
}

func script() error {
	return xerrors.ErrUnsupported
}

func set() error {
	return xerrors.ErrUnsupported
}
