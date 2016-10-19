// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package slashinit provides the '/init' command that mounts and pivots to the
// 'root' kernel parameter before executing its '/sbin/init'.  The machine may
// reassign the Hook closure to perform target specific tasks prior to the
// 'root' pivot. The kernel command may include 'goes=overwrite' to force copy
// of '/bin/goes' from the initrd to the named 'root'.
package slashinit

import (
	"io"
	"os"
	"syscall"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/log"
)

const (
	Name = "/init"
	zero = uintptr(0)
)

type cmd struct{}

var Hook = func() error { return nil }

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name }

func init() {
	if os.Getpid() != 1 {
		return
	}
	if os.Args[0] != Name {
		return
	}
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

func (cmd) Main(_ ...string) error {
	root := os.Getenv("root")
	err := Hook()
	if err != nil {
		return err
	}
	if len(root) > 0 {
		_, err = os.Stat("/newroot")
		if os.IsNotExist(err) {
			err = os.Mkdir("/newroot", os.FileMode(0755))
			if err != nil {
				return err
			}
		}
		err = command.Main("mount", root, "/newroot")
		if err != nil {
			return err
		}
		for _, dir := range []struct {
			name string
			mode os.FileMode
		}{
			{"/newroot/bin", 0775},
			{"/newroot/sbin", 0755},
			{"/newroot/usr", 0755},
			{"/newroot/usr/bin", 0755},
		} {
			if _, err = os.Stat(dir.name); os.IsNotExist(err) {
				err = os.Mkdir(dir.name, dir.mode)
				if err != nil {
					return err
				}
			}
		}
		for _, cp := range []struct {
			src, dst string
		}{
			{"/usr/bin/goes", "/newroot/usr/bin/goes"},
			{"/usr/bin/gdbserver", "/newroot/usr/bin/gdbserver"},
		} {
			_, err = os.Stat(cp.dst)
			if os.IsNotExist(err) ||
				os.Getenv("goes") == "overwrite" {
				r, err := os.Open(cp.src)
				if err != nil {
					return err
				}
				defer r.Close()
				w, err := os.Create(cp.dst)
				if err != nil {
					return err
				}
				defer w.Close()
				_, err = io.Copy(w, r)
				if err != nil {
					return err
				}
				if err = os.Chmod(cp.dst, 0755); err != nil {
					return err
				}
			}
		}
		for _, ln := range []struct {
			src, dst string
		}{
			{"../bin/goes", "/newroot/sbin/init"},
		} {
			if _, err = os.Stat(ln.dst); os.IsNotExist(err) {
				err = os.Symlink(ln.src, ln.dst)
				if err != nil {
					return err
				}
			}
		}
		for _, mv := range []struct {
			src  string
			dst  string
			mode os.FileMode
		}{
			{"/run", "/newroot/run", 0755},
			{"/sys", "/newroot/sysfs", 0555},
			{"/proc", "/newroot/proc", 0555},
			{"/dev", "/newroot/dev", 0755},
		} {
			if _, err = os.Stat(mv.dst); os.IsNotExist(err) {
				err = os.Mkdir(mv.dst, os.FileMode(mv.mode))
				if err != nil {
					return err
				}
			}
			err = syscall.Mount(mv.src, mv.dst, "",
				syscall.MS_MOVE, "")
			if err != nil {
				return err
			}
		}
		if err = os.Chdir("/newroot"); err != nil {
			return err
		}
		for _, fn := range []string{
			"/usr/bin/gdbserver",
			"/init",
			"/bin/goes",
		} {
			syscall.Unlink(fn)
		}
		for _, dir := range []string{
			"/run",
			"/sys",
			"/proc",
			"/dev",
			"/usr/bin",
			"/usr",
			"/bin",
		} {
			syscall.Rmdir(dir)
		}
		err = syscall.Mount("/newroot", "/", "", syscall.MS_MOVE, "")
		if err != nil {
			return err
		}
		if err = syscall.Chroot("."); err != nil {
			return err
		}
	}
	for _, dir := range []struct {
		name string
		mode os.FileMode
	}{
		{"/root", 0700},
		{"/tmp", 01777},
		{"/var", 0755},
	} {
		if _, err = os.Stat(dir.name); os.IsNotExist(err) {
			err = os.Mkdir(dir.name, dir.mode)
			if err != nil {
				return err
			}
		}
	}
	for _, ln := range []struct {
		src, dst string
	}{
		{"../run", "/var/run"},
	} {
		if _, err = os.Stat(ln.dst); os.IsNotExist(err) {
			err = os.Symlink(ln.src, ln.dst)
			if err != nil {
				log.Print("err", "ln", ln.dst, "->", ln.src,
					":", err)
			}
		}
	}
	for _, mnt := range []struct {
		dir    string
		dev    string
		fstype string
	}{
		{"/tmp", "tmpfs", "tmpfs"},
	} {
		err = syscall.Mount(mnt.dev, mnt.dir, mnt.fstype, zero, "")
		if err != nil {
			log.Print("err", "mount", mnt.dir, ":", err)
		}
	}
	if err = os.Setenv("PATH", "/bin:/usr/bin"); err != nil {
		return err
	}
	if err = os.Setenv("SHELL", "/bin/goes"); err != nil {
		return err
	}
	if err = os.Chdir("/root"); err != nil {
		return err
	}
	if err = os.Setenv("HOME", "/root"); err != nil {
		return err
	}
	if len(os.Getenv("TERM")) == 0 {
		if err = os.Setenv("TERM", "linux"); err != nil {
			return err
		}
	}
	const sbininit = "/sbin/init"
	_, err = os.Stat(sbininit)
	if err == nil {
		err = syscall.Exec(sbininit, []string{sbininit}, []string{
			"PATH=" + os.Getenv("PATH"),
			"SHELL=" + os.Getenv("SHELL"),
			"HOME=" + os.Getenv("HOME"),
			"TERM=" + os.Getenv("TERM"),
		})
	} else {
		err = command.Main(sbininit)
	}
	return err
}
