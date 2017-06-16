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

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "!"
	Apropos = "run an external command"
	Usage   = "! COMMAND [ARGS]... [&]"
	Man     = `
DESCRIPTION
	Sh-bang!

	Command executes in background if last argument ends with '&'.
	The standard i/o redirections apply.`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Kind() cmd.Kind    { return cmd.DontFork }

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
		signal.Ignore(syscall.SIGINT)
		// FIXME how to kill subprocess with SIGINT
		return cmd.Run()
	}
}

func (Command) Man() lang.Alt  { return man }
func (Command) String() string { return Name }
func (Command) Usage() string  { return Usage }
