// Copyright © 2016-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package stop provides the named command that kills all of the daemons
// associated with this executable.
package stop

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/internal/assert"
	"github.com/platinasystems/goes/internal/kill"
	"github.com/platinasystems/goes/lang"
)

const EtcGoesStop = "/etc/goes/stop"

type Command struct {
	// Machines may use Hook to run something between the kill of all
	// daemons and the removal of the socks and pids directories.
	Hook func() error

	g *goes.Goes
}

func (*Command) String() string { return "stop" }

func (*Command) Usage() string { return "stop [-stop=URL] [SIGNAL]" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "stop this goes machine",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Stop all embedded daemons.

OPTIONS
	-stop URL
		Specifies the URL of the machine's stop script that's
		sourced immediately before killing all daemons.
		default: /etc/goes/start`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Kind() cmd.Kind { return cmd.DontFork }

func (c *Command) Main(args ...string) error {
	parm, args := parms.New(args, "-stop")
	err := assert.Root()
	if err != nil {
		return err
	}
	stop := parm.ByName["-stop"]
	if len(stop) == 0 && haveEtcGoesStop() {
		stop = EtcGoesStop
	}
	if len(stop) > 0 {
		err = c.g.Main("source", stop)
		if err != nil {
			return fmt.Errorf("source %s: %v", stop, err)
		}
	}
	sig := syscall.SIGTERM
	if len(args) == 1 {
		var found bool
		sig, found = sigByName[strings.ToUpper(args[0])]
		if !found {
			return fmt.Errorf("signal: %s: unknown", args[0])
		}
	}
	err = kill.All(sig)
	if c.Hook != nil {
		if t := c.Hook(); err != nil || t != nil {
			if err != nil {
				err = t
			}
			kill.All(syscall.SIGKILL)
		}
	}
	return err
}

func haveEtcGoesStop() bool {
	_, err := os.Stat(EtcGoesStop)
	return err == nil
}

var sigByName = map[string]syscall.Signal{
	"SIGABR":    syscall.SIGABRT,
	"SIGALRM":   syscall.SIGALRM,
	"SIGBUS":    syscall.SIGBUS,
	"SIGCHLD":   syscall.SIGCHLD,
	"SIGCLD":    syscall.SIGCLD,
	"SIGCONT":   syscall.SIGCONT,
	"SIGFPE":    syscall.SIGFPE,
	"SIGHUP":    syscall.SIGHUP,
	"SIGILL":    syscall.SIGILL,
	"SIGINT":    syscall.SIGINT,
	"SIGIO":     syscall.SIGIO,
	"SIGIOT":    syscall.SIGIOT,
	"SIGKILL":   syscall.SIGKILL,
	"SIGPIPE":   syscall.SIGPIPE,
	"SIGPOLL":   syscall.SIGPOLL,
	"SIGPROF":   syscall.SIGPROF,
	"SIGPWR":    syscall.SIGPWR,
	"SIGQUIT":   syscall.SIGQUIT,
	"SIGSEGV":   syscall.SIGSEGV,
	"SIGSTKFLT": syscall.SIGSTKFLT,
	"SIGSTOP":   syscall.SIGSTOP,
	"SIGSYS":    syscall.SIGSYS,
	"SIGTERM":   syscall.SIGTERM,
	"SIGTRAP":   syscall.SIGTRAP,
	"SIGTSTP":   syscall.SIGTSTP,
	"SIGTTIN":   syscall.SIGTTIN,
	"SIGTTOU":   syscall.SIGTTOU,
	"SIGUNUSED": syscall.SIGUNUSED,
	"SIGURG":    syscall.SIGURG,
	"SIGUSR1":   syscall.SIGUSR1,
	"SIGUSR2":   syscall.SIGUSR2,
	"SIGVTALRM": syscall.SIGVTALRM,
	"SIGWINCH":  syscall.SIGWINCH,
	"SIGXCPU":   syscall.SIGXCPU,
	"SIGXFSZ":   syscall.SIGXFSZ,
}
