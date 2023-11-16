// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
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

func route(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join .Path " "}} [<option>]... [<addr|prefix> [<gateway>] [<mask>]]
{{.Synopsis}}

Options{{SprintDefault .Flags}}`
	cmd := path[len(path)-1]
	fs := netrt.FlagSet(cmd)
	if complete.Parameter.Value(ctx) {
		style.Completions(args, fs)
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if help.Wanted(ctx, fs) {
		return style.Usage(usage, struct {
			Path     []string
			Synopsis string
			Flags    *flag.FlagSet
		}{path, routeSynopsis[cmd], fs})
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
			fmt.Fprint(w, nrt)
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
