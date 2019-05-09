// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bang

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "!" }

func (Command) Usage() string {
	return "! COMMAND [ARGS]... [&]"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "run an external command",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Sh-bang!

	Command executes in background if last argument ends with '&'.
	The standard i/o redirections apply.`,
	}
}

func (Command) Kind() cmd.Kind { return cmd.DontFork }

func (Command) Main(args ...string) error {
	var background bool

	if len(args) == 0 {
		return fmt.Errorf("COMMAND: missing")
	}

	if n := len(args); args[n-1] == "&" {
		background = true
		args = args[:n-1]
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if background {
		go func() {
			err := cmd.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, cmd.Args[0], ": ", err)
			}
		}()
		return nil
	} else {
		return cmd.Run()
	}
}
