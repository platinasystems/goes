// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package rmmod

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/flags"
)

type rmmod struct{}

func New() rmmod { return rmmod{} }

func (rmmod) String() string { return "rmmod" }
func (rmmod) Usage() string  { return "rmmod [OPTION]... MODULE..." }

func (rmmod) Main(args ...string) error {
	flag, args := flags.New(args, "-f", "-q", "-v")
	if len(args) == 0 {
		return fmt.Errorf("MODULE: missing")
	}
	u := 0
	if flag["-f"] {
		u |= syscall.O_TRUNC
	}
	for _, name := range args {
		bp, err := syscall.BytePtrFromString(name)
		if err != nil {
			return err
		}
		_, _, e := syscall.RawSyscall(syscall.SYS_DELETE_MODULE,
			uintptr(unsafe.Pointer(bp)), uintptr(u), 0)
		if e != 0 {
			if !flag["-q"] {
				return fmt.Errorf("%v", e)
			}
		} else if flag["-v"] {
			fmt.Println("Removed", name)
		}
	}
	return nil
}

func (rmmod) Complete(args ...string) (c []string) {
	if len(args) == 0 {
		return
	}
	if args[0] == "rmmod" {
		args = args[1:]
	}
	if len(args) == 0 {
		return
	}

	f, err := os.Open("/proc/modules")
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		x := strings.Fields(line)
		if strings.HasPrefix(x[0], args[len(args)-1]) {
			c = append(c, x[0])
		}
	}
	return
}

func (rmmod) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "remove a module from the Linux Kernel",
	}
}

func (rmmod) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	rmmod - remove a module from the Linux Kernel

SYNOPSIS
	rmmod [OPTION]... MODULE

DESCRIPTION
	Remove the named MODULE from the Linux Kernel.
	(MODULE must support unloading)

OPTIONS
	-v	verbose
	-f	force
	-q	silently ignore errors`,
	}
}
