// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package core_util

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

func Hostname(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] [name]
Set or print system host name.

{{flags .}}`)

	dFlag := flag.Bool("d", false, "only print domain")
	fFlag := flag.Bool("f", true,
		"print fully qualified domain name (FQDN)")
	sFlag := flag.Bool("s", false, "print name w/o domain")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); len(args) > 0 {
		return Sethostname(args[0])
	}

	hn, err := os.Hostname()
	if err != nil {
		return err
	}
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
	fmt.Println(hn)
	return ctx.Err()
}
