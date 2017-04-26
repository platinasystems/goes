// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/platinasystems/go/internal/goes/lang"
)

const (
	Name    = "exec"
	Apropos = "execute a file"
	Usage   = "exec COMMAND..."
	Man     = `
DESCRIPTION
	Replace the current goes process with the given command.`
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
	if len(args) == 0 {
		return fmt.Errorf("COMMAND: missing")
	}

	path, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
	}
	env := os.Environ()
	err = syscall.Exec(path, args, env)
	if err != nil {
		return fmt.Errorf("%s: %v", args[0], err)
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
