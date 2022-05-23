// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hostname

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	fs := flag.NewFlagSet("hostname", flag.ContinueOnError)
	dFlag := fs.Bool("d", false, "only print domain")
	fFlag := fs.Bool("f", true, "print fully qualified domain name (FQDN)")
	sFlag := fs.Bool("s", false, "print name w/o domain")
	fs.Usage = func() {
		path.Usage(w, "[<options>] [<name>]\n",
			"Set or print system host name.\n",
			fs,
		)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if path.HasComplete() {
		complete.Last(w, args, fs)
		return nil
	}
	if path.HasHelp() {
		fs.Usage()
		return nil
	}
	if args = fs.Args(); len(args) > 0 {
		return host.Set(args[0])
	}
	hn, err := host.Name.Value()
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
	fmt.Fprintln(w, hn)
	return ctx.Err()
}
