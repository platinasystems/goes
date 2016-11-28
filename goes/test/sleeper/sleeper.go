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

	"github.com/platinasystems/go/parms"
)

const Name = "sleeper"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) Daemon() int    { return -1 }
func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [-s SECONDS] [MESSAGE]..." }

func (cmd) Main(args ...string) error {
	var sec uint

	parm, args := parms.New(args, "-s")
	if len(parm["-s"]) == 0 {
		parm["-s"] = "3"
	}

	if len(args) == 0 {
		args = []string{"yawn"}
	}

	msg := strings.Join(args, " ")
	if len(msg) == 0 {
		msg = "yawn"
	}

	_, err := fmt.Sscan(parm["-s"], &sec)
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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "periodic message logger",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	sleeper - periodic message logger

SYNOPSIS
	sleeper [-s SECONDS] [MESSAGE]...

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
	goes>`,
	}
}
