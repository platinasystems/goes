// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vnet

import (
	"bytes"
	"fmt"
	"os"

	"github.com/platinasystems/go/goes"
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

type Interface interface {
	Apropos() lang.Alt
	Close() error
	Help(...string) string
	Kind() goes.Kind
	Main(...string) error
	Man() lang.Alt
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Man() lang.Alt     { return man }
func (cmd) Close() error      { return internal.Conn.Close() }
func (cmd) Kind() goes.Kind   { return goes.DontFork }
func (cmd) String() string    { return Name }
func (cmd) Usage() string     { return Usage }

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

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)
