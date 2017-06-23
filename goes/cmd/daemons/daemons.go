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

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/go/internal/redis"
)

const (
	Name    = "goes-daemons"
	Apropos = "start redisd then all other daemons"
	Usage   = "goes-daemons [OPTIONS]..."
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() *Command { return new(Command) }

type Command struct {
	g *goes.Goes
}

func (*Command) Apropos() lang.Alt   { return apropos }
func (c *Command) Goes(g *goes.Goes) { c.g = g }
func (*Command) Kind() cmd.Kind      { return cmd.Hidden }

func (c *Command) Main(args ...string) (err error) {
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

	err = c.daemon(done, append([]string{"redisd"}, args...)...)
	if err != nil {
		return
	}
	daemons++

	err = redis.Hwait(redis.DefaultHash, "redis.ready", "true",
		10*time.Second)
	if err != nil {
		return
	}

	for _, name := range c.g.Names {
		v := c.g.ByName(name)
		k := cmd.WhatKind(v)
		if k.IsDaemon() && name != "redisd" {
			err := c.daemon(done, name)
			if err != nil {
				log.Print("daemon", "err", name, ": ", err)
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

func (*Command) String() string { return Name }
func (*Command) Usage() string  { return Usage }

func (c *Command) daemon(done chan<- struct{}, args ...string) error {
	rout, wout, err := os.Pipe()
	if err != nil {
		return err
	}
	rerr, werr, err := os.Pipe()
	if err != nil {
		return err
	}
	d := c.g.Fork(args...)
	d.Stdin = nil
	d.Stdout = wout
	d.Stderr = werr
	d.Dir = "/"
	d.Env = []string{
		"PATH=" + prog.Path(),
		"TERM=linux",
	}
	err = d.Start()
	if err != nil {
		return err
	}
	id := fmt.Sprintf("%s.%s[%d]", prog.Base(), args[0], d.Process.Pid)
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
