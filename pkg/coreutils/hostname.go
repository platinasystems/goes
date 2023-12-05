// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/flagctx"
	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

const HostnameUsageTemplate = `
usage: {{.Path}} [<options>] [<name>]
Set or print system host name.
{{.Flag}}`

func HostnameUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: pathctx.StringIn(ctx),
		Flag: flagctx.StringIn(ctx),
	}
}

func Hostname(ctx context.Context, args ...string) error {
	fs := flag.NewSilentFlagSet("hostname")
	ctx = flagctx.Parameter.With(ctx, fs)
	dFlag := fs.Bool("d", false, "only print domain")
	fFlag := fs.Bool("f", true, "print fully qualified domain name (FQDN)")
	sFlag := fs.Bool("s", false, "print name w/o domain")
	if flag.Search[bool]("complete") {
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return usage.Error(HostnameUsageTemplate[1:],
			HostnameUsageData(ctx))
	}
	args = fs.Args()
	if len(args) > 0 {
		return host.Rename(args[0])
	}
	hn := host.Name()
	if dot := strings.Index(hn, "."); dot > 0 {
		switch {
		case *dFlag:
			hn = hn[dot+1:]
		case *fFlag:
			// default
		case *sFlag:
			hn = hn[:dot]
		}
	}
	fmt.Fprintln(wctx.Parameter.In(ctx), hn)
	return ctx.Err()
}
