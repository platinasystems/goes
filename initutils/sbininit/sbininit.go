// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package sbininit provides both '/sbin/init' and '/usr/sbin/goesd' goes
// commands that run a redis server prior to running all of the registered goes
// daemons. After starting the daemons, '/sbin/init' returns to 'goes.Goes'
// that then runs the console shell; where as, '/usr/sbin/goesd' instead waits
// for a kill signal then stops all daemons.
//
// A machine main may reassign the Hook closure to perform target specific
// tasks prior to running the daemons. Similarly, the 'GOESRC' environment
// variable may name a script to run before the daemons.
//
// The 'REDISD_DEVS' environment variable may be a space separated list of
// redis listening net devices. The default is "lo".
package sbininit

import (
	"os"

	"github.com/platinasystems/go/initutils/internal"
	"github.com/platinasystems/go/log"
)

type sbinInit struct{}

func New() sbinInit { return sbinInit{} }

func (sbinInit) String() string { return "/sbin/init" }
func (sbinInit) Usage() string  { return "/sbin/init" }

func (sbinInit) Main(_ ...string) error {
	defer func(stdin, stdout, stderr *os.File) {
		os.Stdin, os.Stdout, os.Stderr = stdin, stdout, stderr
	}(os.Stdin, os.Stdout, os.Stderr)
	var err error
	if os.Stdin, err = os.Open(os.DevNull); err != nil {
		return err
	}
	if os.Stdout, _ = log.Pipe("info"); err != nil {
		return err
	}
	if os.Stderr, _ = log.Pipe("err"); err != nil {
		return err
	}
	return internal.Init.Start()
}
