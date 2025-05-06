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

func (rt Route) op(ctx context.Context, args []string) error {
	xflag.TemplateUsage(Usage[rt])

	defineRouteOpFlags()
	rt.defineGWFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
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

var (
	// Route Flags.
	routeDebug     = false
	routeExpire    = time.Duration(0)
	routeFib       = ""
	routeFlags     = ""
	routeGenMask   = false
	routeHopCount  = uint(0)
	routeHost      = false
	routeIfa       = ""
	routeIface     = false
	routeIfp       = ""
	routeInet      = false
	routeInet6     = false
	routeJail      = ""
	routeMetric    = uint(0)
	routeMetrics   = ""
	routeMTU       = uint(0)
	routeNet       = false
	routeNumeric   = false
	routePrefixlen = -1
	routeProtocol  = "boot"
	routeQuiet     = false
	routeRTT       = uint(0)
	routeRTTVar    = uint(0)
	routeScope     = "global"
	routeSSThresh  = uint(0)
	routeTable     = "main"
	routeTestMode  = false
	routeTo        = "unicast"
	routeTOS       = uint(0)
	routeVerbose   = false
)

func defineRouteOpFlags() {
	xflag.Define(&routeDebug, "d", "Debug mode.")
	if HaveFibs {
		xflag.Define(&routeFib, "fib",
			"A comma separated list of FIB IDs other than default.")
	}
	xflag.Define(&routeHost, "host", "Host <destination>.")
	xflag.Define(&routeInet, "inet", "Address hint or filter.")
	xflag.Define(&routeInet, "4", "aka -inet")
	xflag.Define(&routeInet6, "inet6", "Address hint or filter.")
	xflag.Define(&routeInet6, "6", "aka -inet6")
	if xexec.CanJail {
		xflag.Define(&routeJail, "j", "Run inside jail.")
	}
	xflag.Define(&routeNet, "net", "Network <destination>.")
	xflag.Define(&routeNumeric, "n", "Numeric address output.")
	xflag.Define(&routePrefixlen, "prefixlen",
		"If >= 0, use instead of 1st arg/<suffix> or 3rd arg.")
	xflag.Define(&routeQuiet, "q", "Suppress most output.")
	xflag.Define(&routeTestMode, "t", "Test mode.")
	xflag.Define(&routeVerbose, "v", "Verbose output.")
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
