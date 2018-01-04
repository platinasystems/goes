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
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

const procSelfNsNet = "/proc/self/ns/net"

const varRunNetnsMode = syscall.S_IRWXU |
	syscall.S_IRGRP |
	syscall.S_IXGRP |
	syscall.S_IROTH |
	syscall.S_IXOTH

type Command struct{}

func (Command) String() string { return "add" }

func (Command) Usage() string {
	return `ip netns add NETNSNAME`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "create network namespace",
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
		err = syscall.Mkdir(rtnl.VarRunNetns, varRunNetnsMode)
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
	f.Close()

	del := func() {
		syscall.Unmount(fn, syscall.MNT_DETACH)
		syscall.Unlink(fn)
	}

	err = syscall.Unshare(syscall.CLONE_NEWNET)
	if err != nil {
		del()
		return err
	}
	err = syscall.Mount(procSelfNsNet, fn, "none", syscall.MS_BIND, "")
	if err != nil {
		del()
		return err
	}
	return nil
}
