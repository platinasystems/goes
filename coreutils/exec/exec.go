// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package exec

import (
	"fmt"
	"os"
	os_exec "os/exec"
	"syscall"
)

type exec struct{}

func New() exec { return exec{} }

func (exec) String() string { return "exec" }
func (exec) Usage() string  { return "exec COMMAND..." }

func (exec) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("COMMAND: missing")
	}

	path, err := os_exec.LookPath(args[0])
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

func (exec) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "execute a file",
	}
}

func (exec) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	exec - execute a file"

SYNOPSIS
	exec COMMAND [ARG]...

DESCRIPTION
	Replace the current goes process with the given command.`,
	}
}
