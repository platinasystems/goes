// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package route

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/netrt"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xexec"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

type Route string

const (
	Add     Route = "add"
	Append  Route = "append"
	Change  Route = "change"
	Delete  Route = "delete"
	Flush   Route = "flush"
	Get     Route = "get"
	Monitor Route = "monitor"
	Prepend Route = "prepend"
	Replace Route = "replace"
	Test    Route = "Test"
)

var (
	routeExpire time.Duration

	routeFib, routeFlags, routeIfa, routeIfp, routeJail, routeMetrics string

	routeHopCount, routeMetric, routeMTU, routeRTT, routeRTTVar,
	routeSSThresh, routeTOS uint

	routeDebug, routeGenMask, routeHost, routeIface, routeInet, routeInet6,
	routeNet, routeNumeric, routeQuiet, routeTestMode, routeVerbose bool

	routePrefixlen = -1
	routeProtocol  = "boot"
	routeScope     = "global"
	routeTable     = "main"
	routeTo        = "unicast"
)

func (rt Route) op(ctx context.Context, args []string) error {
	xflag.TemplateUsage(Usage[rt])

	err := xflag.Labels{
		{"d", "Debug mode.", &routeDebug},
		{"host", "Host <destination>.", &routeHost},
		{"inet", "Address hint or filter.", &routeInet},
		{"4", "aka -inet", &routeInet},
		{"inet6", "Address hint or filter.", &routeInet6},
		{"6", "aka -inet6", &routeInet6},
		{"net", "Network <destination>.", &routeNet},
		{"n", "Numeric address output.", &routeNumeric},
		{"prefixlen",
			"If >= 0, use instead of 1st arg/<suffix> or 3rd arg.",
			&routePrefixlen,
		},
		{"q", "Suppress most output.", &routeQuiet},
		{"t", "Test mode.", &routeTestMode},
		{"v", "Verbose output.", &routeVerbose},
	}.Define()
	if err != nil {
		return err
	}
	if HaveFibs {
		err = xflag.Labels{
			{"fib",
				"A comma separated list of FIB IDs != default.",
				&routeFib,
			},
		}.Define()
		if err != nil {
			return err
		}
	}
	if xexec.CanJail {
		err = xflag.Labels{
			{"j", "Run inside jail.", &routeJail},
		}.Define()
		if err != nil {
			return err
		}
	}

	rt.defineGWFlags()

	if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	if len(routeJail) > 0 {
		if err = xexec.Jail(ctx, routeJail); err != nil {
			return err
		}
	}

	fibs := []int{-1}
	if len(routeFib) > 0 {
		fibs, err = sscanFibs(strings.Split(routeFib, ","))
		if err != nil {
			return err
		}
	}

	dst, gw, err := rt.dst(ctx)
	if err != nil {
		return err
	}

	for _, fib := range fibs {
		if nrt, e := rt.req(ctx, fib, dst, gw); e != nil {
			if fib >= 0 {
				err = fmt.Errorf("fib[%d]: %w", fib, e)
			} else {
				err = e
			}
			break
		} else if rt == "get" {
			show(ctx, nrt)
		}
	}
	return err
}

func (rt Route) String() string { return string(rt) }

func flush(ctx context.Context, args []string) error {
	xflag.TemplateUsage(Usage[Flush])
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()
	fibs, err := sscanFibs(args)
	if err != nil {
		return err
	}
	_ = fibs
	return xerrors.FIXME("tbd")
}

func monitor(ctx context.Context, args []string) error {
	xflag.TemplateUsage(Usage[Monitor])
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()
	fibs, err := sscanFibs(args)
	if err != nil {
		return err
	}
	_ = fibs
	return xerrors.FIXME("tbd")
}

func isNumericAddr(s string) bool {
	r := []rune(s)[0]
	return unicode.IsNumber(r) || r == ':'
}

func sscanFibs(args []string) ([]int, error) {
	if !HaveFibs {
		return []int{-1}, nil
	}
	fibs := make([]int, len(args))
	for i, arg := range args {
		if _, err := fmt.Sscan(arg, &fibs[i]); err != nil {
			return fibs, fmt.Errorf("fib[%d]: %w", i, err)
		}
	}
	return fibs, nil
}

func show(ctx context.Context, nrt netrt.Rt) {
	dstip := nrt.Dst()
	gwip := nrt.GW()
	line := nrt.Line()
	if !dstip.IsValid() {
		return
	}
	fmt.Print(dstip)
	if n := nrt.Bits(); n > 0 {
		fmt.Print("/", n)
	}
	fmt.Print("->")
	if gwip.IsValid() {
		fmt.Print(gwip)
	} else if ha := nrt.HA(); len(ha) > 0 {
		s := strings.Replace(ha.String(), ":", ".", -1)
		fmt.Print(s)
	} else if nif, err := netif.Indexed(ctx, line); err == nil {
		fmt.Print(nif.Name)
	} else {
		fmt.Print("line#", line)
	}
	fmt.Println()
}
