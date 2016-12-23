// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package daemons starts redisd followed by all other configured daemons.
package daemons

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
)

const Name = "goes-daemons"

func New() *cmd { return new(cmd) }

type cmd goes.ByName

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (*cmd) Kind() goes.Kind { return goes.Hidden }
func (*cmd) String() string  { return Name }

func (*cmd) Usage() string { return "goes-daemons [OPTIONS]..." }

func (c *cmd) Main(args ...string) (err error) {
	var daemons int
	done := make(chan struct{})

	defer func() {
		if err != nil {
			log.Print("daemon", "err", err)
		} else {
			log.Print("daemon", "info", "done")
		}
	}()

	signal.Ignore(syscall.SIGTERM)

	err = c.daemon(done, "redisd", args...)
	if err != nil {
		return
	}
	daemons++

	err = redis.Hwait(redis.DefaultHash, "redis.ready", "true",
		10*time.Second)
	if err != nil {
		return
	}

	for name, g := range goes.ByName(*c) {
		if g.Kind.IsDaemon() && g.Name != "redisd" {
			err := c.daemon(done, name)
			if err != nil {
				log.Print("daemon", "err", g.Name, ": ", err)
			} else {
				daemons++
			}
		}
	}

	for i := 0; i < daemons; i++ {
		<-done
	}
	return
}

func (c *cmd) daemon(done chan<- struct{}, arg0 string, args ...string) error {
	rout, wout, err := os.Pipe()
	if err != nil {
		return err
	}
	rerr, werr, err := os.Pipe()
	if err != nil {
		return err
	}
	d := exec.Command(goes.Prog(), args...)
	d.Args[0] = arg0
	d.Stdin = nil
	d.Stdout = wout
	d.Stderr = werr
	d.Dir = "/"
	d.Env = []string{
		"PATH=" + goes.Path(),
		"TERM=linux",
	}
	err = d.Start()
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s.%s[%d]", goes.ProgBase(), arg0, d.Process.Pid)
	go log.LinesFrom(rout, id, "info")
	go log.LinesFrom(rerr, id, "err")
	go func(d *exec.Cmd, wout, werr *os.File, done chan<- struct{}) {
		if err := d.Wait(); err != nil {
			fmt.Fprintln(werr, err)
		} else {
			fmt.Fprintln(wout, "done")
		}
		wout.Sync()
		werr.Sync()
		wout.Close()
		werr.Close()
		done <- struct{}{}
	}(d, wout, werr, done)
	return nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "start redisd then all other daemons",
	}
}
