// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ls

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/platinasystems/go/goes/internal/flags"
)

const Name = "ls"

var PathSeparatorString = string([]byte{os.PathSeparator})

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [OPTION]... [FILE]..." }

func (cmd) Main(args ...string) error {
	var err error
	var ls func(string, []string) error

	flag, args := flags.New(args, "-l", "-C", "-1")

	switch {
	case flag["-1"]:
		ls = one
	case flag["-l"]:
		ls = long
	case flag["-C"]:
		fallthrough
	default:
		ls = tabulate
	}

	files := make([]string, 0, 128)
	dirs := make([]string, 0, 128)
	defer func() {
		files = files[:0]
		dirs = dirs[:0]
	}()
	if len(args) == 0 {
		dirs = append(dirs, ".")
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
						dirs = append(dirs, name)
					} else {
						files = append(files, name)
					}
				}
			}
		}
	}
	if len(files) > 0 {
		err = ls("", files)
		if len(dirs) > 0 {
			fmt.Println()
		}
	}
	shouldPrintDirName := len(dirs) > 1 || len(files) > 0
	for i, dir := range dirs {
		if shouldPrintDirName {
			if i > 0 {
				fmt.Println()
			}
			fmt.Print(dir, ":\n")
		}
		files = files[:0]
		fis, err := ioutil.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, fi := range fis {
			files = append(files, filepath.Join(dir, fi.Name()))
		}
		err = ls(dir+PathSeparatorString, files)
	}
	return err
}

// List one file per line.
func one(prefix string, names []string) error {
	sort.Strings(names)
	for _, name := range names {
		fmt.Println(strings.TrimPrefix(name, prefix))
	}
	return nil
}

func long(prefix string, names []string) error {
	sort.Strings(names)
	for _, name := range names {
		fi, err := os.Stat(name)
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
			if err != nil {
				return err
			}
			fmt.Printf("%12s %2d %4d %4d %10d %s %s -> %s\n",
				fi.Mode().String(),
				st.Nlink,
				st.Uid,
				st.Gid,
				fi.Size(),
				fi.ModTime().Format("Jan 02 15:04"),
				fi.Name(),
				lnk)
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
func tabulate(prefix string, names []string) error {
	sort.Strings(names)
	columns := 80
	if env := os.Getenv("COLUMNS"); len(env) > 0 {
		fmt.Sscan(env, &columns)
		if columns < 24 {
			columns = 24
		}
	}

	width := 8
	for i := 0; i < len(names); {
		names[i] = strings.TrimPrefix(names[i], prefix)
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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "list directory contents",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	ls - list directory contents

SYNOPSIS
	ls [OPTION]... [FILE]...

DESCRIPTION

	List information about the FILEs (the current directory by default).

OPTIONS

	-C	list entries by columns (default)
	-l	long listing format
	-1	list one entry per line`,
	}
}
