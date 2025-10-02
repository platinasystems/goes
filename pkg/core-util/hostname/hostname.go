// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hostname

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

	var d, s bool
	f := true
	err := xflag.Labels{
		{"d", "Only print domain name.", &d},
		{"f", "Print fully qualified domain name.", &f},
		{"s", "Print name w/o domain.", &s},
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if args = flag.Args(); len(args) > 0 {
		return Sethostname(args[0])
	}

	hn, err := os.Hostname()
	if err != nil {
		return err
	}
	if dot := strings.Index(hn, "."); dot > 0 {
		if d {
			hn = hn[dot+1:]
		} else if s {
			hn = hn[:dot]
		}
	}
	fmt.Println(hn)
	return ctx.Err()
}
