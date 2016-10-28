// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package npu

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const npuCmdSock = "/run/goes/socks/npu"

type npu struct{}

func New() npu { return npu{} }

func (npu) String() string { return "npu" }
func (npu) Usage() string  { return "npu [COMMAND-STRING]..." }

func (npu) Main(args ...string) error {
	conn, err := net.Dial("unix", npuCmdSock)
	if err != nil {
		return err
	}
	defer conn.Close()
	fmt.Fprintln(conn, strings.Join(args, " ")+"\nquit\n")
	for {
		var buf [4096]byte
		var n int
		n, err = conn.Read(buf[:])
		if err != nil {
			break
		}
		os.Stdout.Write(buf[:n])
	}
	return err
}

func (npu) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "send commands to hidden cli",
	}
}

func (npu) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	npu - send commands to hidden cli

SYNOPSIS
	npu [command]

DESCRIPTION
	Send argument to npu cli

EXAMPLES
	npu	"show interfaces"`,
	}
}
