// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Hostname(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<options>] [<name>]
Set or print system host name.
{{SprintDefault .Flags}}`
	fs := flag.New("hostname")
	dFlag := fs.Bool("d", false, "only print domain")
	fFlag := fs.Bool("f", true, "print fully qualified domain name (FQDN)")
	sFlag := fs.Bool("s", false, "print name w/o domain")
	if complete.Parameter.Value(ctx) {
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if help.Wanted(ctx, fs) {
		return style.Usage(usage, struct {
			Path  []string
			Flags *flag.FlagSet
		}{path, fs})
	}
	args = fs.Args()
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
	fmt.Fprintln(w, hn)
	return ctx.Err()
}
