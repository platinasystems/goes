// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_checkzone

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdb"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var (
	mutable = log.New(os.Stdout, "", log.Lshortfile)
	errata  = xlog.Unmute(mutable)
	verbose = xlog.Unmute(mutable)
)

func NamedCheckZone(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [-flags] {zone} {file | -}
Mimic BIND9's config verification tool.

{{flags .}}`)

	wg := new(sync.WaitGroup)
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		wg.Wait()
	}()

	addCommandLineFlags()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()

	if flags.v {
		if mm := xprogram.MainModule(); mm != nil {
			fmt.Println(mm.Version)
		} else {
			fmt.Println("(unavailable)")
		}
		return nil
	}
	if flags.q {
		verbose = xlog.Mute(verbose)
	}
	switch len(args) {
	case 0:
		return xerrors.Incomplete("zone")
	case 1:
		return xerrors.Incomplete("file")
	}

	zone, fn := args[0], args[1]

	wg.Add(1)
	go xdnsdb.Routine(ctx, wg, verbose)

	if err = xdnsdb.Include(ctx, zone, fn); err != nil {
		return err
	}

	xdnsdb.Dump(os.Stdout)
	return nil

}
