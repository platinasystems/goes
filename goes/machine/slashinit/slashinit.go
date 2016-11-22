// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package slashinit provides the '/init' command that mounts and pivots to the
// 'goesroot' kernel parameter before executing its '/sbin/init'.  The machine
// may reassign the Hook closure to perform target specific tasks prior to the
// 'goesroot' pivot. The kernel command may include 'goes=overwrite' to force
// copy of '/bin/goes' from the initrd to the named 'goesroot'.
//
// If the target root is not mountable, the 'goesinstaller' parameter specifies
// an installer/recovery system to use to repair the system. The parameter to
// this is three comma-seperated URLs. The first is mandatory, and is the
// kernel to load. The second is the optional initramfs to load. The third is
// the optional FDT to load. The kernel is loaded via the kexec command.
package slashinit

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/cavaliercoder/grab"
	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/log"
	"github.com/platinasystems/go/url"
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

func (cmd) makeRootDirs() (err error) {
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
	return nil
}

func (cmd) makeRootFiles() (err error) {
	for _, cp := range []struct {
		src, dst string
	}{
		{"/init", "/newroot/usr/bin/goes"},
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
	return nil
}

func (cmd) makeRootLinks() (err error) {
	for _, ln := range []struct {
		src, dst string
	}{
		{"../usr/bin/goes", "/newroot/sbin/init"},
	} {
		if _, err = os.Stat(ln.dst); os.IsNotExist(err) {
			err = os.Symlink(ln.src, ln.dst)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (cmd) moveVirtualFileSystems() (err error) {
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
	return nil
}

func (cmd) unlinkRootFiles() error {
	for _, fn := range []string{
		"/usr/bin/gdbserver",
		"/init",
		"/bin/goes",
	} {
		syscall.Unlink(fn)
	}
	return nil
}

func (cmd) rmdirRootDirs() error {
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
	return nil
}

func (cmd) makeTargetDirs () (err error) {
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
	return nil
}

func (cmd) makeTargetLinks() (err error) {
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
	return nil
}

func (cmd) mountTargetVirtualFilesystems() (err error) {
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
	return nil
}

func (c cmd) Main(_ ...string) error {
	goesRootEnv := os.Getenv("goesroot")
	goesRoot := filepath.SplitList(goesRootEnv)
	goesinstaller := os.Getenv("goesinstaller")
	err := Hook()
	if err != nil {
		return err
	}
	if len(goesRoot) >= 1 && len(goesRoot[0]) > 0 {
		_, err = os.Stat("/newroot")
		if os.IsNotExist(err) {
			err = os.Mkdir("/newroot", os.FileMode(0755))
			if err != nil {
				log.Print("err", "mkdir", "/newroot", ":", err)
			}
		}
		err = command.Main("mount", goesRoot[0], "/newroot")
		if err != nil {
			if len(goesinstaller) > 0 {
				params := strings.Split(goesinstaller, ",")
				err = installer(params)
				if err != nil {
					log.Print("err", "installer", params[0],
						":", err)
				}
			} else {
				log.Print("err", "mount", goesRoot[0],
					"/newroot", ":", err)
			}
		} else {
			if len(goesRoot) >= 2 && len(goesRoot[1]) > 0 {
				err := command.Main("source", goesRoot[1])
				if err != nil {
					log.Print("err", "source", goesRoot[1], ":",
						err)
				}
			}
			c.makeRootDirs()
			c.makeRootFiles()
			c.makeRootLinks()
			c.moveVirtualFileSystems()
			
			if err = os.Chdir("/newroot"); err != nil {
				return err
			}
			c.unlinkRootFiles()
			c.rmdirRootDirs()
			err = syscall.Mount("/newroot", "/", "", syscall.MS_MOVE, "")
			if err != nil {
				return err
			}
			if err = syscall.Chroot("."); err != nil {
				return err
			}
			c.makeTargetDirs()
			c.makeTargetLinks()
			c.mountTargetVirtualFilesystems()
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
				err = command.Main("start")
			}
			return err
		}
	}
	return nil
}

func installer(params []string) error {
	if len(params) < 1 || len(params[0]) == 0 {
		return fmt.Errorf("KERNEL: missing")
	}

	reqs := make([]*grab.Request, 0)

	req, err := grab.NewRequest(params[0])
	if err != nil {
		return err
	}
	req.Filename = "kernel"
	reqs = append(reqs, req)

	if len(params) >= 2 && len(params[1]) > 0 {
		req, err := grab.NewRequest(params[1])
		if err != nil {
			return err
		}
		req.Filename = "initramfs"
		reqs = append(reqs, req)
	}

	if len(params) >= 3 && len(params[2]) > 0 {
		req, err := grab.NewRequest(params[2])
		if err != nil {
			return err
		}
		req.Filename = "fdt"
		reqs = append(reqs, req)
	}

	successes, err := url.FetchReqs(0, reqs)
	if err != nil {
		return err
	}

	if successes == len(reqs) {
		fmt.Printf("All files loaded successfully!")
	}

	return command.Main("kexec", "-e", "-k", "kernel", "-i", "initramfs",
		"-c", "console=ttyS0,115200")
}
