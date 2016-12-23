// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package insmod

import (
	"fmt"
	"io/ioutil"
	"strings"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/url"
)

const Name = "insmod"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }

func (cmd) Usage() string {
	return Name + " [OPTION]... FILE [NAME[=VAL[,VAL]]]..."
}

func (cmd) Main(args ...string) error {
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
	if flag["-v"] {
		fmt.Println("Installed", args[0])
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "insert a module into the Linux Kernel",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	insmod - insert a module into the Linux Kernel

SYNOPSIS
	insmod [OPTION]... FILE [NAME[=VALUE[,VALUE]]...

OPTIONS
	-v	verbose
	-f	force`,
	}
}
