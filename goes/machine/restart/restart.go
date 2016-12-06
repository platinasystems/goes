// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package restart provides the named command that stops; then starts  all of
// the daemons associated with this executable.
package restart

import "github.com/platinasystems/go/goes"

func New() *goes.Goes {
	cmd := new(cmd)
	return &goes.Goes{
		Name:   "restart",
		ByName: cmd.ByName,
		Main:   cmd.Main,
		Usage:  "restart",
		Apropos: map[string]string{
			"en_US.UTF-8": "stop, then start this goes machine",
		},
	}
}

type cmd struct {
	byName goes.ByName
}

func (cmd *cmd) ByName(byName goes.ByName) {
	cmd.byName = byName
}

func (cmd *cmd) Main(args ...string) error {
	err := cmd.byName.Main("stop")
	if err != nil {
		return err
	}
	return cmd.byName.Main("start")
}
