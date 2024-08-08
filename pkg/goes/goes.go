// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package goes provides a GO-Embedded-System where the importing main program
// [xmaps.Install] [Features] before calling [Exec], e.g.
//
//	var features = []map[string]any{
//		goes_util.Features,
//		core_util.Features,
//		net_tool.Features,
//		...
//	}
//
//	func init() { goes.Install(features...)	}
//
//	main() { goes.Main() }
//
// [Features] may be explict functions or implied by type:
//
//	error	Return.
//
//	map[string]any
//		Select sub-feature with next argument(s).
//
//	func(ctx context.Context, args []string) error
//		Explicit function.
//
//	func(ctx context.Context, complete bool, args []string) error
//		Explicit function thet prints last argument completion when
//		flagged.
//
//	func() (any, error)
//		Imply [PrintResults].
//
// Otherwise, imply [PrintOrScanObject].
//
// Goes has these intrinsic features:
//
//	complete [feature [args]]
//		Print prefix match of last argument.
//
//	help [feature [args]]
//		Print feature usage.
package goes

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xmaps"
	"github.com/platinasystems/goes/v2/pkg/xos"
	"github.com/platinasystems/goes/v2/pkg/xslices"
	"github.com/platinasystems/goes/v2/pkg/xutf8"
)

type Preemption uint8

const (
	None Preemption = iota
	Complete
	Help
)

var Features = make(map[string]any)

func Install(features ...map[string]any) {
	xmaps.Install(Features, features...)
}

// Perform feature in an interruptible context.
//
// Precedence: [os.Args][0], [filepath.Base](os.Args[0]), os.Args[1]
func Main() {
	var err error

	defer func() {
		if err == nil || errors.Is(err, flag.ErrHelp) {
			return
		}
		ecode := 1
		if ee, ok := err.(*exec.ExitError); ok {
			ecode = ee.ExitCode()
			if len(ee.Stderr) == 0 {
				err = nil
			} else {
				err = errors.New(string(ee.Stderr))
			}
		}
		fields := strings.Fields(flag.CommandLine.Name())
		err = xerrors.Label(err, fields...)
		if err != nil {
			os.Stderr.WriteString(xutf8.AlineString(err.Error()))
		}
		if ecode != 0 {
			os.Exit(ecode)
		}
	}()

	prog := filepath.Base(os.Args[0])
	args := os.Args[1:]

	Features["complete"] = IntrinsicComplete
	Features["help"] = IntrinsicHelp

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, xos.Termination...)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	flag.CommandLine.Init(prog, flag.ContinueOnError)
	v, ok := Features[os.Args[0]]
	if ok {
		flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	} else if v, ok = Features[prog]; ok {
	} else if len(args) == 0 {
		v = IntrinsicHelp
	} else {
		v = Features
	}

	err = Do(ctx, None, v, args)
}

func Do(
	ctx context.Context,
	preempt Preemption,
	feature any,
	args []string,
) error {
	switch t := feature.(type) {
	case func(context.Context, []string) error:
		switch preempt {
		case None:
			return t(ctx, args)
		case Complete:
			return nil
		case Help:
			return t(ctx, xslices.Prepend(args, "-h"))
		}
	case func(context.Context, bool, []string) error:
		switch preempt {
		case None:
			return t(ctx, false, args)
		case Complete:
			return t(ctx, true, args)
		case Help:
			return t(ctx, false, xslices.Prepend(args, "-h"))
		}
	case error:
		switch preempt {
		case None:
			return t
		case Complete:
			return nil
		case Help:
			return t
			flag.CommandLine.Usage()
		}
	case map[string]any:
		return ImpliedSelect(ctx, preempt, t, args)
	case func() (any, error):
		switch preempt {
		case None:
			return ImpliedPrintResults(ctx, t, args)
		case Complete:
			return nil
		case Help:
			return ImpliedPrintResults(ctx, t,
				xslices.Prepend(args, "-h"))
		}
	default:
		switch preempt {
		case None:
			return ImpliedPrintOrScanObject(ctx, false, t, args)
		case Complete:
			return ImpliedPrintOrScanObject(ctx, true, t, args)
		case Help:
			return ImpliedPrintOrScanObject(ctx, false, t,
				xslices.Prepend(args, "-h"))
		}
	}
	return nil
}

func PrintMatchingKeys(m map[string]any, prefixes ...string) {
	keys := xmaps.Match(m, strings.HasPrefix, prefixes...)
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k)
	}
}

func Reselect(ctx context.Context, args []string) error {
	return ImpliedSelect(ctx, None, Features, args)
}
