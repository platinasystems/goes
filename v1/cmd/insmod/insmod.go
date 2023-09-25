// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package insmod

import (
	"fmt"
	"io/ioutil"
	"strings"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/url"
)

type Command struct{}

func (Command) String() string { return "insmod" }

func (Command) Usage() string {
	return "insmod [OPTION]... FILE [NAME[=VAL[,VAL]]]..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "insert a module into the Linux Kernel",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
OPTIONS
	-v	verbose
	-f	force`,
	}
}

func (Command) Main(args ...string) error {
	flag, args := flags.New(args, "-f", "-v")
	if len(args) == 0 {
		return fmt.Errorf("FILE: missing")
	}
	f, err := url.Open(args[0])
	if err != nil {
		return err
	}
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	params := strings.Join(args[1:], " ")
	bp, err := syscall.BytePtrFromString(params)
	if err != nil {
		return err
	}
	_, _, e := syscall.RawSyscall(syscall.SYS_INIT_MODULE,
		uintptr(unsafe.Pointer(&contents[0])),
		uintptr(len(contents)),
		uintptr(unsafe.Pointer(bp)))
	if e != 0 {
		return fmt.Errorf("%v", e)
	}
	if flag.ByName["-v"] {
		fmt.Println("Installed", args[0])
	}
	return nil
}
