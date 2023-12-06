// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const HostnameUsageTemplate = `
usage: {{.Path}} [<options>] [<name>]
Set or print system host name.
{{.Flag}}`

func HostnameUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: strings.Join(ctxparm.Strings.In(ctx), " "),
		Flag: ctxparm.SprintFlagsIn(ctx),
	}
}

func Hostname(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	flags := usage.NewFlags("hostname")
	ctx = ctxparm.Flags.With(ctx, flags)
	dFlag := flags.Bool("d", false, "only print domain")
	fFlag := flags.Bool("f", true, "print fully qualified domain name (FQDN)")
	sFlag := flags.Bool("s", false, "print name w/o domain")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *usage.Help {
		return usage.Error(HostnameUsageTemplate[1:],
			HostnameUsageData(ctx))
	}
	args = flags.Args()
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
	fmt.Fprintln(ctxparm.Writer.In(ctx), hn)
	return ctx.Err()
}
