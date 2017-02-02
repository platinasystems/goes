// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package uninstall provides the named command that stops and removes
// /usr/bin/goes and it's associated files.
package uninstall

import (
	"os/exec"
	"syscall"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/prog"
)

const Name = "uninstall"
const EtcInitdGoes = "/etc/init.d/goes"
const EtcDefaultGoes = "/etc/default/goes"
const BashCompletionGoes = "/usr/share/bash-completion/completions/goes"

// Machines may use this Hook to complete its removal.
var Hook = func() error { return nil }

type cmd goes.ByName

func New() *cmd { return new(cmd) }

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return Name }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(...string) error {
	err := goes.ByName(*c).Main("stop")
	if err != nil {
		return err
	}
	exec.Command("/usr/sbin/update-rc.d", "goes", "remove").Run()
	err = Hook()
	syscall.Unlink(EtcInitdGoes)
	syscall.Unlink(EtcDefaultGoes)
	syscall.Unlink(BashCompletionGoes)
	syscall.Unlink(prog.Install)
	return err
}

func (*cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "uninstall this goes machine",
	}
}
