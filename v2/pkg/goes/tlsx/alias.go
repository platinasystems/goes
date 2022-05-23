// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/alias"
)

func Alias(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	fs := flag.NewFlagSet("alias", flag.ContinueOnError)
	fs.Usage = func() {
		path.Usage(w, "[<name> [<subject-key-id>]]\n",
			"Add or print subject-key-id association.")
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	args = fs.Args()
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() {
		fs.Usage()
		return nil
	}
	switch len(args) {
	case 0:
		alias.Range(func(aka, ski string) bool {
			fmt.Fprint(w, aka, ": ", ski, "\n")
			return true
		})
	case 1:
		if ski, ok := alias.Load(args[0]); ok {
			fmt.Fprintln(w, ski)
		}
	case 2:
		err = alias.Store(args[0], args[1])
	default:
		err = fmt.Errorf("unexpected arg(s)", args[2:])
	}
	return err
}
