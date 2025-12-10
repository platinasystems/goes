// Copyright © 2015-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ls

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xtext"
	"golang.org/x/term"
)

const LsUsage = `
usage: {{.Name}} [flags] [<file(s)>]

{{flags .}}`

var (
	Ls_1,
	Ls_C,
	Ls_l bool
)

var LsFlags = xflag.Labels{
	{"1", "List one entry per line.", &Ls_1},
	{"C", "Force multi-column output." +
		" (default for terminal output)", &Ls_C},
	{"l", "Long listing format", &Ls_l},
}

var pgsz xtext.PageSize

func Ls(ctx context.Context, args []string) error {
	xflag.TemplateUsage(LsUsage)
	err := LsFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	args = flag.Args()

	var ls func([]string) error
	var fns, dns []string

	switch {
	case Ls_1:
		ls = ls1
	case Ls_l:
		ls = lsl
	case Ls_C:
		ls = lsC
	default:
		if term.IsTerminal(int(os.Stdout.Fd())) {
			ls = lsC
		} else {
			ls = ls1
		}
	}

	pgsz = xtext.GetPageSize(os.Stdout)

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
func ls1(names []string) error {
	for _, name := range names {
		fmt.Println(filepath.Base(name))
	}
	return nil
}

// Long format.
func lsl(names []string) error {
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
func lsC(names []string) error {
	for i, name := range names {
		names[i] = filepath.Base(name)
	}
	sort.Strings(names)
	_, err := fmt.Print(xtext.TopDownLeftRight{pgsz, names})
	return err
}
