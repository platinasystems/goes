// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package chmod

import (
	"context"
	"flag"
	"os"
	"strconv"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const ChmodUsage = `
usage: {{.Name}} <mode> <file>
Change file mode to given octal value.
`

func Chmod(ctx context.Context, args []string) error {
	xflag.TemplateUsage(ChmodUsage)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()
	switch len(args) {
	case 0:
		return xerrors.Incomplete("mode")
	case 1:
		return xerrors.Incomplete("file")
	}

	u64, err := strconv.ParseUint(args[0], 8, 32)
	if err != nil {
		return xerrors.Label(err, "mode")
	}

	mode := os.FileMode(uint32(u64))

	for _, fn := range args[1:] {
		if err = os.Chmod(fn, mode); err != nil {
			break
		}
	}
	return err
}
