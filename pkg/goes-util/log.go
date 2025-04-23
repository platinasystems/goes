// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xos"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func Log(ctx context.Context, complete bool, args []string) error {
	xflag.TemplateUsage(`
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

	cmd := exec.CommandContext(ctx, xprogram.Path(), args...)
	cmd.Args[0] = xprogram.MainName()
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

	var wg sync.WaitGroup
	cp := func(w io.Writer, r io.Reader) {
		var b []byte
		defer wg.Done()
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			b = sc.Bytes()
			b = bytes.TrimSpace(b)
			if len(b) > 0 {
				w.Write(b)
			}
		}
	}

	wg.Add(1)
	go cp(outLog, outPipe)
	wg.Add(1)
	go cp(errLog, errPipe)

	if err = cmd.Start(); err != nil {
		return err
	}

	err = cmd.Wait()
	wg.Wait()
	return err
}
