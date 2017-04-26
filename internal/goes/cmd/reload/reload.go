// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package reload provides the named command that sends SIGHUP to all of the
// daemons associated with this executable.
package reload

import (
	"syscall"

	"github.com/platinasystems/go/internal/assert"
	"github.com/platinasystems/go/internal/goes/lang"
	"github.com/platinasystems/go/internal/kill"
)

const (
	Name    = "reload"
	Apropos = "SIGHUP this goes machine"
	Usage   = "reload"
)

type Interface interface {
	Apropos() lang.Alt
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }

func (cmd) Main(...string) error {
	err := assert.Root()
	if err != nil {
		return err
	}
	return kill.All(syscall.SIGHUP)
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
