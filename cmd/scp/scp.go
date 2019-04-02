// Copyright © 2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package scp

import (
	"fmt"

	"github.com/platinasystems/flags"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) Kind() cmd.Kind { return cmd.NoCLIFlags }

func (Command) String() string { return "scp" }

func (Command) Usage() string { return "scp [-t] [-f] DIR" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "securely copy a file",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	scp implemenets the server side of the SCP protocol.

OPTIONS
	-f	Specifies source mode
	-t	Specifies sink mode`,
	}
}

func (Command) Main(args ...string) error {
	flag, args := flags.New(args, "-f", "-t")
	i := 0
	if flag.ByName["-f"] {
		i++
	}
	if flag.ByName["-t"] {
		i++
	}
	if i != 1 {
		return fmt.Errorf("scp requires one of -f or -t")
	}
	if len(args) == 0 {
		return fmt.Errorf("scp requires a directory or filename")
	}
	if len(args) > 1 {
		return fmt.Errorf("unexpected %v", args)
	}
	if flag.ByName["-f"] {
		return sourceMode(args...)
	}
	return sinkMode(args...)
}

func sourceMode(args ...string) error {
	return fmt.Errorf("No source mode yet")
}

func sinkMode(args ...string) error {
	return fmt.Errorf("No sink mode yet")
}
