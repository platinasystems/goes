// Copyright © 2016-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package stop provides the named command that kills all of the daemons
// associated with this executable.
package stop

import (
	"fmt"
	"os"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/internal/assert"
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
			fmt.Printf("source %s: %s\n", stop, err)
		}
	}
	err = c.g.Main("daemons", "stop")
	if err != nil {
		fmt.Printf("Error from daemons stop: %s\n", err)
	}
	return err
}

func haveEtcGoesStop() bool {
	_, err := os.Stat(EtcGoesStop)
	return err == nil
}
