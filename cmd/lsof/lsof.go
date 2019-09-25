// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lsof

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "lsof" }

func (Command) Usage() string {
	return "lsof [OPTION]..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "list open files",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Print open files.`,
	}
}

func (Command) Main(args ...string) error {
	pidlist := []int{}

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	fns, err := ioutil.ReadDir("/proc")
	if err != nil {
		return fmt.Errorf("Error reading /proc: %s", err)
	}

	for _, fn := range fns {
		pid, err := strconv.ParseUint(fn.Name(), 10, 64)
		if err != nil {
			continue
		}
		pidlist = append(pidlist, int(pid))
	}
	sort.Ints(pidlist)
	for _, pid := range pidlist {
		pidStr := strconv.Itoa(pid)
		f := "/proc/" + pidStr
		showFile(pidStr, "cwd", f+"/cwd", nil)
		showFile(pidStr, "txt", f+"/exe", nil)
		showFile(pidStr, "rtd", f+"/root", nil)
		showFileSorted(pidStr, f+"/fd")
		showFileDir(pidStr, "map", f+"/map_files", false)
	}
	return nil
}

func showFile(pid, kind, link string, dups map[string]bool) (err error) {
	cmdline := ""
	cl, err := ioutil.ReadFile("/proc/" + pid + "/cmdline")
	if err == nil {
		cls := strings.Split(string(cl), "\x00")
		if len(cls) > 0 {
			clf := strings.Fields(cls[0])
			if len(clf) > 0 {
				cmdline = filepath.Base(clf[0])
			}
		}
	}

	file, err := os.Readlink(link)
	if err != nil {
		return
	}
	if dups != nil {
		if dups[file] {
			return
		}
		dups[file] = true
	}
	fmt.Printf("%9s %s %s %s\n", cmdline, pid, kind, file)
	return nil
}

func showFileSorted(pid, dir string) (err error) {
	fdlist := []int{}
	fns, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("Error reading %s: %s", dir, err)
	}

	for _, fn := range fns {
		fd, err := strconv.ParseUint(fn.Name(), 10, 64)
		if err != nil {
			continue
		}
		fdlist = append(fdlist, int(fd))
	}
	sort.Ints(fdlist)
	for _, fd := range fdlist {
		fn := strconv.Itoa(fd)
		showFile(pid, fn, dir+"/"+fn, nil)
	}
	return nil
}

func showFileDir(pid, kind, dir string, dups bool) (err error) {
	var dupList map[string]bool
	if !dups {
		dupList = make(map[string]bool)
	}
	fns, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("Error reading %s: %s", dir, err)
	}

	for _, fn := range fns {
		k := kind
		if k == "" {
			k = fn.Name()
		}
		showFile(pid, k, dir+"/"+fn.Name(), dupList)
	}
	return nil
}
