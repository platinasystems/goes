// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package echo

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/internal/flags"
)

const Name = "echo"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [-n] [STRING]..." }

func (cmd) Main(args ...string) error {
	flag, args := flags.New(args, "-n")
	s := strings.Join(args, " ")
	if flag["-n"] {
		fmt.Print(s)
	} else {
		fmt.Println(s)
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "print a line of text",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	echo - print a line of text

SYNOPSIS
	echo [-n] [STRING]...

DESCRIPTION
	Echo the STRING(s) to standard output.

	-n     do not output the trailing newline`,
	}
}
