// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vnet

import (
	"bytes"
	"fmt"
	"os"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/vnet/internal"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "vnet"
	Apropos = "send commands to hidden cli"
	Usage   = "vnet COMMAND [ARG]..."
	Man     = `
DESCRIPTION
	Send argument to vnet cli

EXAMPLES
	vnet	"show interfaces"`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) Close() error      { return internal.Conn.Close() }
func (Command) Kind() cmd.Kind    { return cmd.DontFork }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
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

func (Command) Help(...string) string {
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
