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

const (
	Hostname_d_Flag xflag.Xbool = "d Only print domain name."
	Hostname_f_Flag xflag.Xbool = "f Print fully qualified domain name."
	Hostname_s_Flag xflag.Xbool = "s Print name w/o domain."
)

func Hostname(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [name]
Set or print system host name.

{{flags .}}`)

	dFlag := Hostname_d_Flag.Define(false)
	fFlag := Hostname_f_Flag.Define(true)
	sFlag := Hostname_s_Flag.Define(false)

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
