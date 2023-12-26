// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
)

const routeUsage = `
usage: {{branch .}} [<option>]... [<command> [[<modifer>]... <arg>]...]
Manipulate routing tables.

Options{{flags .}}
Commands
{{root .}}`

func Route(ctx context.Context, args []string) error {
	opts := routeOptions()
	ctx = goes.FlagsContext(ctx, opts)
	ctx = goes.RootContext(ctx, routeModBranch)
	if goes.ContextComplete(ctx) {
		return nil
	}
	err := flagset.SilentParse(opts, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			ctx = goes.HelpContext(ctx, true)
		} else {
			return err
		}
	}
	if name := flagset.Search[string](opts, "jail"); len(name) > 0 {
		if err = EnterJail(ctx, name); err != nil {
			return err
		}
	}
	args = opts.Args()
	if len(args) == 0 {
		if goes.ContextHelp(ctx) {
			return goes.Usage(ctx, routeUsage)
		}
		return fmt.Errorf("FIXME aka `netstart -r`")
	}
	return goes.Select(ctx, args)
}

func routeFibs(args []string) ([]int, error) {
	if !haveFibs {
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

func routeFlush(ctx context.Context, args []string) error {
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, routeFlushUsage)
	}
	fibs, err := routeFibs(args)
	if err != nil {
		return err
	}
	_ = fibs
	return FIXME
}

func routeMonitor(ctx context.Context, args []string) error {
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, routeMonitorUsage)
	}
	fibs, err := routeFibs(args)
	if err != nil {
		return err
	}
	_ = fibs
	return FIXME
}

func routeOptions() *flag.FlagSet {
	opts := new(flag.FlagSet)
	opts.Bool("d", false, "Debug mode.")
	if haveFibs {
		opts.String("fib", "",
			"A comma separated list of FIB IDs other than default.")
	}
	if HaveJails {
		opts.String("j", "", "Run inside jail.")
	}
	opts.Bool("n", false, "Numeric address output.")
	opts.Bool("q", false, "Suppress all output.")
	opts.Bool("t", false, "Test mode.")
	opts.Bool("v", false, "Verbose output.")
	return opts
}

func routeModCmd(ctx context.Context, args []string) error {
	branch := goes.ContextBranch(ctx)
	cmd := branch[len(branch)-1]

	opts := goes.ContextFlags(ctx)
	s := flagset.Search[string](opts, "fib")
	fibs, err := routeFibs(strings.Split(s, ","))
	if err != nil {
		return err
	}

	dstopts := routeDestinationOptions()
	goes.TemplateFuncs["dstopts"] = func() string {
		return flagset.Sprint(dstopts)
	}

	gwopts := routeGatewayOptions()
	goes.TemplateFuncs["gwopts"] = func() string {
		return flagset.Sprint(gwopts)
	}

	if goes.ContextComplete(ctx) {
		return nil
	}

	err = flagset.SilentParse(dstopts, args)
	if goes.ContextHelp(ctx) || errors.Is(err, flag.ErrHelp) {
		return goes.Usage(ctx, routeModUsage[cmd])
	} else if err != nil {
		return err
	}

	args = dstopts.Args()
	if len(args) > 0 {
		return ErrNoDst
	}
	dstarg := args[0]

	err = flagset.SilentParse(gwopts, args[1:])
	if errors.Is(err, flag.ErrHelp) {
		return goes.Usage(ctx, routeModUsage[cmd])
	} else if err != nil {
		return err
	}
	args = gwopts.Args()

	var maskarg string
	if len(args) > 1 {
		maskarg = args[1]
	}

	dst, err := routeDestination(ctx, dstarg, maskarg, dstopts)
	if err != nil {
		return err
	}

	var gw any
	if len(args) > 0 {
		gw, err = routeGateway(ctx, args[0], gwopts)
		if err != nil {
			return err
		}
	}

	w := goes.ContextStdout(ctx)
	for _, fib := range fibs {
		nrt, err := routeModReq(ctx, cmd, fib, dst, gw, gwopts)
		if err != nil {
			if fib >= 0 {
				err = fmt.Errorf("fib[%d]: %w", fib, err)
			}
			return err
		} else if cmd == "get" {
			routeShow(w, nrt)
		}
	}
	return nil
}

func routeShow(w io.Writer, nrt netrt.NetRt) {
	dstip := nrt.Dst()
	gwip := nrt.GW()
	line := nrt.Line()
	if !dstip.IsValid() {
		return
	}
	fmt.Fprint(w, dstip)
	if n := nrt.Bits(); n > 0 {
		fmt.Fprint(w, "/", n)
	}
	fmt.Fprint(w, "->")
	if gwip.IsValid() {
		fmt.Fprint(w, gwip)
	} else if ha := nrt.HA(); len(ha) > 0 {
		s := strings.Replace(ha.String(), ":", ".", -1)
		fmt.Fprint(w, s)
	} else if nif := netif.Indexed(line); nif != nil {
		fmt.Fprint(w, nif.Name)
	} else {
		fmt.Fprint(w, "line#", line)
	}
	fmt.Fprintln(w)
}
