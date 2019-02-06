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

	"github.com/ramr/go-reaper"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/goes/internal/assert"
	"github.com/platinasystems/goes/internal/parms"
	"github.com/platinasystems/goes/internal/prog"
)

func New() *Command { return new(Command) }

type Command struct {
	// Machines may use Hook to run something before redisd and other
	// daemons.
	Hook func() error

	// Machines may use ConfHook to run something after all daemons start
	// and before source of start command script.
	ConfHook func() error

	// GPIO init hook for machines than need it
	ConfGpioHook func() error

	g *goes.Goes
}

func (*Command) String() string { return "start" }

func (*Command) Usage() string {
	return "start [-start=URL] [REDIS OPTIONS]..."
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "start this goes machine",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
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

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	parm, args := parms.New(args, "-start", "-stop")

	err := assert.Root()
	if err != nil {
		return err
	}
	if prog.Name() != prog.Install && prog.Base() != "init" {
		return fmt.Errorf("use `%s start`", prog.Install)
	}
	if c.Hook != nil {
		if err = c.Hook(); err != nil {
			return err
		}
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

	start := parm.ByName["-start"]
	if len(start) == 0 {
		if _, xerr := os.Stat("/etc/goes/start"); xerr == nil {
			start = "/etc/goes/start"
		}
	}

	if c.ConfGpioHook != nil {
		if err = c.ConfGpioHook(); err != nil {
			return err
		}
	}

	if len(start) > 0 {
		if c.ConfHook != nil {
			if err = c.ConfHook(); err != nil {
				return err
			}
		}
		err = c.g.Main("source", start)
		if err != nil {
			return err
		}
	}

	if os.Getpid() != 1 {
		return nil
	}

	go reaper.Reap()

	go daemons.Wait()

	for {
		if _, found := c.g.ByName["login"]; found {
			if err = c.run("login"); err != nil {
				fmt.Fprintln(os.Stderr, "login:", err)
				time.Sleep(3 * time.Second)
				continue
			}
		}
		err = c.run("cli")
		if err == io.EOF {
			err = nil
		}
		if err != nil {
			fmt.Fprint(os.Stderr, prog.Base(), ": ", err, "\n")
		}
	}

}

func (c *Command) run(args ...string) error {
	x := c.g.Fork(args...)
	x.Stdin = os.Stdin
	x.Stdout = os.Stdout
	x.Stderr = os.Stderr
	x.Dir = "/"
	return x.Run()
}
