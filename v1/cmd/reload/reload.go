// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package reload provides the named command that sends SIGHUP to all of the
// daemons associated with this executable.
package reload

import (
	"syscall"

	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/goes/internal/assert"
	"github.com/platinasystems/goes/internal/kill"
)

type Command struct{}

func (Command) String() string { return "reload" }

func (Command) Usage() string { return "reload" }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "SIGHUP this goes machine",
	}
}

func (Command) Main(...string) error {
	err := assert.Root()
	if err != nil {
		return err
	}
	return kill.All(syscall.SIGHUP)
}
