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
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/netns"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

type Command struct{}

func (Command) String() string { return "pids" }

func (Command) Usage() string {
	return `ip netns pids NETNSNAME`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "list PIDS in given network namespace",
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

func (Command) Complete(args ...string) (list []string) {
	var larg string
	n := len(args)
	if n > 0 {
		larg = args[n-1]
	}
	list = netns.CompleteName(larg)
	return
}
