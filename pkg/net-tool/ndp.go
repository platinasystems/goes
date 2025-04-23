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
	aFlag := flag.Bool("a", false, "Show current entries.")
	flag.Uint("A", 0, "Repeat show interval (seconds).")
	cFlag := flag.Bool("c", false, "Erase all entries.")
	dFlag := flag.String("d", "", "Delete specified entry.")
	fFlag := flag.String("f", "", "Table configuration file.")
	HFlag := flag.Bool("H", false, "Harmonize routing and neighbor tables.")
	iFlag := flag.String("i", "",
		"View information for the specified interface.")
	IFlag := flag.String("I", "",
		"Set, “show” or “delete” default interface.")
	flag.Bool("l", false, "Show link-layer reachability information.")
	flag.Bool("n", false, "Don't resolve numeric addresses to hostnames.")
	pFlag := flag.Bool("p", false, "Show prefix list.")
	PFlag := flag.Bool("P", false,
		"Flush all the entries in the prefix list.")
	rFlag := flag.Bool("r", false, "Show default router list.")
	RFlag := flag.Bool("R", false,
		"Flush all the entries in the default router list.")
	sFlag := flag.Bool("s", false, "Register an NDP entry for a node.")
	flag.Bool("t", false, "Show timestamp for each entry.")
	flag.Bool("x", false,
		"Show extended link-layer reachability information.")
	flag.Bool("w", false,
		"Show node's cryptographically generated address.")
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	switch {
	case len(*fFlag) > 0:
		err = script()
	case *aFlag || *cFlag:
		_, err = dump(ctx)
	case len(*dFlag) > 0:
		err = remove()
	case len(*iFlag) > 0:
		err = ifinfo(ctx, *iFlag)
	case len(*IFlag) > 0:
		err = defif()
	case *pFlag:
		err = prefixes()
	case *rFlag:
		err = routers()
	case *sFlag:
		err = set()
	case *HFlag:
		err = harmonize()
	case *PFlag || *RFlag:
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
	if !xflag.Get[bool]("t") {
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
		if !xflag.Get[bool]("n") {
			l, err := net.LookupAddr(name)
			if err == nil && len(l) > 0 {
				name = l[0]
			}
		}
		if xflag.Get[bool]("t") {
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
