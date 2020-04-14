// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package echo

import (
	"fmt"
	"strings"

	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "echo" }

func (Command) Usage() string { return "echo [-n] [STRING]..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print a line of text",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Echo the STRING(s) to standard output.

	-n     do not output the trailing newline`,
	}
}

func (Command) Main(args ...string) error {
	flag, args := flags.New(args, "-n")
	s := strings.Join(args, " ")
	if flag.ByName["-n"] {
		fmt.Print(s)
	} else {
		fmt.Println(s)
	}
	return nil
}
