// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

const Name = "exec"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " COMMAND..." }

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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "execute a file",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	exec - execute a file"

SYNOPSIS
	exec COMMAND [ARG]...

DESCRIPTION
	Replace the current goes process with the given command.`,
	}
}
