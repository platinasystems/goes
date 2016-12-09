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

const Name = "start"

// Machines may use Hook to run something before redisd and other daemons.
var Hook = func() error { return nil }

// Machines may use PubHook to publish redis "key: value" strings before any
// daemons are run.
var PubHook = func(chan<- string) error { return nil }

// Machines may use ConfHook to run something after all daemons start and
// before source of start command script.
var ConfHook = func() error { return nil }

// A non-empty Machine is published to redis as "machine: Machine"
var Machine string

var RedisDevs []string

func New() *cmd { return new(cmd) }

type cmd goes.ByName

func (*cmd) String() string { return Name }
func (*cmd) Usage() string  { return "start [OPTION]..." }

func (c *cmd) ByName(byName goes.ByName) { *c = cmd(byName) }

func (c *cmd) Main(args ...string) error {
	byName := goes.ByName(*c)
	parm, args := parms.New(args, "-start", "-stop")
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
	if err = byName.Main(redisd...); err != nil {
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
	for name, g := range byName {
		if g.Kind == goes.Daemon && g.Name != "redisd" {
			if err = byName.Main(name); err != nil {
				return err
			}
		}
	}
	start := parm["-start"]
	if len(start) == 0 {
		if _, xerr := os.Stat("/etc/goes/start"); xerr == nil {
			start = "/etc/goes/start"
		}
	}
	if len(start) > 0 {
		if err = ConfHook(); err != nil {
			return err
		}
		if err = byName.Main("source", start); err != nil {
			return err
		}
	}
	if os.Getpid() == 1 {
		_, login := byName["login"]
		for {
			if login {
				err = byName.Main("login")
				if err != nil {
					fmt.Println("login:", err)
					time.Sleep(3 * time.Second)
					continue
				}
			}
			err = byName.Main("cli")
			if err != nil && err != io.EOF {
				fmt.Println(err)
				<-make(chan struct{})
			}
		}
	}
	return nil
}

func (c *cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "start this goes machine",
	}
}

func (c *cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	start - start this goes machine

SYNOPSIS
	start [-start=URL] [REDIS OPTIONS]...

DESCRIPTION
	Start a redis server followed by the machine and its embedded daemons.

OPTIONS
	-start URL
		Specifies the URL of the machine's configuration script that's
		sourced immediately after start of all daemons.
		default: /etc/goes/start

SEE ALSO
	redisd`,
	}
}
