// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package delete

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"syscall"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "delete"
	Apropos = "remove network namespace"
	Usage   = `ip [-all] netns delete [NETNSNAME]`
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

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	opt, args := options.New(args)

	switch {
	case opt.Flags.ByName["-all"]:
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
		varRunNetns, err := ioutil.ReadDir(rtnl.VarRunNetns)
		if err != nil {
			return err
		}
		for _, fi := range varRunNetns {
			if err := del(fi.Name()); err != nil {
				return err
			}
		}
	case len(args) == 0:
		return fmt.Errorf("NETNSNAME: missing")
	case len(args) == 1:
		return del(args[0])
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	return nil
}

func del(name string) error {
	fn := filepath.Join(rtnl.VarRunNetns, name)
	if err := syscall.Unmount(fn, syscall.MNT_DETACH); err != nil {
		return err
	}
	return syscall.Unlink(fn)
}
