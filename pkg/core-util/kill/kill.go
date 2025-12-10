// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package kill

import (
	"context"
	"flag"
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xtext"
)

const KillUsage = `
usage: {{.Name}} [flags] <pid>...
Signal process.

  -s <name>
  -<name>
  -<number>
	Signal name or number. (default "` + KillDefaultSignalName + `")

  -l	List signal names.
`

var (
	KillErrIncompletePid    = xerrors.Incomplete("pid")
	KillErrIncompleteSignal = xerrors.Incomplete("signal")
)

func Kill(ctx context.Context, args []string) error {
	var ok bool
	sig := KillDefaultSignal

	xflag.TemplateUsage(KillUsage)
	if len(args) == 0 {
		return KillErrIncompletePid
	}
	if args[0] == "-h" {
		flag.CommandLine.Usage()
		return flag.ErrHelp
	}
	if args[0] == "-l" {
		killList()
		return nil
	}
	if args[0] == "-s" {
		if len(args) < 2 {
			return KillErrIncompleteSignal
		}
		if sig, ok = killSignalNamed[strings.ToLower(args[1])]; !ok {
			return xerrors.Invalid(args[1])
		}
		args = args[2:]
	} else if strings.HasPrefix(args[0], "-") {
		s := strings.TrimPrefix(args[0], "-")
		if len(s) == 0 {
			return KillErrIncompleteSignal
		}
		if unicode.IsNumber([]rune(s)[0]) {
			u, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				return err
			}
			sig = Signal(u)
		} else if sig, ok = killSignalNamed[strings.ToLower(s)]; !ok {
			return xerrors.Invalid(s)
		}
		args = args[1:]
	}
	if len(args) == 0 {
		return KillErrIncompletePid
	}
	for _, arg := range args {
		if pid, err := strconv.ParseInt(arg, 10, 0); err != nil {
			return err
		} else if proc, err := os.FindProcess(int(pid)); err != nil {
			return err
		} else if err = proc.Signal(sig); err != nil {
			return err
		}
	}
	return nil
}

func killList() {
	text := make([]string, len(killSignalNamed))
	for i, s := range slices.SortedFunc(maps.Keys(killSignalNamed),
		func(i, j string) int {
			return int(killSignalNamed[i]) - int(killSignalNamed[j])
		}) {
		text[i] = fmt.Sprintf("%2d) %s", killSignalNamed[s], s)
	}
	pgsz := xtext.GetPageSize(os.Stdout)
	fmt.Print(xtext.TopDownLeftRight{pgsz, text})
}
