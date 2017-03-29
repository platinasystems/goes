// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package stop provides the named command that kills all of the daemons
// associated with this executable.
package stop

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/assert"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/kill"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/go/internal/sockfile"
)

const Name = "stop"
const EtcGoesStop = "/etc/goes/stop"

// Machines may use Hook to run something between the kill of all daemons and
// the removal of the socks and pids directories.
var Hook = func() error { return nil }

func New() *cmd { return new(cmd) }

type cmd goes.ByName

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return "stop [OPTION]..." }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	parm, args := parms.New(args, "-start", "-stop")
	err := assert.Root()
	if err != nil {
		return err
	}
	if prog.Name() != prog.Install && prog.Base() != "init" {
		return fmt.Errorf("use `%s stop`", prog.Install)
	}
	stop := parm["-stop"]
	if len(stop) == 0 && haveEtcGoesStop() {
		stop = EtcGoesStop
	}
	if len(stop) > 0 {
		err = goes.ByName(*c).Main("source", stop)
		if err != nil {
			return fmt.Errorf("source %s: %v", stop, err)
		}
	}
	err = kill.All(syscall.SIGTERM)
	time.Sleep(5 * time.Second)
	if e := kill.All(syscall.SIGKILL); err == nil {
		err = e
	}
	if t := Hook(); t != nil {
		if err != nil {
			err = t
		}
	}
	os.RemoveAll(sockfile.Dir)
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

func haveEtcGoesStop() bool {
	_, err := os.Stat(EtcGoesStop)
	return err == nil
}
