// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package restart provides the named command that stops; then starts  all of
// the daemons associated with this executable.
package restart

import "github.com/platinasystems/go/command"

const Name = "restart"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd) Main(...string) error {
	err := command.Main("stop")
	if err != nil {
		return err
	}
	return command.Main("start")
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "stop, then start this goes machine",
	}
}
