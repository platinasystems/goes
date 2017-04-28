// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This telnet daemon is only run from an embedded machine's /init, not
// /usr/bin/goes start
package telnetd

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/kr/pty"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/telnet/command"
	"github.com/platinasystems/go/internal/telnet/option"
)

const (
	Name    = "telnetd"
	Apropos = "FIXME"
	Usage   = "telnetd"
)

type Interface interface {
	Apropos() lang.Alt
	Kind() goes.Kind
	Main(...string) error
	String() string
	Usage() string
}

func New() Interface { return cmd{} }

type cmd struct{}

func (cmd) Apropos() lang.Alt { return apropos }
func (cmd) Kind() goes.Kind   { return goes.Daemon }
func (cmd) String() string    { return Name }
func (cmd) Usage() string     { return Usage }

func (cmd) Main(args ...string) error {
	ln, err := net.Listen("tcp", ":23")
	if err != nil {
		return err
	}
	defer ln.Close()
	signal.Ignore(syscall.SIGCHLD)
	dir := "/"
	if fi, err := os.Stat("/root"); err == nil && fi.IsDir() {
		dir = "/root"
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		optionNegotiation(conn)

		pts, tty, err := pty.Open()
		if err != nil {
			return err
		}
		proc, err := os.StartProcess("/bin/goes",
			[]string{"goes"},
			&os.ProcAttr{
				Dir: dir,
				Env: []string{
					"PATH=/usr/bin:/bin",
					"TERM=xterm",
				},
				Files: []*os.File{
					tty,
					tty,
					tty,
				},
				Sys: &syscall.SysProcAttr{
					Setsid:  true,
					Setctty: true,
					Ctty:    0,
					Pgid:    0,
				},
			})
		tty.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			pts.Close()
			continue
		}
		go func() { io.Copy(conn, pts) }()
		go nullFilteredCopy(pts, conn)
		go func() {
			proc.Wait()
			pts.Close()
			conn.Close()
		}()
	}
	return nil
}

func optionNegotiation(conn net.Conn) error {
	ack := make([]byte, 1024)
	conn.Write([]byte{command.IAC, command.WILL, option.ECHO})
	n, err := conn.Read(ack)
	if err != nil {
		return err
	}
	if 0 != bytes.Compare(ack[:n],
		[]byte{command.IAC, command.DO, option.ECHO}) {
		conn.Write([]byte("telnet WILL ECHO Failed"))
		return nil
	}
	conn.Write([]byte{
		command.IAC, command.DONT, option.ECHO,
		command.IAC, command.WILL, option.SGA,
	})
	n, err = conn.Read(ack)
	if err != nil {
		return err
	}
	if 0 != bytes.Compare(ack[:n],
		[]byte{command.IAC, command.DO, option.SGA}) {
		conn.Write([]byte("telnet WILL SGA failed"))
		return nil
	}
	conn.Write([]byte{command.IAC, command.DO, option.SGA})
	n, err = conn.Read(ack)
	if err != nil {
		return err
	}
	if 0 != bytes.Compare(ack[:n],
		[]byte{command.IAC, command.WILL, option.SGA}) {
		conn.Write([]byte("telnet DO SGA failed"))
		return nil
	}
	conn.Write([]byte{command.IAC, command.WONT, option.LINEMODE})
	return nil
}

func nullFilteredCopy(w io.Writer, r io.Reader) {
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if err != nil {
			return
		}
		if n <= 0 {
			continue
		}
		if buf[n-1] == 0 {
			n -= 1
		}
		if _, err = w.Write(buf[:n]); err != nil {
			return
		}
	}
}

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
