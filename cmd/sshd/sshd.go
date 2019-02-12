// Copyright © 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package sshd is a ssh server daemon

package sshd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/gliderlabs/ssh"

	"github.com/kr/pty"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/lang"
)

type Command struct {
	g    *goes.Goes
	done chan struct{}
}

func (*Command) String() string { return "sshd" }

func (*Command) Usage() string { return "sshd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "ssh server daemon",
	}
}

func (c *Command) Close() error {
	close(c.done)
	return nil
}

func (c *Command) Goes(g *goes.Goes) { c.g = g }

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func (c *Command) Main(args ...string) error {
	ssh.Handle(func(s ssh.Session) {
		cmd := exec.Command("/proc/self/exe", "cli")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			f, err := pty.Start(cmd)
			if err != nil {
				panic(err)
			}
			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()
			go func() {
				io.Copy(f, s) // stdin
			}()
			io.Copy(s, f) // stdout
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	})

	err := ssh.ListenAndServe(":2222", nil)
	return err
}
