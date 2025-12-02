// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nslookup

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
)

func Nslookup(ctx context.Context, args []string) error {
	var sep string

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [-|<name>|<address> [<server>]]
`)
	err := xflag.Labels{
		xmain.ConfigFlag,
		xdnsdoh.ConfigFlag,
	}.Define()
	err = flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); len(args) == 0 {
		return xerrors.FIXME("implement interactive mode")
	}
	arg0 := args[0]

	resolver, err := xdnsdoh.Resolver()
	if err != nil {
		return err
	}

	if net.ParseIP(arg0) == nil {
		var ips []net.IP
		ips, err = resolver.LookupIP(ctx, "ip", xdnsdoh.FQDN(arg0))
		if err == nil {
			for _, ip := range ips {
				fmt.Print(sep, ip)
				sep = ", "
			}
			fmt.Println()
		}
	} else {
		var names []string
		names, err = resolver.LookupAddr(ctx, arg0)
		if err == nil {
			for _, name := range names {
				fmt.Print(sep, name)
				sep = ", "
			}
			fmt.Println()
		}
	}
	return err
}
