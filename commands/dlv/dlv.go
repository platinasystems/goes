// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// +build !netgo

package dlv

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/derekparker/delve/cmd/dlv/cmds"
	. "github.com/derekparker/delve/version"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/parms"
	"github.com/platinasystems/go/version"
)

type dlv struct{}
type dlvd struct{}

func New() []interface{} {
	return []interface{}{
		dlv{},
		dlvd{},
	}
}

func (dlv) String() string { return "dlv" }
func (dlv) Usage() string  { return "dlv COMMAND [ARGS]..." }

func (dlvd) String() string { return "dlvd" }
func (dlvd) Usage() string  { return "dlvd -p PORT COMMAND [ARGS]..." }

func (dlv) Main(args ...string) error {
	return delve("", args...)
}

func (dlvd) Main(args ...string) error {
	parm, args := parms.New(args, "-l")
	if len(parm["-l"]) == 0 {
		parm["-l"] = "localhost:2345"
	}
	return delve(parm["-l"], args...)
}

func delve(l string, args ...string) error {
	DelveVersion.Build = version.Version
	os.Args = []string{"dlv"}
	if len(args) == 0 {
		// fall through to just show delve help
	} else if c := command.Find(args[0]); c == nil {
		os.Args = append(os.Args, args...)
	} else if command.IsDaemon(args[0]) {
		pidf, err := os.Open("/run/goes/pids/" + args[0])
		if err != nil {
			return err
		}
		pid, err := bufio.NewReader(pidf).ReadString('\n')
		pidf.Close()
		if err != nil {
			return err
		}
		if len(pid) < 2 {
			return fmt.Errorf("invalid pid: %s", pid)
		}
		os.Args = append(os.Args, "attach", pid[:len(pid)-1])
		if len(l) > 0 {
			os.Args = append(os.Args, "--headless", "-l", l)
		}
	} else {
		os.Args = append(os.Args, "exec")
		f, err := ioutil.TempFile("", "goes-dlv-")
		if err != nil {
			return err
		}
		fn := f.Name()
		fmt.Fprintf(f, "break /%s.*Main/\ncontinue\n", args[0])
		f.Close()
		defer os.Remove(fn)
		os.Args = append(os.Args, "--init", fn)
		os.Args = append(os.Args, "/usr/bin/goes")
		if len(l) > 0 {
			os.Args = append(os.Args, "--headless", "-l", l)
		}
		os.Args = append(os.Args, args...)
	}
	return cmds.New().Execute()
}
