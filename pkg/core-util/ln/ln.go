// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ln

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const LnUsage = `
usage: {{.Name}} [flags] -t <directory> <target>...
       {{.Name}} [flags] -T <target> <name>
       {{.Name}} [flags] <target> <link>
       {{.Name}} [flags] <target>... <directory>
       {{.Name}} [flags] <target-path>
Create a link to the specified target(s).

{{flags .}}`

var (
	Ln_b,
	Ln_f,
	Ln_s,
	Ln_T,
	Ln_v bool

	Ln_S = "~"

	Ln_t string
)

var LnFlags = xflag.Labels{
	{"S", "Backup file suffix.", &Ln_S},
	{"T", "Name argument is a normal file.", &Ln_T},
	{"b", "Backup target (if exists) before making link.", &Ln_b},
	{"f", "Remove existing destinations.", &Ln_f},
	{"s", "Make symlinks instead of hardlinks.", &Ln_s},
	{"t", "Specify directory in which to create the links.", &Ln_t},
	{"v", "Verbose logging.", &Ln_v},
}

func Ln(ctx context.Context, args []string) error {
	xflag.TemplateUsage(LnUsage)
	err := LnFlags.Define()
	if err != nil {
		return err
	}
	err = flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	args = flag.Args()

	if Ln_T {
		switch len(args) {
		case 0:
			err = xerrors.Incomplete("target")
		case 1:
			err = xerrors.Incomplete("name")
		case 2:
			err = ln(args[0], args[1])
		default:
			err = xerrors.Invalid(args[2:])
		}
	} else if len(Ln_t) > 0 {
		switch len(args) {
		case 0:
			err = xerrors.Incomplete("directory")
		case 1:
			err = xerrors.Incomplete("target(s)")
		default:
			if err = chkdir(Ln_t); err == nil {
				for _, t := range args {
					base := filepath.Base(t)
					l := filepath.Join(Ln_t, base)
					if err = ln(t, l); err != nil {
						break
					}
				}
			}
		}
	} else {
		switch len(args) {
		case 0:
			err = xerrors.Incomplete("target")
		case 1:
			if filepath.Dir(args[0]) == "." {
				err = xerrors.Incomplete("link")
			} else {
				var wd string
				if wd, err = os.Getwd(); err == nil {
					base := filepath.Base(args[0])
					l := filepath.Join(wd, base)
					err = ln(args[0], l)
				}
			}
		case 2:
			err = ln(args[0], args[1])
		default:
			dir := args[len(args)-1]
			if err = chkdir(dir); err == nil {
				for _, t := range args[:len(args)-1] {
					base := filepath.Base(t)
					l := filepath.Join(dir, base)
					if err = ln(t, l); err != nil {
						break
					}
				}
			}
		}
	}
	return err
}

func ln(target, link string) error {
	var err error
	if _, err = os.Stat(link); err == nil {
		if !Ln_f {
			return xerrors.Label(fs.ErrExist, link)
		}
		if Ln_b {
			bu := link + Ln_S
			if err = os.Link(link, bu); err != nil {
				return err
			}
		}
		if err = (os.Remove(link)); err != nil {
			return err
		}
		if Ln_v {
			fmt.Println("Removed", link)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	linked := "Linked"
	if Ln_s {
		linked = "Symlinked"
		err = os.Symlink(target, link)
	} else {
		err = os.Link(target, link)
	}
	if err == nil && Ln_v {
		fmt.Println(linked, target, "to", link)
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
