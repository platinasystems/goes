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

	"github.com/platinasystems/go/goes/machine/internal"
	"github.com/platinasystems/go/pidfile"
	"github.com/platinasystems/go/sockfile"
)

const Name = "stop"

// Machines may use Hook to run something between the kill of all daemons and
// the removal of the socks and pids directories.
var Hook = func() error { return nil }

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd) Main(...string) error {
	err := internal.AssertRoot()
	if err != nil {
		return err
	}
	err = internal.KillAll(syscall.SIGTERM)
	time.Sleep(time.Second)
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

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "stop this goes machine",
	}
}
