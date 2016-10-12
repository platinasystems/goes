// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package cat

import (
	"io"
	"os"
	"syscall"

	"github.com/platinasystems/go/url"
)

type cat struct{}

func New() cat { return cat{} }

func (cat) String() string { return "cat" }
func (cat) Usage() string  { return "cat [FILE]..." }

func (p *cat) Main(args ...string) error {
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

func (cat) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print concatenated files",
	}
}

func (cat) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	cat - print concatenates files

SYNOPSIS
	cat [OPTION]... [FILE]...

DESCRIPTION
	Concatenate FILE(s), or standard input, to standard output.

	With no FILE, or when FILE is -, read standard input.

EXAMPLES
	cat f - g
		Output f's contents, then standard input, then g's contents.

		cat	Copy standard input to standard output.`,
	}
}
