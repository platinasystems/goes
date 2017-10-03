// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exec

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/ip/internal/netns"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
)

const (
	Name    = "exec"
	Apropos = "network namespace"
	Usage   = "ip netns exec [ -all | NETNSNAME ] COMMAND..."
	Man     = `
SEE ALSO
	ip man netns || ip netns -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() *Command { return new(Command) }

type Command struct {
	g *goes.Goes
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (c *Command) Goes(g *goes.Goes) { c.g = g }
func (*Command) Man() lang.Alt       { return man }
func (*Command) String() string      { return Name }
func (*Command) Usage() string       { return Usage }

func (c *Command) Main(args ...string) error {
	flag, args := flags.New(args, []string{"-a", "-all"})
	if flag.ByName["-a"] {
		if len(args) == 0 {
			return fmt.Errorf("command: missing")
		}
		dir, err := ioutil.ReadDir("/var/run/netns")
		if err != nil {
			return err
		}
		a := make([]string, len(args)+2)
		a[0] = "exec"
		for i, arg := range args {
			a[i+2] = arg
		}
		for _, fi := range dir {
			a[1] = fi.Name()
			x := c.g.Fork(a...)
			x.Stdin = os.Stdin
			x.Stdout = os.Stdout
			x.Stderr = os.Stderr
			if err = x.Run(); err != nil {
				return err
			}
		}
		return nil
	}
	if len(args) == 0 {
		return fmt.Errorf("NETNSNAME: missing")
	}
	name := args[0]
	args = args[1:]
	if len(args) == 0 {
		return fmt.Errorf("command: missing")
	}
	err := netns.Switch(name)
	if err != nil {
		return err
	}

	x := exec.Command(args[0], args[1:]...)
	x.Stdin = os.Stdin
	x.Stdout = os.Stdout
	x.Stderr = os.Stderr
	return x.Run()
}

func (*Command) Complete(args ...string) (list []string) {
	var larg string
	n := len(args)
	if n > 0 {
		larg = args[n-1]
	}
	for _, name := range []string{
		"-all",
	} {
		if strings.HasPrefix(name, larg) {
			list = append(list, name)
		}
	}
	if len(list) == 0 && n == 1 {
		list = netns.CompleteName(larg)
	}
	return
}
