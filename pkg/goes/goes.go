// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"embed"
	"fmt"
	"io"
	"maps"
	"os"
	"os/signal"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Do(ctx context.Context, subsys any, args []string) error {
	switch t := subsys.(type) {
	case error:
		return t
	case map[string]any:
		return Select(RootContext(ctx, t), args)
	case func(context.Context, []string) error:
		return t(ctx, args)
	case embed.FS:
		return ImplicitPrintEmbedFS(ctx, t, args)
	case []byte:
		return ImplicitPrintOrReadBytes(ctx, t, args)
	case string:
		return ImplicitPrintString(ctx, t, args)
	case fmt.Stringer:
		return ImplicitPrintStringer(ctx, t, args)
	case func() ([]byte, error):
		return ImplicitPrintBytesResult(ctx, t, args)
	case func() string:
		return ImplicitPrintStringResult(ctx, t, args)
	case JSONer:
		return ImplicitPrintOrUnmarshalJSON(ctx, t, args)
	case Texter:
		return ImplicitPrintOrUnmarshalText(ctx, t, args)
	default:
		return ImplicitPrintOrScanObject(ctx, subsys, args)
	}
}

// Execute subsystem with an interruptible context.
func Exec(ctx context.Context, subsys any, args []string) {
	ctx, stop := signal.NotifyContext(ctx, termination.Signals...)
	defer stop()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := Do(ctx, subsys, args)
	if err == nil {
		return
	}
	if IsUsage(err) {
		FprintErr(os.Stdout, err)
		return
	}
	if !IsMarked(err) && !egress.IsMarked(err) {
		err = Mark(ctx, err)
	}
	FprintErr(os.Stderr, err)
	os.Exit(1)
}

func FprintErr(w io.Writer, err error) {
	const nl = "\n"
	es := err.Error()
	if !strings.HasSuffix(es, nl) {
		es += nl
	}
	w.Write([]byte(es))
}

// Merge IntegralCommands into Root then Exec Select of Root with args.
func Main() {
	maps.Copy(Root, IntegralCommands)
	maps.Copy(RootDaemons(), IntegralDaemons)
	maps.Copy(RootShows(), IntegralShow)
	Exec(context.Background(), Select, os.Args[1:])
}

// Select subsystem from args.
func Select(ctx context.Context, args []string) (err error) {
	root := ContextRoot(ctx)
	if len(root) == 0 {
		err = ErrEmpty
	} else if len(args) == 0 {
		if ContextComplete(ctx) {
			err = complete.Last(args, root)
		} else {
			err = IntegralHelp(ctx, args)
		}
	} else if v, ok := root[args[0]]; ok {
		ctx = AppendBranchContext(ctx, args[0])
		err = Do(ctx, v, args[1:])
	} else if len(args) == 1 && ContextComplete(ctx) {
		err = complete.Last(args, root)
	} else if len(ContextBranch(ctx)) == 1 {
		err = IntegralCommand(ctx, args)
	} else {
		ctx = AppendBranchContext(ctx, args[0])
		err = ErrNotFound
	}
	if err != nil && !IsMarked(err) && !IsUsage(err) &&
		!egress.IsMarked(err) {
		err = Mark(ctx, err)
	}
	return
}
