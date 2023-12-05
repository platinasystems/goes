// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/context/flagctx"
	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
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
		Path:     pathctx.StringIn(ctx),
		Synopsis: synopsis,
		Flag:     flagctx.StringIn(ctx),
	}
}

func route(ctx context.Context, args ...string) error {
	path := pathctx.Parameter.In(ctx)
	cmd := path[len(path)-1]
	fs := flag.NewSilentFlagSet(cmd)
	ctx = flagctx.Parameter.With(ctx, fs)
	if flag.Search[bool]("complete") {
		return complete.Last(args, fs)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return usage.Error(RouteUsageTemplate[1:],
			RouteUsageData(ctx, routeSynopsis[cmd]))
	}
	args = fs.Args()

	switch cmd {
	case "add":
		err = netrt.Add(ctx, fs)
	case "change":
		err = netrt.Change(ctx, fs)
	case "delete":
		err = netrt.Delete(ctx, fs)
	case "flush":
		err = netrt.Flush(ctx, fs)
	case "get", "show":
		if nrt, gerr := netrt.Get(ctx, fs); gerr == nil {
			fmt.Fprint(wctx.Parameter.In(ctx), nrt)
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
