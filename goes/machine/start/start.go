// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package start provides the named command that runs a redis server followed
// by all of the configured daemons. If the PID is 1, start doesn't return;
// instead, it iterates and command shell.
package start

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/internal/cmdline"
	"github.com/platinasystems/go/goes/internal/parms"
	"github.com/platinasystems/go/goes/machine/internal"
	"github.com/platinasystems/go/goes/sockfile"
	"github.com/platinasystems/go/redis"
	. "github.com/platinasystems/go/version"
)

// Machines may use Hook to run something before redisd and other daemons.
var Hook = func() error { return nil }

// Machines may use PubHook to publish redis "key: value" strings before any
// daemons are run.
var PubHook = func(chan<- string) error { return nil }

// Machines may use ConfHook to run something after all daemons start and
// before source of config..
var ConfHook = func() error { return nil }

// A non-empty Machine is published to redis as "machine: Machine"
var Machine string

var RedisDevs []string

func New() *goes.Goes {
	cmd := new(cmd)
	return &goes.Goes{
		Name:   "start",
		ByName: cmd.ByName,
		Main:   cmd.Main,
		Usage:  "start [OPTION]...",
		Apropos: map[string]string{
			"en_US.UTF-8": "start this goes machine",
		},
		Man: map[string]string{
			"en_US.UTF-8": `NAME
	start - start this goes machine

SYNOPSIS
	start [-conf=URL] [REDIS OPTIONS]...

DESCRIPTION
	Start a redis server followed by the machine and its embedded daemons.

OPTIONS
	-conf URL
		Specifies the URL of the machine's configuration script that's
		sourced immediately after start of all daemons.

SEE ALSO
	redisd`,
		},
	}
}

type cmd struct {
	byName goes.ByName
}

func (cmd *cmd) ByName(byName goes.ByName) {
	cmd.byName = byName
}

func (cmd *cmd) Main(args ...string) error {
	parm, args := parms.New(args, "-conf")
	redisd := []string{"redisd"}
	if len(args) > 0 {
		redisd = append(redisd, args...)
	} else if len(RedisDevs) > 0 {
		redisd = append(redisd, RedisDevs...)
	} else if itfs, err := net.Interfaces(); err == nil {
		for _, itf := range itfs {
			redisd = append(redisd, itf.Name)
		}
	}
	err := internal.AssertRoot()
	if err != nil {
		return err
	}
	_, err = os.Stat(sockfile.Path("redisd"))
	if err == nil {
		return fmt.Errorf("already started")
	}
	if err = Hook(); err != nil {
		return err
	}
	if err = cmd.byName.Main(redisd...); err != nil {
		return err
	}
	pub, err := redis.Publish(redis.Machine)
	if err != nil {
		return err
	}
	defer close(pub)
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	pub <- fmt.Sprint("hostname: ", hostname)
	pub <- fmt.Sprint("version: ", Version)
	if len(Machine) > 0 {
		pub <- fmt.Sprint("machine: ", Machine)
	}
	keys, cl, err := cmdline.New()
	if err != nil {
		return err
	}
	for _, k := range keys {
		pub <- fmt.Sprintf("cmdline.%s: %s", k, cl[k])
	}
	if err = PubHook(pub); err != nil {
		return err
	}
	for name, g := range cmd.byName {
		if g.Kind == goes.Daemon {
			if err = cmd.byName.Main(name); err != nil {
				return err
			}
		}
	}
	if s := parm["-conf"]; len(s) > 0 {
		if err = ConfHook(); err != nil {
			return err
		}
		if err = cmd.byName.Main("source", s); err != nil {
			return err
		}
	}
	if os.Getpid() == 1 {
		_, login := cmd.byName["login"]
		for {
			if login {
				err = cmd.byName.Main("login")
				if err != nil {
					fmt.Println("login:", err)
					time.Sleep(3 * time.Second)
					continue
				}
			}
			err = cmd.byName.Main("cli")
			if err != nil && err != io.EOF {
				fmt.Println(err)
				<-make(chan struct{})
			}
		}
	}
	return nil
}
