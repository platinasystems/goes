// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package route

import (
	"context"
	"flag"
	"fmt"
	"strings"
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

const (
	DebugFlag  xflag.KeyUsage[bool]   = "d Debug mode."
	ExpireFlag xflag.KeyUsage[int]    = "expire Seconds from now."
	FibFlag    xflag.KeyUsage[string] = "fib " +
		"A comma separated list of FIB IDs other than default."
	FlagsFlag xflag.KeyUsage[string] = "flags " +
		"A comma separated list." + gwFlags
	GenmaskFlag  xflag.KeyUsage[string] = "genmask Generate netmask."
	HopCountFlag xflag.KeyUsage[uint]   = "hopcount FIXME"
	HostFlag     xflag.KeyUsage[bool]   = "host Host <destination>."
	IfaFlag      xflag.KeyUsage[string] = "ifa " +
		"A MAC address of a point-to-point peer?"
	IfaceFlag xflag.KeyUsage[bool] = "iface " +
		"Inticates <gateway> is a point-to-point interface name."
	IfpFlag xflag.KeyUsage[string] = "ifp " +
		"A point-to-point peer interface and MAC."
	InetFlag    xflag.KeyUsage[bool]   = `inet Address hint or filter.`
	Inet6Flag   xflag.KeyUsage[bool]   = `inet6 Address hint or filter.`
	JailFlag    xflag.KeyUsage[string] = `j Run inside jail.`
	MetricFlag  xflag.KeyUsage[uint]   = "metric FIXME"
	MetricsFlag xflag.KeyUsage[string] = "metrics " +
		"A comma separated NAME=VALUE." + gwMetrics
	MTUFlag       xflag.KeyUsage[uint] = "mtu FIXME"
	NetFlag       xflag.KeyUsage[bool] = "net Network <destination>."
	NumericFlag   xflag.KeyUsage[bool] = "n Numeric address output."
	PrefixlenFlag xflag.KeyUsage[int]  = "prefixlen " +
		"If >= 0, use instead of 1st arg/<suffix> or 3rd arg."
	ProtocolFlag xflag.KeyUsage[string] = "protocol " +
		"{boot, kernel, redirect, static}"
	QuietFlag  xflag.KeyUsage[bool]   = "q Suppress most output."
	RTTFlag    xflag.KeyUsage[uint]   = "rtt FIXME"
	RTTVarFlag xflag.KeyUsage[uint]   = "rttvar FIXME"
	ScopeFlag  xflag.KeyUsage[string] = "scope " +
		"{global, nowhere, host, link, site}"
	SSThreshFlag xflag.KeyUsage[uint]   = "ssthresh FIXME"
	TableFlag    xflag.KeyUsage[string] = "table " +
		"{compat, default, main, local}"
	TestFlag xflag.KeyUsage[bool]   = "t Test mode."
	ToFlag   xflag.KeyUsage[string] = "to " +
		"{unicast, broadcast, blackhole, etc.}"
	TOSFlag     xflag.KeyUsage[uint] = "tos Set type-of-service."
	VerboseFlag xflag.KeyUsage[bool] = "v Verbose output."
)

func (rt Route) op(ctx context.Context, args []string) error {
	xflag.TemplateUsage(Usage[rt])

	DebugFlag.Define(false)
	if HaveFibs {
		FibFlag.Define("")
	}
	HostFlag.Define(false)
	InetFlag.Define(false, "4")
	Inet6Flag.Define(false, "6")
	if xexec.CanJail {
		JailFlag.Define("")
	}
	NetFlag.Define(false)
	NumericFlag.Define(false)
	PrefixlenFlag.Define(-1)
	QuietFlag.Define(false)
	TestFlag.Define(false)
	VerboseFlag.Define(false)
	rt.defineGWFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if s := JailFlag.Value(); len(s) > 0 {
		if err = xexec.Jail(ctx, s); err != nil {
			return err
		}
	}

	fibs := []int{-1}
	if s := FibFlag.Value(); len(s) > 0 {
		if fibs, err = sscanFibs(strings.Split(s, ",")); err != nil {
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
