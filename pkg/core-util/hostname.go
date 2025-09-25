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

var Hostname_d = xflag.New[bool]("d", `Only print domain name.`, nil)
var Hostname_f = xflag.New[bool]("f", `Print fully qualified domain name.`,
	func() bool { return true })
var Hostname_s = xflag.New[bool]("s", `Print name w/o domain.`, nil)

var HostnameFlags = []xflag.Definer{
	Hostname_d,
	Hostname_f,
	Hostname_s,
}

func Hostname(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [name]
Set or print system host name.

{{flags .}}`)

	for _, f := range HostnameFlags {
		f.Define()
	}

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
		case Hostname_d.Value():
			hn = hn[dot+1:]
		case Hostname_f.Value():
			// default
		case Hostname_s.Value():
			hn = hn[:dot]
		}
	}
	fmt.Println(hn)
	return ctx.Err()
}
