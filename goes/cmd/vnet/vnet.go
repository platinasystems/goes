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

type Command struct{}

func (Command) String() string { return "vnet" }

func (Command) Usage() string { return "vnet COMMAND [ARG]..." }

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "send commands to hidden cli",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	Send argument to vnet cli

EXAMPLES
	vnet	"show interfaces"`,
	}
}

func (Command) Close() error { return internal.Conn.Close() }

func (Command) Kind() cmd.Kind { return cmd.DontFork }

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
