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
	DebugFlag     xflag.Xbool   = "d Debug mode."
	ExpireFlag    xflag.Xint    = "expire Seconds from now."
	FibFlag       xflag.Xstring = "fib A comma separated list of FIB IDs other than default."
	FlagsFlag     xflag.Xstring = "flags A comma separated list." + gwFlags
	GenmaskFlag   xflag.Xstring = "genmask Generate netmask."
	HopCountFlag  xflag.Xuint   = "hopcount FIXME"
	HostFlag      xflag.Xbool   = "host Host <destination>."
	IfaFlag       xflag.Xstring = "ifa A MAC address of a point-to-point peer?"
	IfaceFlag     xflag.Xbool   = "iface Inticates <gateway> is a point-to-point interface name."
	IfpFlag       xflag.Xstring = "ifp A point-to-point peer interface and MAC."
	InetFlag      xflag.Xbool   = `inet Address hint or filter.`
	Inet6Flag     xflag.Xbool   = `inet6 Address hint or filter.`
	JailFlag      xflag.Xstring = `j Run inside jail.`
	MetricFlag    xflag.Xuint   = "metric FIXME"
	MetricsFlag   xflag.Xstring = "metrics A comma separated NAME=VALUE." + gwMetrics
	MTUFlag       xflag.Xuint   = "mtu FIXME"
	NetFlag       xflag.Xbool   = "net Network <destination>."
	NumericFlag   xflag.Xbool   = "n Numeric address output."
	PrefixlenFlag xflag.Xint    = "prefixlen If >= 0, use instead of 1st arg/<suffix> or 3rd arg."
	ProtocolFlag  xflag.Xstring = "protocol {boot, kernel, redirect, static}"
	QuietFlag     xflag.Xbool   = "q Suppress most output."
	RTTFlag       xflag.Xuint   = "rtt FIXME"
	RTTVarFlag    xflag.Xuint   = "rttvar FIXME"
	ScopeFlag     xflag.Xstring = "scope {global, nowhere, host, link, site}"
	SSThreshFlag  xflag.Xuint   = "ssthresh FIXME"
	TableFlag     xflag.Xstring = "table {compat, default, main, local}"
	TestFlag      xflag.Xbool   = "t Test mode."
	ToFlag        xflag.Xstring = "to {unicast, broadcast, blackhole, etc.}"
	TOSFlag       xflag.Xuint   = "tos Set type-of-service."
	VerboseFlag   xflag.Xbool   = "v Verbose output."
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
