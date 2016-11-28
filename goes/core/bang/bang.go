// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package bang

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const Name = "!"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " COMMAND [ARGS]..." }

func (cmd) Main(args ...string) error {
	var background bool

	if len(args) == 0 {
		return fmt.Errorf("COMMAND: missing")
	}

	if n := len(args); args[n-1] == "&" {
		background = true
		args = args[:n-1]
	}

	cmd := exec.Command(args[0], args[1:]...)

	if background {
		go func() {
			err := cmd.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, cmd.Args[0], ": ", err)
			}
		}()
		return nil
	} else {
		signal.Ignore(syscall.SIGINT)
		// FIXME how to kill subprocess with SIGINT
		return cmd.Run()
	}
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "run an external command",
	}
}

func (*cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	! - run an external command"

SYNOPSIS
	! COMMAND [ARG]... [&]

DESCRIPTION
	Sh-bang!

	Command executes in background if last argument ends with '&'.
	The standard i/o redirections apply.`,
	}
}
