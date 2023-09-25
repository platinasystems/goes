// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ln

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

type ln struct {
	Flags *flags.Flags
	Parms *parms.Parms
}

func (Command) String() string { return "ln" }

func (Command) Usage() string {
	return `
	ln [OPTION]... -t DIRECTORY TARGET...
	ln [OPTION]... -T TARGET LINK
	ln [OPTION]... TARGET LINK
	ln [OPTION]... TARGET... DIRECTORY
	ln [OPTION]... TARGET`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "make links between files",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Create a link LINK or DIR/TARGET to the specified TARGET(s)

OPTIONS
	-s	Make symlinks instead of hardlinks
	-f	Remove existing destinations
	-backup	Make a backup of the target (if exists) before link operation
	-suffix SUFFIX
		Use suffix instead of ~ when making backup files
	-v	verbose`,
	}
}

func (Command) Main(args ...string) error {
	var err error
	var ln ln

	ln.Flags, args = flags.New(args, "-f", "-s", "-b", "-T", "-v")
	ln.Parms, args = parms.New(args, "-S", "-t")

	if ln.Flags.ByName["-T"] {
		switch len(args) {
		case 0:
			return fmt.Errorf("TARGET LINK: missing")
		case 1:
			return fmt.Errorf("LINK: missing")
		case 2:
		default:
			return fmt.Errorf("%v:unexpected", args[2:])
		}
		err = ln.ln(args[0], args[1])
	} else if dir := ln.Parms.ByName["-t"]; len(dir) > 0 {
		if len(args) == 0 {
			return fmt.Errorf("TARGET...: missing")
		}
		err = valid(dir)
		if err == nil {
			for _, t := range args {
				l := filepath.Join(dir, filepath.Base(t))
				err = ln.ln(t, l)
				if err != nil {
					break
				}
			}
		}
	} else {
		switch len(args) {
		case 0:
			return fmt.Errorf("TARGET: missing")
		case 1:
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			l := filepath.Join(wd, filepath.Base(args[0]))
			err = ln.ln(args[0], l)
		case 2:
			err = ln.ln(args[0], args[1])
		default:
			dir := args[len(args)-1]
			err = valid(dir)
			if err == nil {
				for _, t := range args[:len(args)-1] {
					b := filepath.Base(t)
					l := filepath.Join(dir, b)
					err = ln.ln(t, l)
					if err != nil {
						break
					}
				}
			}
		}
	}
	return err
}

func (ln *ln) ln(target, link string) error {
	var err error
	if _, err = os.Stat(link); err == nil {
		if !ln.Flags.ByName["-f"] {
			return fmt.Errorf("%s: exists", link)
		}
		if ln.Flags.ByName["-b"] {
			bu := link + ln.Parms.ByName["-S"]
			if err = os.Link(link, bu); err != nil {
				return err
			}
		}
		if err = (os.Remove(link)); err != nil {
			return err
		}
		if ln.Flags.ByName["-v"] {
			fmt.Println("Removed", link)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	linked := "Linked"
	if ln.Flags.ByName["-s"] {
		linked = "Symlinked"
		err = os.Symlink(target, link)
	} else {
		err = os.Link(target, link)
	}
	if err != nil {
		return err
	}
	if ln.Flags.ByName["-v"] {
		fmt.Println(linked, target, "to", link)
	}
	return nil
}

func valid(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s: isn't a directory", dir)
	}
	return nil
}
