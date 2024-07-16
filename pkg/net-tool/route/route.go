// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
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

type modOptions struct {
	debug,
	numeric,
	quiet,
	test,
	verbose *bool
	fibs,
	jail *string
	dst *destinationOptions
	gw  *gatewayOptions
}

func flush(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, flushUsage)
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

func (opts *modOptions) mod(
	ctx context.Context,
	op string,
	args []string,
) error {
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("destination")
	}

	if opts.jail != nil {
		if err = xexec.Jail(ctx, *opts.jail); err != nil {
			return err
		}
	}

	fibs := []int{-1}
	if opts.fibs != nil {
		fibs, err = sscanFibs(strings.Split(*opts.fibs, ","))
		if err != nil {
			return err
		}
	}

	dstarg := args[0]
	args = args[1:]

	var maskarg string
	if len(args) > 1 {
		maskarg = args[1]
	}

	dst, err := opts.dst.prefix(ctx, dstarg, maskarg)
	if err != nil {
		return err
	}

	var gw any
	if opts.gw != nil {
		if len(args) == 0 {
			return xerrors.Incomplete("gateway")
		}
		if gw, err = opts.gw.lookup(ctx, args[0]); err != nil {
			return err
		}
	}

	for _, fib := range fibs {
		nrt, err := opts.req(ctx, op, fib, dst, gw)
		if err != nil {
			if fib >= 0 {
				err = fmt.Errorf("fib[%d]: %w", fib, err)
			}
			return err
		} else if op == "get" {
			show(nrt)
		}
	}
	return nil
}

func monitor(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, monitorUsage)
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

func show(nrt netrt.NetRt) {
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
	} else if nif := netif.Indexed(line); nif != nil {
		fmt.Print(nif.Name)
	} else {
		fmt.Print("line#", line)
	}
	fmt.Println()
}

func newModOptions() *modOptions {
	opts := &modOptions{
		debug: flag.Bool("d", false,
			"Debug mode."),
		numeric: flag.Bool("n", false,
			"Numeric address output."),
		quiet: flag.Bool("q", false,
			"Suppress most output."),
		test: flag.Bool("t", false,
			"Test mode."),
		verbose: flag.Bool("v", false,
			"Verbose output."),
	}
	if haveFibs {
		opts.fibs = flag.String("fib", "",
			"A comma separated list of FIB IDs other than default.")
	}
	if xexec.CanJail {
		opts.jail = flag.String("j", "",
			"Run inside jail.")
	}
	return opts
}
