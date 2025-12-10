// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cp

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const CpUsage = `
usage: {{.Name}} [flags] [-T] <source> <destination>
       {{.Name}} [flags] <source>... <directory>
       {{.Name}} [flags] -t <directory> <source>...
Copy source to destination, or multiple source(s) to a directory.

{{flags .}}`

var (
	Cp_T,
	Cp_v bool
	Cp_t string
)

var CpFlags = xflag.Labels{
	{"T", "Treat destination as a normal file.", &Cp_T},
	{"t", "Copy all sources to directory.", &Cp_t},
	{"v", "Verbose logging.", &Cp_v},
}

func Cp(ctx context.Context, args []string) error {
	xflag.TemplateUsage(CpUsage)
	err := CpFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	args = flag.Args()

	if Cp_T {
		switch len(args) {
		case 0:
			err = xerrors.Incomplete("source")
		case 1:
			err = xerrors.Incomplete("destination")
		case 2:
			err = cp(args[0], args[1])
		default:
			err = xerrors.Invalid(args[2:])
		}
	} else if len(Cp_t) > 0 {
		if len(args) == 0 {
			err = xerrors.Incomplete("source")
		} else if err = chkdir(Cp_t); err == nil {
			for _, source := range args {
				dest := filepath.Join(Cp_t, filepath.Base(source))
				if err = cp(source, dest); err != nil {
					break
				}
			}
		}
	} else {
		switch len(args) {
		case 0:
			err = xerrors.Incomplete("source")
		case 1:
			var wd string
			if wd, err = os.Getwd(); err == nil {
				dest := filepath.Join(wd, filepath.Base(args[0]))
				err = cp(args[0], dest)
			}
		case 2:
			err = cp(args[0], args[1])
		default:
			dir := args[len(args)-1]
			if err = chkdir(dir); err == nil {
				for _, t := range args[:len(args)-1] {
					b := filepath.Base(t)
					l := filepath.Join(dir, b)
					if err = cp(t, l); err != nil {
						break
					}
				}
			}
		}
	}
	return err
}

func cp(source, dest string) error {
	var w io.WriteCloser
	var i fs.FileInfo
	r, err := os.Open(source)
	if err != nil {
		return err
	}
	defer r.Close()
	if i, err = os.Stat(dest); err == nil {
		w, err = os.OpenFile(dest, os.O_WRONLY|os.O_TRUNC, i.Mode())
	} else if !errors.Is(err, fs.ErrNotExist) {
	} else if i, err = r.Stat(); err == nil {
		w, err = os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, i.Mode())
	}
	if err == nil {
		defer w.Close()
		if _, err = io.Copy(w, r); err == nil && Cp_v {
			fmt.Println("Copied", source, "to", dest)
		}
	}
	return err
}

func chkdir(name string) error {
	fi, err := os.Stat(name)
	if err == nil {
		if !fi.IsDir() {
			err = xerrors.Invalid(name)
		}
	}
	return err
}
