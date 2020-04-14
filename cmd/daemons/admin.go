// Copyright 2016-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package daemons

import (
	"fmt"
	"os"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/atsock"
	"github.com/platinasystems/goes/lang"
)

var Admin = &goes.Goes{
	NAME:  "daemon",
	USAGE: "daemon COMMAND",
	APROPOS: lang.Alt{
		lang.EnUS: "daemon admin",
	},
	ByName: map[string]cmd.Cmd{
		"log":     Log{},
		"restart": Restart{},
		"start":   Start{},
		"status":  Status{},
		"stop":    Stop{},
	},
}

var empty = struct{}{}

type Log struct{}
type Restart struct{}
type Status struct{}
type Start struct{}
type Stop struct{}

func (Log) String() string { return "log" }

func (Log) Usage() string {
	return "daemon log [TEXT]..."
}

func (Log) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "append and show daemon log",
	}
}

func (Log) Main(args ...string) error {
	var s string
	cl, err := atsock.NewRpcClient(sockname())
	if err != nil {
		return err
	}
	defer cl.Close()
	if err = cl.Call("Daemons.Log", args, &s); err == nil {
		os.Stdout.WriteString(s)
	}
	return err
}

func (Restart) String() string { return "restart" }

func (Restart) Usage() string {
	return "daemon restart [PID]..."
}

func (Restart) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "daemon restart",
	}
}

func (Restart) Main(args ...string) error {
	pids, err := pids(args)
	if err != nil {
		return err
	}
	cl, err := atsock.NewRpcClient(sockname())
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("Daemons.Restart", pids, &empty)
}

func (Start) String() string { return "start" }

func (Start) Usage() string {
	return "daemon start"
}

func (Start) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "daemon start DAEMON [ARG]...",
	}
}

func (Start) Main(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing DAEMON [ARG]...")
	}
	cl, err := atsock.NewRpcClient(sockname())
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("Daemons.Start", args, &empty)
}

func (Status) String() string { return "status" }

func (Status) Usage() string {
	return "daemon status"
}

func (Status) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "show daemons",
	}
}

func (Status) Main(args ...string) error {
	var s string
	cl, err := atsock.NewRpcClient(sockname())
	if err != nil {
		return err
	}
	defer cl.Close()
	if err = cl.Call("Daemons.List", struct{}{}, &s); err == nil {
		os.Stdout.WriteString(s)
	}
	return err
}

func (Stop) String() string { return "stop" }

func (Stop) Usage() string {
	return "daemon stop [PID]..."
}

func (Stop) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "daemon stop",
	}
}

func (Stop) Main(args ...string) error {
	pids, err := pids(args)
	if err != nil {
		return err
	}
	cl, err := atsock.NewRpcClient(sockname())
	if err != nil {
		return err
	}
	defer cl.Close()
	return cl.Call("Daemons.Stop", pids, &empty)
}

func pids(args []string) ([]int, error) {
	if len(args) == 0 {
		return []int{}, nil
	}
	pids := make([]int, len(args))
	for i, arg := range args {
		var pid int
		_, err := fmt.Sscan(arg, &pid)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", arg, err)
		}
		pids[i] = pid
	}
	return pids, nil
}
