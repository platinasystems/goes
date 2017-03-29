// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vnet

import (
	"bytes"
	"fmt"
	"os"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/goes/cmd/vnet/internal"
)

const Name = "vnet"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) Close() error    { return internal.Conn.Close() }
func (cmd) Kind() goes.Kind { return goes.DontFork }
func (cmd) String() string  { return Name }
func (cmd) Usage() string   { return "vnet COMMAND [ARG]..." }

func (cmd) Main(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("no COMMAND")
	}
	err := internal.Conn.Connect()
	if err != nil {
		return err
	}
	err = internal.Conn.Exec(os.Stdout, args...)
	return err
}

func (cmd) Help(...string) string {
	err := internal.Conn.Connect()
	if err != nil {
		return err.Error()
	}
	buf := new(bytes.Buffer)
	err = internal.Conn.Exec(buf, "help")
	if err != nil {
		return err.Error()
	} else {
		return buf.String()
	}
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "send commands to hidden cli",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	vnet - send commands to hidden cli

SYNOPSIS
	vnet [command]

DESCRIPTION
	Send argument to vnet cli

EXAMPLES
	vnet	"show interfaces"`,
	}
}
