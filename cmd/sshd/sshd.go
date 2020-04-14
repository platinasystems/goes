// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package sshd is a ssh server daemon
package sshd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/gliderlabs/ssh"

	"github.com/kr/pty"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/log"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/ssh_key_helper"

	gossh "golang.org/x/crypto/ssh"
)

type Command struct {
	g        *goes.Goes
	done     chan struct{}
	Addr     string
	FailSafe bool
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

func (c *Command) Main(args ...string) (err error) {
	goesDir := "/etc/goes"
	if _, err := os.Stat(goesDir); os.IsNotExist(err) {
		err = os.Mkdir(goesDir, os.FileMode(0555))
		if err != nil {
			return err
		}
	}
	keyDir := "/etc/goes/sshd"
	if _, err := os.Stat(keyDir); os.IsNotExist(err) {
		err = os.Mkdir(keyDir, os.FileMode(0600))
		if err != nil {
			return err
		}
	}
	err = ssh_key_helper.MakeRSAKeyPair("/etc/goes/sshd/id_rsa", false)
	if err != nil {
		return err
	}
	c.done = make(chan struct{})

	srv := &ssh.Server{
		Addr: ":22",
	}
	if c.Addr != "" {
		srv.Addr = c.Addr
	}

	srv.Handle(func(s ssh.Session) {
		cmdline := s.Command()
		if len(cmdline) == 0 {
			cmdline = []string{"cli"}
		}
		cmd := exec.Command("/proc/self/exe", cmdline[1:]...)
		cmd.Args[0] = cmdline[0]
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
			cmd.Stdin = s // blocks exit - do not know why
			cmd.Stdout = s
			cmd.Stderr = s.Stderr()
			err := cmd.Start()
			if err != nil {
				fmt.Printf("cmd.Start() returns %s\n", err)
				s.Exit(1)
			}
			err = cmd.Wait()
			log.Print("sshd wait exited ", err)
			if err == nil {
				s.Exit(0)
			} else {
				s.Exit(1)
			}
		}
	})

	err = srv.SetOption(ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		// check permissions on authorized_keys
		authKeys, err := ioutil.ReadFile("/etc/goes/sshd/authorized_keys")
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Printf("Error reading authorized keys: %s\n", err)
				return c.FailSafe
			}
			authKeys, err = ioutil.ReadFile("/etc/goes/sshd/authorized_keys.default")
			if err != nil {
				fmt.Printf("Error reading authorized keys.default: %s\n", err)
				return c.FailSafe
			}
		}

		for len(authKeys) > 0 {
			authKey, _, _, rest, err := gossh.ParseAuthorizedKey(authKeys)
			if err != nil {
				fmt.Printf("Error parsing authorized_keys: %s\n", err)
				return false
			}
			if ssh.KeysEqual(authKey, key) {
				return true
			}
			authKeys = rest
		}
		return false // No matching key found
	}))

	err = srv.SetOption(ssh.HostKeyFile("/etc/goes/sshd/id_rsa"))
	if err != nil {
		return err
	}

	go func() {
		_ = srv.ListenAndServe()
	}()

	for {
		select {
		case <-c.done:
			return nil
		}
	}
	return err
}
