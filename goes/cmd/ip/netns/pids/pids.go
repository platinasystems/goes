// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pids

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "pids"
	Apropos = "list PIDS in given network namespace"
	Usage   = `ip netns pids NETNSNAME`
	Man     = `
SEE ALSO
	ip man netns || ip netns -man
	man ip || ip -man`
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
	if n := len(args); n == 0 {
		return fmt.Errorf("NETNSNAME: missing")
	} else if n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	name := args[0]
	nameStat, err := os.Stat(filepath.Join(rtnl.VarRunNetns, name))
	if err != nil {
		return err
	}
	nameStatSt := nameStat.Sys().(*syscall.Stat_t)
	matches, err := filepath.Glob("/proc/*/ns/net")
	if err != nil {
		return err
	}
	for _, procNsNet := range matches {
		procNsNetStat, err := os.Stat(procNsNet)
		if err != nil {
			return err
		}
		procNsNetStatSt := procNsNetStat.Sys().(*syscall.Stat_t)
		if procNsNetStatSt.Dev == nameStatSt.Dev &&
			procNsNetStatSt.Ino == nameStatSt.Dev {
			s := strings.TrimPrefix(procNsNet, "/proc/")
			fmt.Println(strings.TrimSuffix(s, "/ns/net"))
		}
	}
	return nil
}
