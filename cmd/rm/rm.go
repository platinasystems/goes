// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "rm" }

func (Command) Usage() string {
	return "rm [OPTION]... FILE..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "remove files or directories",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Remove named files.  By default, it does not remove directories.

OPTIONS
	-f	ignore nonexistent files and arguments, never prompt

	-r	remove directories and their contents recursively

	-d	remove empty directories

	-v	verbose

	Include a relative of full path prefix to remove a file whose name
	starts with a '-'; for example:

              rm ./-f
              rm ./-v`,
	}
}

func (Command) Main(args ...string) error {
	flag, args := flags.New(args, "-d", "-f", "-r", "-v")
	for _, fn := range args {
		fi, err := os.Stat(fn)
		if err != nil {
			if os.IsNotExist(err) {
				if flag.ByName["-f"] {
					continue
				}
			}
			if err != nil {
				return err
			}
		}
		if fi.IsDir() {
			err = rmdir(fn, flag)
			if err != nil {
				return err
			}
		} else {
			if err = os.Remove(fn); err != nil {
				return err
			}
			if flag.ByName["-v"] {
				fmt.Println("Removed", fn)
			}
		}
	}
	return nil
}

func rmdir(dn string, flag *flags.Flags) error {
	fis, err := ioutil.ReadDir(dn)
	if err != nil {
		return err
	}
	if len(fis) > 0 {
		if !flag.ByName["-r"] {
			return fmt.Errorf("%s: isn't empty", dn)
		}
		for _, fi := range fis {
			fn := filepath.Join(dn, fi.Name())
			if fi.IsDir() {
				rmdir(fn, flag)
			} else {
				err = os.Remove(fn)
				if err != nil {
					return err
				}
				if flag.ByName["-v"] {
					fmt.Println("Removed", fi.Name())
				}
			}
		}
	}

	if !flag.ByName["-d"] && !flag.ByName["-r"] {
		return fmt.Errorf("%s: is a directory", dn)
	}
	if err = os.Remove(dn); err != nil {
		return err
	}
	if flag.ByName["-v"] {
		fmt.Println("Removed", dn)
	}
	return nil
}
