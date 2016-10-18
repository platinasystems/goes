// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package sbininit provides `/sbin/init` that is run from `/init`.  This
// starts a redis server followed by all configured daemons before repatedly
// running the cli on the console.
//
// If present, this sources `/etc/goes` which set these variables.
//
//	REDISD		list of net devices that the server listens to
//			default: lo
//	MACHINED	machined arguments
package sbininit

import (
	"fmt"
	"os"
	"strings"

	"github.com/platinasystems/go/command"
)

// If present, /etc/goes is sourced before running redisd, machined, and
// the remaining damons.
const EtcGoes = "/etc/goes"

// Machines may use this Hook to run something before redisd, machined, etc.
var Hook = func() error { return nil }

type sbinInit struct{}

func New() sbinInit { return sbinInit{} }

func (sbinInit) String() string { return "/sbin/init" }
func (sbinInit) Usage() string  { return "/sbin/init" }

func (sbinInit) Main(...string) error {
	if pid := os.Getpid(); pid != 1 {
		return fmt.Errorf("%d: pid not 1", pid)
	}
	err := Hook()
	if err != nil {
		return err
	}
	if _, err = os.Stat(EtcGoes); err == nil {
		err = command.Main("source", EtcGoes)
		if err != nil {
			return err
		}
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
	login := command.Find("login") != nil
	for {
		if login {
			err = command.Main("login")
			if err != nil {
				fmt.Println("login:", err)
				continue
			}
		}
		command.Main("cli")
	}
	return nil
}
