// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vnet

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/net/vnet/internal"
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
	iargs := make([]interface{}, len(args))
	for i, s := range args {
		iargs[i] = s
	}
	fmt.Fprintln(internal.Conn, iargs...)
	fmt.Fprintln(internal.Conn, "break")
	for {
		var buf [4096]byte
		var n int
		n, err = internal.Conn.Read(buf[:])
		if err != nil {
			break
		}
		x := bytes.Index(buf[:n], []byte("unknown: break"))
		if x >= 0 {
			if x > 0 {
				os.Stdout.Write(buf[:x])
			}
			break
		} else {
			os.Stdout.Write(buf[:n])
		}
	}
	return err
}

func (cmd) Help(...string) string {
	err := internal.Conn.Connect()
	if err != nil {
		return err.Error()
	}
	buf := make([]byte, 4*4096)
	fmt.Fprintln(internal.Conn, "help")
	fmt.Fprintln(internal.Conn, "break")
	for i, n := 0, 0; i < len(buf); i += n {
		n, err = internal.Conn.Read(buf[i:])
		if err != nil {
			if err != io.EOF {
				return err.Error()
			}
			return string(buf[:i+n])
		}
		x := bytes.Index(buf[:i+n], []byte("unknown: break"))
		if x >= 0 {
			return string(buf[:x])
		}
	}
	return string(buf)
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
