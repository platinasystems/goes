// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/platinasystems/go/goes/lang"
)

type Command struct{}

func (Command) String() string { return "exec" }

func (Command) Usage() string { return "exec COMMAND..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "execute a file",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Replace the current goes process with the given command.`,
	}
}

func (Command) Main(args ...string) error {
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
