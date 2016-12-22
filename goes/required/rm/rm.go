// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/goes/internal/flags"
)

const Name = "rm"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [OPTION]... FILE..." }

func (cmd) Main(args ...string) error {
	flag, args := flags.New(args, "-d", "-f", "-r", "-v")
	for _, fn := range args {
		fi, err := os.Stat(fn)
		if err != nil {
			if os.IsNotExist(err) {
				if flag["-f"] {
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
			if flag["-v"] {
				fmt.Println("Removed", fn)
			}
		}
	}
	return nil
}

func rmdir(dn string, flag flags.Flag) error {
	fis, err := ioutil.ReadDir(dn)
	if err != nil {
		return err
	}
	if len(fis) > 0 {
		if !flag["-r"] {
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
				if flag["-v"] {
					fmt.Println("Removed", fi.Name())
				}
			}
		}
	}

	if !flag["-d"] && !flag["-r"] {
		return fmt.Errorf("%s: is a directory", dn)
	}
	if err = os.Remove(dn); err != nil {
		return err
	}
	if flag["-v"] {
		fmt.Println("Removed", dn)
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "remove files or directories",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	rm - remove files or directories

SYNOPSIS
	rm [OPTION]... FILE...

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
