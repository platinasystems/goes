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

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/assert"
	"github.com/platinasystems/go/internal/kill"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/go/internal/sockfile"
)

const (
	Name    = "stop"
	Apropos = "stop this goes machine"
	Usage   = "stop [-stop=URL]"
	Man     = `
DESCRIPTION
	Stop all embedded daemons.

OPTIONS
	-stop URL
		Specifies the URL of the machine's stop script that's
		sourced immediately before killing all daemons.
		default: /etc/goes/start`

	EtcGoesStop = "/etc/goes/stop"
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

// Machines may use Hook to run something between the kill of all daemons and
// the removal of the socks and pids directories.
var Hook = func() error { return nil }

func New() *Command { return new(Command) }

type Command struct {
	g *goes.Goes
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (c *Command) Goes(g *goes.Goes) { c.g = g }
func (*Command) Man() lang.Alt       { return man }
func (*Command) String() string      { return Name }
func (*Command) Usage() string       { return Usage }

func (c *Command) Main(args ...string) error {
	parm, args := parms.New(args, "-stop")
	err := assert.Root()
	if err != nil {
		return err
	}
	if prog.Name() != prog.Install && prog.Base() != "init" {
		return fmt.Errorf("use `%s stop`", prog.Install)
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

func haveEtcGoesStop() bool {
	_, err := os.Stat(EtcGoesStop)
	return err == nil
}
