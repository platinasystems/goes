// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ls

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
)

var PathSeparatorString = string([]byte{os.PathSeparator})

func New() Command { return Command{} }

type Command struct{}

func (Command) String() string { return "ls" }

func (Command) Usage() string {
	return "ls [OPTION]... [FILE]..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "list directory contents",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	List information about the FILEs (the current directory by default).

OPTIONS

	-C	list entries by columns (default)
	-l	long listing format
	-1	list one entry per line`,
	}
}

func (Command) Main(args ...string) error {
	var err error
	var ls func([]string) error
	var fns, dns []string

	flag, args := flags.New(args, "-l", "-C", "-1")

	switch {
	case flag.ByName["-1"]:
		ls = one
	case flag.ByName["-l"]:
		ls = long
	case flag.ByName["-C"]:
		fallthrough
	default:
		ls = tabulate
	}

	if len(args) == 0 {
		_, err := os.Stat(".")
		if err != nil {
			return err
		}
		dns = append(dns, ".")
	} else {
		for _, pat := range args {
			globs, err := filepath.Glob(pat)
			if err != nil {
				return err
			}
			if len(globs) == 0 {
				return fmt.Errorf("%s: %v", pat,
					syscall.ENOENT)
			}
			for _, name := range globs {
				fi, err := os.Stat(name)
				if err == nil {
					if fi.IsDir() {
						dns = append(dns, name)
					} else {
						fns = append(fns, name)
					}
				} else {
					return fmt.Errorf("%s: %v", name, err)
				}
			}
		}
	}
	if len(fns) > 0 {
		err = ls(fns)
		if len(dns) > 0 {
			fmt.Println()
		}
	}
	shouldPrintDirName := len(dns) > 1 || len(fns) > 0
	for i, dn := range dns {
		if shouldPrintDirName {
			if i > 0 {
				fmt.Println()
			}
			fmt.Print(dn, ":\n")
		}
		fns = fns[:0]
		fis, err := ioutil.ReadDir(dn)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			fns = append(fns, filepath.Join(dn, fi.Name()))
		}
		err = ls(fns)
	}
	return err
}

// List one file per line.
func one(names []string) error {
	for _, name := range names {
		fmt.Println(filepath.Base(name))
	}
	return nil
}

func long(names []string) error {
	for _, name := range names {
		fi, err := os.Lstat(name)
		if err != nil {
			return err
		}
		st := fi.Sys().(*syscall.Stat_t)
		switch st.Mode & syscall.S_IFMT {
		case syscall.S_IFBLK, syscall.S_IFCHR:
			maj := uint64(st.Rdev / 256)
			min := uint64(st.Rdev % 256)
			fmt.Printf("%12s %2d %4d %4d %4d, %4d %s %s\n",
				fi.Mode().String(),
				st.Nlink,
				st.Uid,
				st.Gid,
				maj,
				min,
				fi.ModTime().Format("Jan 02 15:04"),
				fi.Name())
		case syscall.S_IFLNK:
			lnk, err := os.Readlink(name)
			fmt.Printf("%12s %2d %4d %4d %10d %s %s -> %s",
				fi.Mode().String(),
				st.Nlink,
				st.Uid,
				st.Gid,
				fi.Size(),
				fi.ModTime().Format("Jan 02 15:04"),
				fi.Name(),
				lnk)
			if err == nil {
				fmt.Printf("\n")
			} else {
				fmt.Printf("%s: %s\n", name, err)
			}
		default:
			fmt.Printf("%12s %2d %4d %4d %10d %s %s\n",
				fi.Mode().String(),
				st.Nlink,
				st.Uid,
				st.Gid,
				fi.Size(),
				fi.ModTime().Format("Jan 02 15:04"),
				fi.Name())
		}
	}
	return nil
}

// Arrange file names in tabular form with names longer than 24 runes printed
// first on separate lines.
func tabulate(names []string) error {
	for i, name := range names {
		names[i] = filepath.Base(name)
	}
	sort.Strings(names)
	columns := 80
	if env := os.Getenv("COLUMNS"); len(env) > 0 {
		fmt.Sscan(env, &columns)
		if columns < 80 {
			columns = 80
		}
	}

	width := 8
	for i := 0; i < len(names); {
		if n := len(names[i]); n > 24 {
			fmt.Println(names[i])
			if i < len(names)-1 {
				copy(names[i:], names[i+1:])
			}
			names = names[:len(names)-1]
		} else {
			i++
			if n >= width {
				width += 8
			}
		}
	}

	tcols := (columns - 1) / width
	tlines := len(names) / tcols
	if tlines*tcols < len(names) {
		tlines += 1
	}

	for line := 0; line < tlines; line++ {
		for tcol := 0; tcol < tcols; tcol++ {
			k := line + (tcol * tlines)
			if k < len(names) {
				fmt.Printf("%-*s", width, names[k])
			}
		}
		fmt.Println()
	}
	return nil
}
