// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

var Route = map[string]any{
	"add":     route,
	"change":  route,
	"delete":  route,
	"flush":   route,
	"get":     route,
	"monitor": route,
}

var routeSynopsis = map[string]string{
	"add":     "Add a route.",
	"change":  "Change aspects of a route (such as its gateway).",
	"delete":  "Delete a specific route.",
	"flush":   "Remove all routes.",
	"get":     "Lookup and display the route for a destination.",
	"monitor": "Continuously report route changes.",
}

const RouteUsageTemplate = `
usage: {{.Path}} [<option>]... [<addr|prefix> [<gateway>] [<mask>]]
{{.Synopsis}}
{{.Flag}}`

func RouteUsageData(ctx context.Context, synopsis string) any {
	return struct{ Path, Synopsis, Flag string }{
		Path:     strings.Join(ctxparm.Strings.In(ctx), " "),
		Synopsis: synopsis,
		Flag:     ctxparm.SprintFlagsIn(ctx),
	}
}

func route(ctx context.Context, args ...string) error {
	path := ctxparm.Strings.In(ctx)
	cmd := path[len(path)-1]
	flags := usage.NewFlags(cmd)
	ctx = ctxparm.Flags.With(ctx, flags)
	afinet := flags.Bool("4", false, "Address hint.")
	flags.BoolVar(afinet, "inet", false, "aka -4.")
	afinet6 := flags.Bool("6", false, "Address hint.")
	flags.BoolVar(afinet6, "inet6", false, "aka -6.")
	iface := flags.Bool("interface", false, "Instead of next-hop.")
	flags.BoolVar(iface, "iface", false, "aka. -interface")
	flags.String("dst", "", "Instead 1st position arg.")
	flags.String("gateway", "", "Instead 2nd position arg.")
	flags.String("mask", "", "Instead 3rd position arg or 1st /<suffix>.")
	flags.Int("prefixlen", -1, "Instead of 1st arg /<suffix>.")
	routeFlagsGOOS(flags)
	if *complete.Help {
		return complete.Last(args, flags)
	}
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *usage.Help {
		return usage.Error(RouteUsageTemplate[1:],
			RouteUsageData(ctx, routeSynopsis[cmd]))
	}
	switch cmd {
	case "add":
		err = netrt.Add(ctx)
	case "change":
		err = netrt.Change(ctx)
	case "delete":
		err = netrt.Delete(ctx)
	case "flush":
		err = netrt.Flush(ctx)
	case "get", "show":
		if nrt, gerr := netrt.Get(ctx); gerr == nil {
			fmt.Fprint(ctxparm.Writer.In(ctx), nrt)
		} else {
			err = gerr
		}
	case "monitor":
		if nrts, monerr := netrt.Monitor(ctx); monerr == nil {
			_ = nrts //FIXME
		} else {
			err = monerr
		}
	}
	return err
}
