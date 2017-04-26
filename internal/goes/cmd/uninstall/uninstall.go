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
	"github.com/platinasystems/go/internal/goes/lang"
	"github.com/platinasystems/go/internal/prog"
)

const (
	Name    = "uninstall"
	Apropos = "uninstall this goes machine"
	Usage   = "uninstall"

	EtcInitdGoes       = "/etc/init.d/goes"
	EtcDefaultGoes     = "/etc/default/goes"
	BashCompletionGoes = "/usr/share/bash-completion/completions/goes"
)

// Machines may use this Hook to complete its removal.
var Hook = func() error { return nil }

type Interface interface {
	Apropos() lang.Alt
	ByName(goes.ByName)
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return new(cmd) }

type cmd goes.ByName

func (*cmd) Apropos() lang.Alt { return apropos }

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

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return Usage }

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
