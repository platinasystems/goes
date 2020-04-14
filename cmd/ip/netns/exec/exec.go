// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exec

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/external/flags"
	"github.com/platinasystems/goes/internal/netns"
	"github.com/platinasystems/goes/lang"
)

var (
	missingCommandErr = errors.New("missing COMMAND")
	missingNameErr    = errors.New("missing NAME")
)

type Command struct {
	g *goes.Goes
}

func (*Command) String() string { return "exec" }

func (*Command) Usage() string {
	return "ip netns exec [ -a[ll] | NAME ] COMMAND [ ARGS ]..."
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "network namespace",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
SEE ALSO
	ip man netns || ip netns -man
	man ip || ip -man`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (c *Command) Main(args ...string) error {
	var x *exec.Cmd
	sigch := make(chan os.Signal)
	signal.Notify(sigch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer signal.Stop(sigch)
	run := func() error {
		x.Stdin = os.Stdin
		x.Stdout = os.Stdout
		x.Stderr = os.Stderr
		err := x.Start()
		if err != nil {
			return fmt.Errorf("%v: %v", x.Args, err)
		}
		done := make(chan error)
		go func() { done <- x.Wait() }()
	again:
		select {
		case err = <-done:
		case sig := <-sigch:
			x.Process.Signal(sig)
			goto again
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
