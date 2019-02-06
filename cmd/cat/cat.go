// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cat

import (
	"io"
	"os"
	"syscall"

	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/goes/internal/url"
)

type Command struct{}

func (Command) String() string { return "cat" }

func (Command) Usage() string {
	return "cat [FILE]..."
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print concatenated files",
	}
}
func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Concatenate FILE(s), or standard input, to standard output.

	With no FILE, or when FILE is -, read standard input.

EXAMPLES
	cat f - g
		Output f's contents, then standard input, then g's contents.

	cat
		Copy standard input to standard output.`,
	}
}

func (Command) Main(args ...string) error {
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
