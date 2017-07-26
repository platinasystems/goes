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
	"syscall"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
)

const (
	Name    = "ls"
	Apropos = "list directory contents"
	Usage   = "ls [OPTION]... [FILE]..."
	Man     = `
DESCRIPTION
	List information about the FILEs (the current directory by default).

OPTIONS

	-C	list entries by columns (default)
	-l	long listing format
	-1	list one entry per line`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	PathSeparatorString = string([]byte{os.PathSeparator})
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var err error
	var ls func([]os.FileInfo) error

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

	files := make([]os.FileInfo, 0, 128)
	dirs := make([]os.FileInfo, 0, 128)
	defer func() {
		files = files[:0]
		dirs = dirs[:0]
	}()
	if len(args) == 0 {
		dir, err := os.Stat(".")
		if err != nil {
			return err
		}
		dirs = append(dirs, dir)
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
						dirs = append(dirs, fi)
					} else {
						files = append(files, fi)
					}
				}
			}
		}
	}
	if len(files) > 0 {
		err = ls(files)
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
			fmt.Print(dir.Name(), ":\n")
		}
		files = files[:0]
		fis, err := ioutil.ReadDir(dir.Name())
		if err != nil {
			return err
		}
		for _, fi := range fis {
			files = append(files, fi)
		}
		err = ls(files)
	}
	return err
}

// List one file per line.
func one(files []os.FileInfo) error {
	for _, file := range files {
		fmt.Println(file.Name())
	}
	return nil
}

func long(files []os.FileInfo) error {
	for _, fi := range files {
		name := fi.Name()
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
				fmt.Printf(": %v\n", name, err)
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
func tabulate(files []os.FileInfo) error {
	names := make([]string, 0, len(files))
	for i := 0; i < len(files); i++ {
		names = append(names, files[i].Name())
	}

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
