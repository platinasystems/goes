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

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/lang"
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

type Interface interface {
	Apropos() lang.Alt
	Kind() goes.Kind
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Kind() goes.Kind   { return goes.DontFork }

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
