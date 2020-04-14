// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package slashinit

import (
	"fmt"
	"os"
	"syscall"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/log"
	"github.com/platinasystems/goes/lang"
)

const zero = uintptr(0)

type Command struct {
	Hook   func() error
	FsHook func() error
	g      *goes.Goes
}

func (*Command) String() string { return "/init" }

func (*Command) Usage() string { return "init" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "bootstrap",
	}
}

func (*Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The '/init' command sets up the initial system environment.
	It mounts the standard virtual filesystems (/dev, /dev/pts,
	/proc, /sys) and redirects the init process I/O to /dev/kmsg.

	The machine has two available hooks. One happens early as the
	filesystems are being set up (FsHook). This is for platform-
	specific filesysetm setup. One happens after the setup in
	an environment where all the filesystems are set up (Hook).
`,
	}
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (*Command) Kind() cmd.Kind { return cmd.DontFork }

func (c *Command) Main(_ ...string) error {
	c.mountVirtualFilesystems()
	c.makeStdioLinks()
	c.redirectStdioKmsg()

	if c.Hook != nil {
		if err := c.Hook(); err != nil {
			log.Print("Error from board hook", err)
		}
	}
	c.makeTargetDirs()
	c.makeTargetLinks()
	if c.FsHook != nil {
		if err := c.FsHook(); err != nil {
			log.Print("Error from filesystem hook: ", err)
		}
	}
	fmt.Printf("starting start\n")
	return c.g.Main("start")
}

func (*Command) mountVirtualFilesystems() {
	for _, mnt := range []struct {
		dir    string
		dev    string
		fstype string
		mode   os.FileMode
	}{
		{"/dev", "devtmpfs", "devtmpfs", 0755},
		{"/dev/pts", "devpts", "devpts", 0755},
		{"/proc", "proc", "proc", 0555},
		{"/sys", "sysfs", "sysfs", 0555},
		{"/run", "tmpfs", "tmpfs", 0755},
		{"/tmp", "tmpfs", "tmpfs", 01777},
	} {
		if _, err := os.Stat(mnt.dir); os.IsNotExist(err) {
			err = os.Mkdir(mnt.dir, os.FileMode(mnt.mode))
			if err != nil {
				log.Print("err", mnt.dir, ": ", err)
				continue
			}
		}
		err := syscall.Mount(mnt.dev, mnt.dir, mnt.fstype, zero, "")
		if err != nil {
			log.Print("err", mnt.dir, ": ", err)
		}
	}
}

func (*Command) makeStdioLinks() {
	for _, ln := range []struct {
		src, dst string
	}{
		{"../proc/self/fd/0", "/dev/stdin"},
		{"../proc/self/fd/1", "/dev/stdout"},
		{"../proc/self/fd/2", "/dev/stderr"},
	} {
		if _, err := os.Stat(ln.dst); os.IsNotExist(err) {
			err = os.Symlink(ln.src, ln.dst)
			if err != nil {
				log.Print("err", "ln", ln.dst, "->", ln.src,
					":", err)
			}
		}
	}
}

func (*Command) redirectStdioKmsg() {
	for fd := 0; fd <= 2; fd++ {
		err := syscall.Close(fd)
		if err != nil {
			log.Print("err", "console close", ":", err)
		}
		_, err = syscall.Open("/dev/kmsg", syscall.O_RDWR, 0)
		if err != nil {
			log.Print("err", "console reopen", ":", err)
		}
	}
}

func (*Command) makeTargetDirs() {
	for _, dir := range []struct {
		name string
		mode os.FileMode
	}{
		{"/root", 0700},
		{"/tmp", 01777},
		{"/var", 0755},
	} {
		if _, err := os.Stat(dir.name); os.IsNotExist(err) {
			err = os.Mkdir(dir.name, dir.mode)
			if err != nil {
				panic(fmt.Errorf("mkdir %s: %s", dir.name, err))
			}
		}
	}
}

func (*Command) makeTargetLinks() {
	for _, ln := range []struct {
		src, dst string
	}{
		{"../run", "/var/run"},
	} {
		if _, err := os.Stat(ln.dst); os.IsNotExist(err) {
			err = os.Symlink(ln.src, ln.dst)
			if err != nil {
				panic(fmt.Errorf("ln %s -> %s: %s",
					ln.dst, ln.src, err))
			}
		}
	}
}
