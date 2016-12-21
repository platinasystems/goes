// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package stop provides the named command that kills all of the daemons
// associated with this executable.
package stop

import (
	"os"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/internal/parms"
	"github.com/platinasystems/go/goes/machine/internal"
	"github.com/platinasystems/go/goes/pidfile"
	"github.com/platinasystems/go/goes/sockfile"
)

const Name = "stop"

// Machines may use Hook to run something between the kill of all daemons and
// the removal of the socks and pids directories.
var Hook = func() error { return nil }

func New() *cmd { return new(cmd) }

type cmd goes.ByName

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return "stop [OPTION]..." }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	byName := goes.ByName(*c)
	parm, args := parms.New(args, "-start", "-stop")
	err := internal.AssertRoot()
	if err != nil {
		return err
	}
	stop := parm["-stop"]
	if len(stop) == 0 {
		if _, xerr := os.Stat("/etc/goes/stop"); xerr == nil {
			stop = "/etc/goes/stop"
		}
	}
	if len(stop) > 0 {
		if err = byName.Main("source", stop); err != nil {
			return err
		}
	}
	err = internal.KillAll(syscall.SIGTERM)
	time.Sleep(5 * time.Second)
	if e := internal.KillAll(syscall.SIGKILL); err == nil {
		err = e
	}
	if t := Hook(); t != nil {
		if err != nil {
			err = t
		}
	}
	os.RemoveAll(sockfile.Dir)
	os.RemoveAll(pidfile.Dir)
	return err
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "stop this goes machine",
	}
}

func (c *cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	stop - stop this goes machine

SYNOPSIS
	stop [-stop=URL]

DESCRIPTION
	Stop all embedded daemons.

OPTIONS
	-stop URL
		Specifies the URL of the machine's stop script that's
		sourced immediately before killing all daemons.
		default: /etc/goes/start`,
	}
}
