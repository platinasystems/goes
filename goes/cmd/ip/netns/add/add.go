// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package add

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "add"
	Apropos = "create network namespace"
	Usage   = `
	ip netns add NETNSNAME
	`
	Man = `
SEE ALSO
	ip man netns || ip netns -man
`
)

const VarRunNetnsMode = syscall.S_IRWXU |
	syscall.S_IRGRP |
	syscall.S_IXGRP |
	syscall.S_IROTH |
	syscall.S_IXOTH

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
	var netnsname string

	_, args = options.New(args)

	switch len(args) {
	case 0:
		return fmt.Errorf("NETNSNAME: missing")
	case 1:
		netnsname = args[0]
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	if _, err := os.Stat(rtnl.VarRunNetns); os.IsNotExist(err) {
		err = syscall.Mkdir(rtnl.VarRunNetns, VarRunNetnsMode)
		if err != nil {
			return err
		}
	}

	// The following came from iprout2:ip/ipnetns.c:netns_add()
	for done := false; ; {
		err := syscall.Mount("", rtnl.VarRunNetns, "none",
			syscall.MS_SHARED|syscall.MS_REC, "")
		if err == nil {
			break
		} else if err != syscall.EINVAL || done {
			return os.NewSyscallError("mount shared", err)
		}
		err = syscall.Mount(rtnl.VarRunNetns, rtnl.VarRunNetns,
			"none", syscall.MS_BIND, "")
		if err != nil {
			return os.NewSyscallError("mount bind", err)
		}
		done = true
	}

	fn := filepath.Join(rtnl.VarRunNetns, netnsname)
	f, err := os.OpenFile(fn, os.O_RDONLY|os.O_CREATE|os.O_EXCL, 0444)
	if err != nil {
		return err
	}
	defer f.Close()

	del := func() {
		syscall.Unmount(fn, syscall.MNT_DETACH)
		syscall.Unlink(fn)
	}

	if err = syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
		del()
		return err
	}
	if err = syscall.Mount("/proc/self/ns/net", fn, "none",
		syscall.MS_BIND, ""); err != nil {
		del()
		return err
	}
	// FIXME the new namespace includes interfaces from the default ns
	return nil
}
