// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package iocmd

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "iocmd"
	Apropos = "read/write the CPU's I/O ports"
	Usage   = "iocmd [[-r] | -w] IO-ADDRESS [-d DATA] [-m MODE]"
	Man     = `
DESCRIPTION
	The iocmd command reads and writes the CPU's I/O ports.
	  -r to read from ioport, default
	  -w to write from ioport
	     IO-ADDRESS is a hex value
	  -d DATA is a hex value
	  -m MODE is one of:
	    b (read byte data, default)
	    w (read word data)
	    l (read long data)`
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

func (cmd) Main(args ...string) (err error) {
	flag, args := flags.New(args, "-r", "-w")
	parm, args := parms.New(args, "-d", "-m")
	if len(args) == 0 {
		return fmt.Errorf("IO-ADDRESS: missing")
	}
	if parm.ByName["-d"] == "" {
		parm.ByName["-d"] = "0x0"
	}

	var a, d, w uint64
	if a, err = strconv.ParseUint(args[0], 0, 32); err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}
	if d, err = strconv.ParseUint(parm.ByName["-d"], 0, 32); err != nil {
		return fmt.Errorf("%s: %v", parm.ByName["-d"], err)
	}
	switch parm.ByName["-m"] {
	case "w":
		w = 2
	case "l":
		w = 4
	default:
		w = 1
	}

	if err = io_reg_rw(flag.ByName["-w"], a, d, w); err != nil {
		return err
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

func io_reg_rw(write bool, addr uint64, dat uint64, wid uint64) (err error) {
	if err = syscall.Iopl(3); err != nil {
		return err
	}
	if err = syscall.Ioperm(int(addr), 1, 1); err != nil {
		return err
	}

	n := 0
	b := make([]byte, wid)
	if !write {
		f, err := os.Open("/dev/port")
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = f.Seek(int64(addr), 0); err != nil {
			return err
		}
		if _, err = f.Read(b); err != nil {
			return err
		}
		fmt.Println("Read value =", b)
		f.Close()
	} else {
		b[0] = byte(dat & 0xff) //FIXME add 16, 32-bit
		f, err := os.OpenFile("/dev/port", os.O_WRONLY, 0755)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err = f.Seek(int64(addr), 0); err != nil {
			return err
		}
		if n, err = f.Write(b); err != nil {
			return err
		}
		f.Sync()
		fmt.Println("Wrote", n, "byte(s)")
		f.Close()
	}

	if err = syscall.Ioperm(int(addr), 1, 0); err != nil {
		return err
	}
	if err = syscall.Iopl(0); err != nil {
		return err
	}
	return nil
}
