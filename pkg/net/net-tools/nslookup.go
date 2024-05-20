// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"errors"
	"fmt"
	"net"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/goes"
)

func Nslookup(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [-|<name>|<address> [<server>]]
Query Internet name server.
{{flags .}}`

	if goes.ContextComplete(ctx) {
		return nil
	}
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, usage)
	}

	if len(args) == 0 {
		return errors.New("FIXME interactive mode")
	}
	var (
		names []string
		ips   []net.IP
		sep   string
	)
	w := goes.ContextStdout(ctx)
	if unicode.IsDigit(([]rune(args[0]))[0]) {
		if names, err = net.LookupAddr(args[0]); err == nil {
			for _, name := range names {
				fmt.Fprint(w, sep, name)
				sep = ", "
			}
			fmt.Fprintln(w)
		}
	} else if ips, err = net.LookupIP(args[0]); err == nil {
		for _, ip := range ips {
			fmt.Fprint(w, sep, ip)
			sep = ", "
		}
		fmt.Fprintln(w)
	}
	return err
}
