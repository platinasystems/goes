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
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/url"
)

const (
	Name    = "insmod"
	Apropos = "insert a module into the Linux Kernel"
	Usage   = "insmod [OPTION]... FILE [NAME[=VAL[,VAL]]]..."
	Man     = `
OPTIONS
	-v	verbose
	-f	force`
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

func (cmd) Man() lang.Alt  { return man }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
