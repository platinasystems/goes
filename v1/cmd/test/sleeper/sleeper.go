// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package sleeper provides a test ticker daemon.
package sleeper

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/parms"
	"github.com/platinasystems/goes/lang"
)

const (
	Name    = "sleeper"
	Apropos = "periodic message logger"
	Usage   = "sleeper [-s SECONDS] [MESSAGE]..."
	Man     = `
DESCRIPTION
	Periodicaly log MESSAGE.  The default is 3 seconds.

EXAMPLE
	Use this to test job control like this:

	goes> sleeper &
	[1] Runing	sleeper
	goes> show log
	2015/10/28 22:06:18 sleeper
	goes> show jobs
	[1] Runing	sleeper
	goes> kill %1
	goes> show jobs
	[1] Terminated	sleeper
	goes>`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) Kind() cmd.Kind    { return cmd.Daemon }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var sec uint

	parm, args := parms.New(args, "-s")
	if len(parm.ByName["-s"]) == 0 {
		parm.ByName["-s"] = "3"
	}

	if len(args) == 0 {
		args = []string{"yawn"}
	}

	msg := strings.Join(args, " ")
	if len(msg) == 0 {
		msg = "yawn"
	}

	_, err := fmt.Sscan(parm.ByName["-s"], &sec)
	if err != nil {
		return err
	}

	d := time.Duration(sec) * time.Second

	ticker := time.NewTicker(d)
	defer ticker.Stop()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sig:
			fmt.Println("killed")
			return nil
		case <-ticker.C:
			fmt.Println(msg)
		}
	}
}
