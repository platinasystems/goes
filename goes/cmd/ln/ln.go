// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ln

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "ln"
	Apropos = "make links between files"
	Usage   = `
	ln [OPTION]... -t DIRECTORY TARGET...
	ln [OPTION]... -T TARGET LINK
	ln [OPTION]... TARGET LINK
	ln [OPTION]... TARGET... DIRECTORY
	ln [OPTION]... TARGET`
	Man = `
DESCRIPTION
	Create a link LINK or DIR/TARGET to the specified TARGET(s)

OPTIONS
	-s	Make symlinks instead of hardlinks
	-f	Remove existing destinations
	-backup	Make a backup of the target (if exists) before link operation
	-suffix SUFFIX
		Use suffix instead of ~ when making backup files
	-v	verbose`
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(args ...string) error {
	var err error
	flag, args := flags.New(args, "-f", "-s", "-b", "-T", "-v")
	parm, args := parms.New(args, "-S", "-t")

	if flag["-T"] {
		switch len(args) {
		case 0:
			return fmt.Errorf("TARGET LINK: missing")
		case 1:
			return fmt.Errorf("LINK: missing")
		case 2:
		default:
			return fmt.Errorf("%v:unexpected", args[2:])
		}
		err = ln(args[0], args[1], flag, parm)
	} else if dir := parm["-t"]; len(dir) > 0 {
		if len(args) == 0 {
			return fmt.Errorf("TARGET...: missing")
		}
		err = valid(dir)
		if err == nil {
			for _, t := range args {
				l := filepath.Join(dir, filepath.Base(t))
				err = ln(t, l, flag, parm)
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
			err = ln(args[0], l, flag, parm)
		case 2:
			err = ln(args[0], args[1], flag, parm)
		default:
			dir := args[len(args)-1]
			err = valid(dir)
			if err == nil {
				for _, t := range args[:len(args)-1] {
					b := filepath.Base(t)
					l := filepath.Join(dir, b)
					err = ln(t, l, flag, parm)
					if err != nil {
						break
					}
				}
			}
		}
	}
	return err
}

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

func ln(target, link string, flag flags.Flag, parm parms.Parm) error {
	var err error
	if _, err = os.Stat(link); err == nil {
		if !flag["-f"] {
			return fmt.Errorf("%s: exists", link)
		}
		if flag["-b"] {
			bu := link + parm["-S"]
			if err = os.Link(link, bu); err != nil {
				return err
			}
		}
		if err = (os.Remove(link)); err != nil {
			return err
		}
		if flag["-v"] {
			fmt.Println("Removed", link)
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	linked := "Linked"
	if flag["-s"] {
		linked = "Symlinked"
		err = os.Symlink(target, link)
	} else {
		err = os.Link(target, link)
	}
	if err != nil {
		return err
	}
	if flag["-v"] {
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

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
