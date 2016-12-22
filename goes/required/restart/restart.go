// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package restart provides the named command that stops; then starts  all of
// the daemons associated with this executable.
package restart

import "github.com/platinasystems/go/goes"

const Name = "restart"

func New() *cmd { return new(cmd) }

type cmd goes.ByName

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return "restart [OPTION]..." }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	byName := goes.ByName(*c)
	err := byName.Main(append([]string{"stop"}, args...)...)
	if err != nil {
		return err
	}
	return byName.Main(append([]string{"start"}, args...)...)
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "stop, then start this goes machine",
	}
}

func (c *cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	restart - stop, then start this goes machine

SYNOPSIS
	restart [STOP, STOP, and REDISD OPTIONS]...

DESCRIPTION
	Run the goes machine stop then start commands.

SEE ALSO
	start, stop, and redisd`,
	}
}
