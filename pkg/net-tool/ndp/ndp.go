// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ndp

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

const NdpUsage = `
usage: {{.Name}} [flags] [args]
Control/diagnose IPv6 neighbor discovery protocol
{{flags .}}`

var (
	Ndp_A int

	Ndp_I, Ndp_d, Ndp_f, Ndp_i string

	Ndp_H, Ndp_P, Ndp_R, Ndp_a, Ndp_c, Ndp_l, Ndp_n, Ndp_p, Ndp_r, Ndp_s,
	Ndp_t, Ndp_x, Ndp_w bool
)

var NdpFlags = xflag.Labels{
	{"A", "Repeat show interval (seconds).", &Ndp_A},
	{"H", "Harmonize routing and neighbor tables.", &Ndp_H},
	{"I", "Set, “show” or “delete” default interface.", &Ndp_I},
	{"P", "Flush all the entries in the prefix list.", &Ndp_P},
	{"R", "Flush all the entries in the default router list.", &Ndp_R},
	{"a", "Show current entries.", &Ndp_a},
	{"c", "Erase all entries.", &Ndp_c},
	{"d", "Delete specified entry.", &Ndp_d},
	{"f", "Table configuration file.", &Ndp_f},
	{"i", "View information for the specified interface.", &Ndp_i},
	{"l", "Show link-layer reachability information.", &Ndp_l},
	{"n", "Don't resolve numeric addresses to hostnames.", &Ndp_n},
	{"p", "Show prefix list.", &Ndp_p},
	{"r", "Show default router list.", &Ndp_r},
	{"s", "Register an NDP entry for a node.", &Ndp_s},
	{"t", "Show timestamp for each entry.", &Ndp_t},
	{"x", "Show extended link-layer reachability information.", &Ndp_x},
	{"w", "Show node's cryptographically generated address.", &Ndp_w},
}

func NDP(ctx context.Context, args []string) error {
	xflag.TemplateUsage(NdpUsage)
	err := NdpFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	switch {
	case len(Ndp_f) > 0:
		err = script()
	case Ndp_a || Ndp_c:
		_, err = dump(ctx)
	case len(Ndp_d) > 0:
		err = remove()
	case len(Ndp_i) > 0:
		err = ifinfo(ctx, Ndp_i)
	case len(Ndp_I) > 0:
		err = defif()
	case Ndp_p:
		err = prefixes()
	case Ndp_r:
		err = routers()
	case Ndp_s:
		err = set()
	case Ndp_H:
		err = harmonize()
	case Ndp_P || Ndp_R:
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
	if !Ndp_t {
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
		if !Ndp_n {
			l, err := net.LookupAddr(name)
			if err == nil && len(l) > 0 {
				name = l[0]
			}
		}
		if Ndp_t {
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
