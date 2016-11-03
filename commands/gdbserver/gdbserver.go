// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package gdbserver

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/platinasystems/go/parms"
)

const UsrBinGdbserver = "/usr/bin/gdbserver"

type gdbserver struct{}

func New() gdbserver { return gdbserver{} }

func (gdbserver) String() string { return "gdbserver" }
func (gdbserver) Usage() string  { return "gdbserver [-p PORT] [PID]" }

func (gdbserver) Main(args ...string) error {
	_, err := os.Stat(UsrBinGdbserver)
	if err != nil {
		return err
	}
	parm, args := parms.New(args, "-p")
	if len(parm["-p"]) == 0 {
		parm["-p"] = "2345"
	}

	pid := strconv.Itoa(os.Getpid())
	if len(args) > 0 {
		pid = args[0]
		args = args[1:]
	}
	if len(args) != 0 {
		fmt.Errorf("%v: unexpected", args)
	}
	c := exec.Command(UsrBinGdbserver, "--attach", "host:"+parm["-p"],
		pid)
	c.Stdin = nil
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err = c.Start(); err != nil {
		return err
	}
	if err = c.Process.Release(); err != nil {
		return err
	}
	time.Sleep(2 * time.Second)
	return nil
}

func (gdbserver) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "run debugger on current process",
	}
}

func (gdbserver) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	gdbserver - run debugger on the current process

SYNOPSIS
	gdbserver [-p PORT] [PID]

DESCRIPTION
	This is wrapper that runs the debugger like this:

		/usr/bin/gdbserver --attach :PORT PID

	The default PORT is 2345.
	The default PID is the current process.

	To connect to this session:

	$ gdb build/MACHINE-ARCH/goes
	(gdb) target remote TARGET:PORT`,
	}
}
