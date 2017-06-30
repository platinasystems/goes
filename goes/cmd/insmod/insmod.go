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

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
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

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

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
