// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package identify

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/platinasystems/goes/cmd/ip/internal/options"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/goes/internal/nl/rtnl"
)

type Command struct{}

func (Command) String() string { return "identify" }

func (Command) Usage() string {
	return `ip netns identify [ PID ]`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "print name of network namespace for given PID",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
SEE ALSO
	ip man netns || ip netns -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	_, args = options.New(args)
	proc := "self"
	if n := len(args); n == 1 {
		proc = args[0]
	} else if n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	procNsNet := filepath.Join("/proc", proc, "ns/net")
	procNsNetStat, err := os.Stat(procNsNet)
	if err != nil {
		return err
	}
	procNsNetStatSt := procNsNetStat.Sys().(*syscall.Stat_t)
	varRunNetns, err := ioutil.ReadDir(rtnl.VarRunNetns)
	if err != nil {
		return err
	}
	name := "default"
	for _, fi := range varRunNetns {
		fn := filepath.Join(rtnl.VarRunNetns, fi.Name())
		nameStat, err := os.Stat(fn)
		if err != nil {
			return err
		}
		nameStatSt := nameStat.Sys().(*syscall.Stat_t)
		if procNsNetStatSt.Dev == nameStatSt.Dev &&
			procNsNetStatSt.Ino == nameStatSt.Ino {
			name = fi.Name()
			break
		}
	}
	fmt.Println(name)
	return nil
}
