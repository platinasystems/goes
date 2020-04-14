// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rmmod

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "rmmod" }

func (Command) Usage() string {
	return "rmmod [OPTION]... MODULE..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "remove a module from the Linux Kernel",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Remove the named MODULE from the Linux Kernel.
	(MODULE must support unloading)

OPTIONS
	-v	verbose
	-f	force
	-q	silently ignore errors`,
	}
}

func (Command) Complete(args ...string) (c []string) {
	n := len(args)
	f, err := os.Open("/proc/modules")
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		x := strings.Fields(line)
		if n == 0 || len(args[n-1]) == 0 ||
			strings.HasPrefix(x[0], args[n-1]) {
			c = append(c, x[0])
		}
	}
	return
}

func (Command) Main(args ...string) error {
	flag, args := flags.New(args, "-f", "-q", "-v")
	if len(args) == 0 {
		return fmt.Errorf("MODULE: missing")
	}
	u := 0
	if flag.ByName["-f"] {
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
			if !flag.ByName["-q"] {
				return fmt.Errorf("%v", e)
			}
		} else if flag.ByName["-v"] {
			fmt.Println("Removed", name)
		}
	}
	return nil
}
