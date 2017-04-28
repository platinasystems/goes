// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cat

import (
	"io"
	"os"
	"syscall"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/url"
)

const (
	Name    = "cat"
	Apropos = "print concatenated files"
	Usage   = "cat [FILE]..."
	Man     = `
DESCRIPTION
	Concatenate FILE(s), or standard input, to standard output.

	With no FILE, or when FILE is -, read standard input.

EXAMPLES
	cat f - g
		Output f's contents, then standard input, then g's contents.

	cat
		Copy standard input to standard output.`
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
	if len(args) == 0 {
		args = []string{"-"}
	}
	for _, fn := range args {
		if fn == "-" {
			io.Copy(os.Stdout, os.Stdin)
		} else {
			f, err := url.Open(fn)
			if err != nil {
				return err
			}
			io.Copy(os.Stdout, f)
			f.Close()
			syscall.Fsync(int(os.Stdout.Fd()))
			if err != nil {
				return err
			}
		}
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
