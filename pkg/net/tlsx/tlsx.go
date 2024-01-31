// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"errors"
	"flag"

	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

var Commands = map[string]any{
	// "approve":   IPC,
	// "deny":      IPC,
	// "exec":      Rexec,
	"generate": map[string]any{
		"self":      GenerateSelf,
		"signature": GenerateSignature,
	},
	// "subscribe": Subscribe,
}

func Command(ctx context.Context, args []string) error {
	return tlsx(goes.RootContext(ctx, Commands), args, `
usage: {{branch .}} [<option>]... <command> [<arg>]...
Execute TLSX command.

Options{{flags .}}
Commands
{{root .}}`)
}

func Daemon(ctx context.Context, args []string) error {
	return FIXME
}

var Shows = map[string]any{
	"self":          ShowSelf,
	"signature":     ShowSignature,
	"subscribers":   ShowSubs,
	"subscriptions": ShowSubs,
	// "registry":      IPC,
	// "subscribers":   SubscribersNames,
	// "subscriptions": SubscriptionsNames,
	// "tenants":       IPC,
}

func Show(ctx context.Context, args []string) error {
	return tlsx(goes.RootContext(ctx, Shows), args, `
usage: {{branch .}} [<option>]... <object>...
Print TLSX object.

Options{{flags .}}
Objects
{{root .}}`)
}

func tlsx(ctx context.Context, args []string, usage string) error {
	opts := options()
	ctx = goes.FlagsContext(ctx, opts)
	if goes.ContextComplete(ctx) {
		return complete.Last(args, goes.ContextRoot(ctx))
	}
	err := flagset.SilentParse(opts, args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			ctx = goes.HelpContext(ctx, true)
		} else {
			return err
		}
	}
	if goes.ContextHelp(ctx) && len(args) == 0 {
		return goes.Usage(ctx, usage)
	}
	args = opts.Args()
	return goes.Select(ctx, args)
}

func options() *flag.FlagSet {
	opts := new(flag.FlagSet)
	OptionalSelfFileName = opts.String("self", DefaultSelfFileName(),
		"PEM encoded TLS certificate file.")
	OptionalSignatureFileName = opts.String("signature",
		DefaultSignatureFileName(),
		"PEM encoded private key file.")
	OptionalSubscribersFileName = opts.String("subscribers",
		DefaultSubscribersFileName(),
		"File containing PEM encoded TLS client certificates.")
	OptionalSubscriptionsFileName = opts.String("subscriptions",
		DefaultSubscriptionsFileName(),
		"File containing PEM encoded TLS peer certificates.")
	return opts
}

func Match(nameOrSKI string) (x *X509) {
	if Self.IsMatch(nameOrSKI) {
		x = &Self.X509
	} else if x = Subscriptions.Match(nameOrSKI); x == nil {
		x = Subscribers.Match(nameOrSKI)
	}
	return
}
