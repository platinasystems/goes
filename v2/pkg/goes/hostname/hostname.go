// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hostname

import (
	"context"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	dFlag := fs.Bool("d", false, "only print domain")
	fFlag := fs.Bool("f", true, "print fully qualified domain name (FQDN)")
	sFlag := fs.Bool("s", false, "print name w/o domain")
	usage := func() error {
		return template.Must(template.New("usage").Parse(`
usage: {{.Command}} [<options>] [<name>]\n",
Set or print system host name.
{{ print .Flags}}`[1:])).Execute(w, struct {
			Command string
			Flags   flags.Flags
		}{
			strings.Join(path, " "),
			fs,
		})
	}
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return usage()
	}
	err := fs.Parse(args)
	if err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
	}
	if args = fs.Args(); len(args) > 0 {
		return host.Set(args[0])
	}
	hn, err := host.Name.ValErr()
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
