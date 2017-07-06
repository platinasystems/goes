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

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "identify"
	Apropos = "print name of network namespace for given PID"
	Usage   = `
	ip netns identify [ PID ]
	`
	Man = `
SEE ALSO
	ip man netns || ip netns -man
`
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

var man = lang.Alt{
	lang.EnUS: Man,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

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
