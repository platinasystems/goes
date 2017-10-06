// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exec

import (
	"errors"
	"fmt"
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
	Usage   = "ip netns exec [ -a[ll] | NAME ] COMMAND [ ARGS ]..."
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
	missingCommandErr = errors.New("missing COMMAND")
	missingNameErr    = errors.New("missing NAME")
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
	var x *exec.Cmd
	run := func() error {
		x.Stdin = os.Stdin
		x.Stdout = os.Stdout
		x.Stderr = os.Stderr
		err := x.Run()
		if err != nil {
			err = fmt.Errorf("%v: %v", x.Args, err)
		}
		return err
	}
	flag, args := flags.New(args, []string{"-a", "-all"})
	if len(args) == 0 {
		err := missingNameErr
		if flag.ByName["-a"] {
			err = missingCommandErr
		}
		return err
	}
	if flag.ByName["-a"] {
		for _, name := range netns.List() {
			x = c.g.Fork(append([]string{"exec", name}, args...)...)
			if err := run(); err != nil {
				return err
			}
		}
		// now default namespace
		x = exec.Command(args[0], args[1:]...)
		return run()
	}
	if len(args) < 2 {
		return missingCommandErr
	}
	if err := netns.Switch(args[0]); err != nil {
		return err
	}
	x = exec.Command(args[1], args[2:]...)
	return run()
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
