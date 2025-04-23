// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tool

import (
	"context"
	"flag"
	"fmt"
	"net"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

func Nslookup(ctx context.Context, args []string) error {
	var sep string

	xflag.TemplateUsage(`
usage: {{.Name}} [-|<name>|<address> [<server>]]
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); len(args) == 0 {
		return xerrors.FIXME("implement interactive mode")
	}

	if unicode.IsDigit(([]rune(args[0]))[0]) {
		if names, err := net.LookupAddr(args[0]); err == nil {
			for _, name := range names {
				fmt.Print(sep, name)
				sep = ", "
			}
			fmt.Println()
		}
	} else if ips, err := net.LookupIP(args[0]); err == nil {
		for _, ip := range ips {
			fmt.Print(sep, ip)
			sep = ", "
		}
		fmt.Println()
	}
	return nil
}
