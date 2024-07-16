// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

func Log(ctx context.Context, complete bool, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} <feature> [args]
Execute feature with output piped to syslog, or if GOOS == darwin, oslog.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		return goes.IntrinsicComplete(ctx, args)
	} else if len(args) == 0 {
		return xerrors.Incomplete("feature")
	}

	errLog, err := xos.OpenErrorLog()
	if err != nil {
		return err
	}
	defer errLog.Close()

	defer func() {
		if err != nil {
			fmt.Fprintln(errLog, err)
		}
	}()

	outLog, err := xos.OpenNoticeLog()
	if err != nil {
		return err
	}
	defer outLog.Close()

	cmd := exec.CommandContext(ctx, xos.Program(), args...)
	cmd.Stdin = os.Stdin

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	defer errPipe.Close()

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	defer outPipe.Close()

	if err = cmd.Start(); err != nil {
		return err
	}

	go func(w io.Writer, r io.Reader) {
		for sc := bufio.NewScanner(r); sc.Scan(); {
			w.Write(sc.Bytes())
		}
	}(outLog, outPipe)

	errData, err := io.ReadAll(errPipe)
	errLog.Write(errData)

	err = cmd.Wait()
	return err
}
