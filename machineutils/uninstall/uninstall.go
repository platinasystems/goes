// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package uninstall provides the named command that stops and removes
// /usr/bin/goes and it's associated files.
package uninstall

import (
	"os"
	"os/exec"

	"github.com/platinasystems/go/machineutils/internal"
)

const Name = "uninstall"
const UsrBinGoes = "/usr/bin/goes"
const EtcInitdGoes = "/etc/init.d/goes"
const EtcDefaultGoes = "/etc/default/goes"
const BashCompletionGoes = "/usr/share/bash-completion/completions/goes"

// Machines may use this Hook to complete its removal.
var Hook = func() error { return nil }

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func (cmd) Main(...string) error {
	err := internal.AssertRoot()
	if err != nil {
		return err
	}
	_, err = os.Stat(UsrBinGoes)
	exec.Command(UsrBinGoes, "stop").Run()
	os.Remove(EtcInitdGoes)
	os.Remove(EtcDefaultGoes)
	os.Remove(BashCompletionGoes)
	exec.Command("/usr/sbin/update-rc.d", "goes", "remove").Run()
	err = Hook()
	os.Remove(UsrBinGoes)
	return err
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "uninstall this goes machine",
	}
}
