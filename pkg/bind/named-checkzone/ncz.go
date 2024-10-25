// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package named_checkzone

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

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

	zone, fn, f := args[0], args[1], os.Stdin

	if fn != "-" {
		if f, err = os.Open(fn); err != nil {
			return err
		} else {
			defer f.Close()
		}
	}

	db := xdnsdb.New(zone)
	if err = db.Fread(ctx, f); err != nil {
		return err
	}

	var last string
	now := time.Now()
	const indent, lineLen = 48, 80
	db.Range(func(zone, name string, rrs []xdnsdb.RR) bool {
		if len(rrs) == 0 {
			fmt.Print("%-24sEMPTY\n", name+"."+zone)
			return true
		}
		if zone != last {
			fmt.Println("$ORIGIN", zone)
			last = zone
		}
		fmt.Printf("%-24s", name)
		for i, rr := range rrs {
			if i > 0 {
				fmt.Printf("%-24s", "")
			}
			var secs uint
			if rr.TTL.After(now) {
				fsecs := rr.TTL.Sub(now).Seconds()
				secs = uint(math.Round(fsecs))
			} else {
			}
			fmt.Printf("%-8d", secs)
			fmt.Printf("%-8s", "IN")
			fmt.Printf("%-8s", rr.V.Type())
			lns := strings.Split(rr.V.String(), "\n")
			if n := len(lns); n == 0 {
				fmt.Println()
			} else if n == 1 {
				fmt.Println(lns[0])
			} else {
				fmt.Print("(", lns[0])
				for _, s := range lns[1:] {
					fmt.Printf("\n%*s%s", indent+1, "", s)
				}
				fmt.Println(")")
			}
		}
		return true
	})
	return nil
}
