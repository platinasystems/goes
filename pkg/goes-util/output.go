// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	"flag"
	"io"
	"os"
	"os/exec"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func Output(ctx context.Context, complete bool, args []string) error {
	var a, e, t bool
	var m uint

	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <file> <feature> [args]
Execute feature with output written to file.

{{flags .}}`)

	xflag.Define(&a, "a", "Append <file> instead of truncate.")
	xflag.Define(&e, "e",
		"Write or tee Stderr to <file> instead of Stdout.")
	xflag.Define(&m, "m", "Output file mode (default 0666).")
	xflag.Define(&t, "t", "Tee to <file> and stdout.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		if len(args) > 1 {
			return goes.IntrinsicComplete(ctx, args[1:])
		}
		// FIXME complete <file>
		return nil
	} else if len(args) == 0 {
		return xerrors.Incomplete("file")
	} else if len(args) == 1 {
		return xerrors.Incomplete("feature")
	}

	oflags := os.O_RDWR | os.O_CREATE
	if a {
		oflags |= os.O_APPEND
	} else {
		oflags |= os.O_TRUNC
	}

	mode := os.FileMode(0666)
	if m != 0 {
		mode = os.FileMode(m)
	}

	f, err := os.OpenFile(args[0], oflags, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd := exec.CommandContext(ctx, xprogram.Path(), args[1:]...)
	cmd.Args[0] = xmain.PackageName()
	cmd.Stdin = os.Stdin

	if t {
		var r io.ReadCloser
		var w io.Writer
		if e {
			cmd.Stdout = os.Stdout
			r, err = cmd.StderrPipe()
			if err != nil {
				return err
			}
			w = os.Stderr
		} else {
			cmd.Stderr = os.Stderr
			r, err = cmd.StdoutPipe()
			if err != nil {
				return err
			}
			w = os.Stdout
		}
		defer r.Close()
		go io.Copy(io.MultiWriter(w, f), r)
	} else if e {
		cmd.Stdout = os.Stdout
		cmd.Stderr = f
	} else {
		cmd.Stdout = f
		cmd.Stderr = os.Stderr
	}
	return cmd.Run()
}
