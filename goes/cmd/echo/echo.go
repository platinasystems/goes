// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package echo

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
)

const (
	Name    = "echo"
	Apropos = "print a line of text"
	Usage   = "echo [-n] [STRING]..."
	Man     = `
DESCRIPTION
	Echo the STRING(s) to standard output.

	-n     do not output the trailing newline`
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
	flag, args := flags.New(args, "-n")
	s := strings.Join(args, " ")
	if flag.ByName["-n"] {
		fmt.Print(s)
	} else {
		fmt.Println(s)
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
