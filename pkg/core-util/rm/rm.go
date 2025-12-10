// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rm

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const RmUsage = `
usage: {{.Name}} [flags] <name>...
Remove named files.  By default, it does not remove directories.

{{flags .}}`

var Rm_d, Rm_f, Rm_r, Rm_v bool

var RmFlags = xflag.Labels{
	{"d", "Remove empty directories.", &Rm_d},
	{"f", "Ignore nonexistent files and arguments, never prompt.", &Rm_f},
	{"r", "Remove directories and their contents recursively.", &Rm_r},
	{"v", "Verbose.", &Rm_v},
}

func Rm(ctx context.Context, args []string) error {
	xflag.TemplateUsage(RmUsage)
	err := RmFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	args = flag.Args()
	if len(args) == 0 {
		return xerrors.Incomplete("name")
	}

	for _, name := range args {
		var fi os.FileInfo
		if fi, err = os.Stat(name); err != nil {
			if os.IsNotExist(err) {
				if Rm_f {
					continue
				}
			}
			if err != nil {
				break
			}
		}
		if fi.IsDir() {
			if err = rmdir(name); err != nil {
				break
			}
		} else if err = os.Remove(name); err != nil {
			break
		} else if Rm_v {
			fmt.Println("Removed", name)
		}
	}
	return err
}

func rmdir(name string) error {
	fis, err := ioutil.ReadDir(name)
	if err != nil {
		return err
	}
	if len(fis) > 0 {
		if !Rm_r {
			return fmt.Errorf("%s: isn't empty", name)
		}
		for _, fi := range fis {
			pn := filepath.Join(name, fi.Name())
			if fi.IsDir() {
				rmdir(pn)
			} else if err = os.Remove(pn); err != nil {
				return err
			} else if Rm_v {
				fmt.Println("Removed", pn)
			}
		}
	}
	if !Rm_d && !Rm_r {
		return fmt.Errorf("%s: is a directory", name)
	}
	if err = os.Remove(name); err != nil {
		return err
	}
	if Rm_v {
		fmt.Println("Removed", name)
	}
	return nil
}
