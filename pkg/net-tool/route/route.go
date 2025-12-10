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
	RouteExpire time.Duration

	RouteFib, RouteFlags, RouteIfa, RouteIfp, RouteJail, RouteMetrics string

	RouteHopCount, RouteMetric, RouteMTU, RouteRTT, RouteRTTVar,
	RouteSSThresh, RouteTOS uint

	RouteDebug, RouteGenMask, RouteHost, RouteIface, RouteInet, RouteInet6,
	RouteNet, RouteNumeric, RouteQuiet, RouteTestMode, RouteVerbose bool

	RoutePrefixlen = -1
	RouteProtocol  = "boot"
	RouteScope     = "global"
	RouteTable     = "main"
	RouteTo        = "unicast"
)

var Flags = xflag.Labels{
	{"d", "Debug mode.", &RouteDebug},
	{"host", "Host <destination>.", &RouteHost},
	{"inet", "Address hint or filter.", &RouteInet},
	{"4", "aka -inet", &RouteInet},
	{"inet6", "Address hint or filter.", &RouteInet6},
	{"6", "aka -inet6", &RouteInet6},
	{"net", "Network <destination>.", &RouteNet},
	{"n", "Numeric address output.", &RouteNumeric},
	{"prefixlen",
		"If >= 0, use instead of 1st arg/<suffix> or 3rd arg.",
		&RoutePrefixlen,
	},
	{"q", "Suppress most output.", &RouteQuiet},
	{"t", "Test mode.", &RouteTestMode},
	{"v", "Verbose output.", &RouteVerbose},
}

var FibFlag = xflag.
	Label{"fib", "A comma separated list of FIB IDs != default.", &RouteFib}
var JailFlag = xflag.Label{"j", "Run inside jail.", &RouteJail}

func (rt Route) op(ctx context.Context, args []string) error {
	xflag.TemplateUsage(Usage[rt])
	err := Flags.Define()
	if err != nil {
		return err
	}
	if HaveFibs {
		err = xflag.Labels{FibFlag}.Define()
		if err != nil {
			return err
		}
	}
	if xexec.CanJail {
		err = xflag.Labels{JailFlag}.Define()
		if err != nil {
			return err
		}
	}
	if rt == Delete || rt == Get {
		if err = GatewayFlags.Define(); err != nil {
			return err
		}
	}
	if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	if len(RouteJail) > 0 {
		if err = xexec.Jail(ctx, RouteJail); err != nil {
			return err
		}
	}

	fibs := []int{-1}
	if len(RouteFib) > 0 {
		fibs, err = sscanFibs(strings.Split(RouteFib, ","))
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
