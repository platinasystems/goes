// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
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
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [name]
Set or print system host name.

{{flags .}}`)

	d := false
	f := true
	s := false
	xflag.Define(&d, "d", `Only print domain name.`)
	xflag.Define(&f, "f", `Print fully qualified domain name.`)
	xflag.Define(&s, "s", `Print name w/o domain.`)

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
		case d:
			hn = hn[dot+1:]
		case f:
			// default
		case s:
			hn = hn[:dot]
		}
	}
	fmt.Println(hn)
	return ctx.Err()
}
