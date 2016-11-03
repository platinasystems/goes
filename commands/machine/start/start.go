// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package start provides the named command that runs a redis server followed
// by a machine specific daemon then all of the configured daemons. If the PID
// is 1, start doesn't return; instead, it iterates and command shell.
package start

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/commands/machine/internal"
	"github.com/platinasystems/go/sockfile"
)

const Name = "start"

// Machines may use this Hook to run something before redisd, machined, etc.
// This is typically used to set these environment variables.
//
//	REDISD		list of net devices that the server listens to
//			default: lo
//	MACHINED	machine specific arguments
var Hook = func() error { return nil }

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd cmd) Main(...string) error {
	err := internal.AssertRoot()
	if err != nil {
		return err
	}
	_, err = os.Stat(sockfile.Path("redisd"))
	if err == nil {
		return fmt.Errorf("already started")
	}
	if err = Hook(); err != nil {
		return err
	}
	args := strings.Fields(os.Getenv("REDISD"))
	if len(args) > 0 {
		err = command.Main(append([]string{"redisd"}, args...)...)
	} else {
		err = command.Main("redisd")
	}
	if err != nil {
		return err
	}
	args = strings.Fields(os.Getenv("MACHINED"))
	if len(args) > 0 {
		err = command.Main(append([]string{"machined"}, args...)...)
	} else {
		err = command.Main("machined")
	}
	if err != nil {
		return err
	}
	for daemon, lvl := range command.Daemon {
		if lvl < 0 {
			continue
		}
		err = command.Main(daemon)
		if err != nil {
			return err
		}
	}
	if os.Getpid() == 1 {
		login := command.Find("login") != nil
		for {
			if login {
				err = command.Main("login")
				if err != nil {
					fmt.Println("login:", err)
					time.Sleep(3 * time.Second)
					continue
				}
			}
			err = command.Main("cli")
			if err != nil && err != io.EOF {
				fmt.Println(err)
				<-make(chan struct{})
			}
		}
	}
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "start this goes machine",
	}
}
