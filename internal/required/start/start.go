// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package start provides the named command that runs a redis server followed
// by all of the configured daemons. If the PID is 1, start doesn't return;
// instead, it iterates and command shell.
package start

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/assert"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/go/internal/sockfile"
)

const Name = "start"

// Machines may use Hook to run something before redisd and other daemons.
var Hook = func() error { return nil }

// Machines may use ConfHook to run something after all daemons start and
// before source of start command script.
var ConfHook = func() error { return nil }

func New() *cmd { return new(cmd) }

type cmd goes.ByName

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return "start [OPTION]..." }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	parm, args := parms.New(args, "-start", "-stop")

	err := assert.Root()
	if err != nil {
		return err
	}
	if prog.Name() != prog.Install && prog.Base() != "init" {
		return fmt.Errorf("use `%s start`", prog.Install)
	}
	_, err = os.Stat(sockfile.Path("redisd"))
	if err == nil {
		return fmt.Errorf("already started")
	}
	err = Hook()
	if err != nil {
		return err
	}
	daemons := exec.Command(prog.Name(), args...)
	daemons.Args[0] = "goes-daemons"
	daemons.Stdin = nil
	daemons.Stdout = nil
	daemons.Stderr = nil
	daemons.Dir = "/"
	daemons.Env = []string{
		"PATH=" + prog.Path(),
		"TERM=linux",
	}
	daemons.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
		Pgid:   0,
	}
	err = daemons.Start()
	if err != nil {
		return err
	}

	start := parm["-start"]
	if len(start) == 0 {
		if _, xerr := os.Stat("/etc/goes/start"); xerr == nil {
			start = "/etc/goes/start"
		}
	}
	if len(start) > 0 {
		err = ConfHook()
		if err != nil {
			return err
		}
		err = goes.ByName(*c).Main("source", start)
		if err != nil {
			return err
		}
	}

	if os.Getpid() != 1 {
		return nil
	}

	go daemons.Wait()

	for {
		if _, found := goes.ByName(*c)["login"]; found {
			if err = run("login"); err != nil {
				fmt.Fprintln(os.Stderr, "login:", err)
				time.Sleep(3 * time.Second)
				continue
			}
		}
		err = run("cli")
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			fmt.Fprint(os.Stderr, prog.Base(), ": ", err, "\n")
		}
	}

}

func run(arg0 string) error {
	x := exec.Command(prog.Name())
	x.Args[0] = arg0
	x.Stdin = os.Stdin
	x.Stdout = os.Stdout
	x.Stderr = os.Stderr
	x.Dir = "/"
	return x.Run()
}

func (c *cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "start this goes machine",
	}
}

func (c *cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	start - start this goes machine

SYNOPSIS
	start [-start=URL] [REDIS OPTIONS]...

DESCRIPTION
	Start a redis server followed by the machine and its embedded daemons.

OPTIONS
	-start URL
		Specifies the URL of the machine's configuration script that's
		sourced immediately after start of all daemons.
		default: /etc/goes/start

SEE ALSO
	redisd`,
	}
}
