// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package reload provides the named command that sends SIGHUP to all of the
// daemons associated with this executable.
package reload

import (
	"syscall"

	"github.com/platinasystems/go/initutils/internal"
)

const Name = "reload"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd) Main(...string) error {
	err := internal.AssertRoot()
	if err != nil {
		return err
	}
	return internal.KillAll(syscall.SIGHUP)
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "SIGHUP this goes machine",
	}
}
