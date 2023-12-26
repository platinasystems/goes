// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

const HostnameUsage = `
usage: {{branch .}} [<options>] [<name>]
Set or print system host name.
{{flags .}}`

func Hostname(ctx context.Context, args []string) error {
	if goes.ContextComplete(ctx) {
		return nil
	}
	var flags flag.FlagSet
	ctx = goes.FlagsContext(ctx, &flags)
	dFlag := flags.Bool("d", false, "only print domain")
	fFlag := flags.Bool("f", true,
		"print fully qualified domain name (FQDN)")
	sFlag := flags.Bool("s", false, "print name w/o domain")
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, HostnameUsage)
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
	fmt.Fprintln(goes.ContextStdout(ctx), hn)
	return ctx.Err()
}
