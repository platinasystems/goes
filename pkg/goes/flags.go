// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"errors"
	"flag"

	"github.com/platinasystems/goes/v2/pkg/context/parameter"
	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
)

func ContextFlags(ctx context.Context) *flag.FlagSet {
	return parameter.Value(ctx, &flag.CommandLine)
}

func FlagsContext(ctx context.Context, flags *flag.FlagSet) context.Context {
	return parameter.Context(ctx, &flag.CommandLine, flags)
}

// If ContextHelp(ctx), immediately return (ctx, nil); otherwise, silently
// parse args with ContextFlags(ctx) and if that returns ErrHelp, return...
//
//	(HelpContext(ctx, true), nil)
//
// ... so that a subsequent ContextHelp returns true.
func ParseFlagsContext(ctx context.Context, args []string) (
	context.Context, error,
) {
	if ContextHelp(ctx) {
		return ctx, nil
	}
	err := flagset.SilentParse(ContextFlags(ctx), args)
	if errors.Is(err, flag.ErrHelp) {
		return HelpContext(ctx, true), nil
	}
	return ctx, err
}

// Return the value of the named flag, or the zero value if unavailable or
// non-convertible.
func SearchContextFlags[T comparable](ctx context.Context, name string) T {
	return flagset.Search[T](ContextFlags(ctx), name)
}

func SprintContextFlags(ctx context.Context) string {
	return flagset.Sprint(ContextFlags(ctx))
}
