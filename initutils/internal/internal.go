// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// This is common to /sbin/init and /usr/bin/goesd
package internal

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/group"
)

const (
	RunGoes          = "/run/goes"
	RunGoesPids      = RunGoes + "/pids"
	RunGoesPidsGoesd = RunGoesPids + "/goesd"
	EtcGoesGoesd     = "/etc/goes/goesd"
)

var (
	Init common

	PidFlags = syscall.O_RDWR | syscall.O_CREAT | syscall.O_TRUNC
	PidPerms = os.FileMode(0644)
	Hook     = func() error { return nil }
	// Machines should set StartDaemons to goes.Command.Daemons().Start()
	StartDaemons = func() error { return nil }
	ErrNotRoot   = errors.New("you aren't root")
)

type common struct {
	Redisd Redisd
	Reg    RedisReg
}

func (p *common) Start() (err error) {
	defer func() {
		if err != nil {
			if p.Reg.Srvr != nil {
				p.Reg.Srvr.Terminate()
			}
		}
	}()
	if os.Geteuid() != 0 {
		err = ErrNotRoot
		return
	}

	syscall.Umask(0002)

	adm := group.Parse()["adm"].Gid()

	_, err = os.Stat(RunGoesPids)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(RunGoesPids, os.FileMode(0755))
			if adm > 0 {
				os.Chown(RunGoes, os.Geteuid(), adm)
				os.Chown(RunGoesPids, os.Geteuid(), adm)
			}
		}
		if err != nil {
			return
		}
	}

	if pid := os.Getpid(); pid > 1 {
		var f *os.File
		f, err = os.OpenFile(RunGoesPidsGoesd, PidFlags, PidPerms)
		if err != nil {
			return
		} else {
			if adm > 0 {
				f.Chown(os.Geteuid(), adm)
			}
			fmt.Fprintln(f, pid)
			f.Close()
		}
	}

	Init.assign("cmdline", &cmdline{})
	Init.assign("daemons", &daemons{})
	Init.assign("standby", &standby{})

	// FIXME redisd appears to be writing to a closed FD
	signal.Ignore(syscall.SIGPIPE)
	rpc.Register(&p.Reg)
	if err = p.Reg.main(); err != nil {
		return
	}
	if err = p.Redisd.main(); err != nil {
		return
	}
	if err = Hook(); err != nil {
		return
	}
	if _, err = os.Stat(EtcGoesGoesd); err == nil {
		err = command.Main("source", EtcGoesGoesd)
		if err != nil {
			return err
		}
	}
	if err = StartDaemons(); err != nil {
		return
	}
	names := os.Getenv("REDISD_DEVS")
	if len(names) == 0 {
		names = "lo"
	}
	for _, name := range strings.Fields(names) {
		dev, err := net.InterfaceByName(name)
		if err == nil && (dev.Flags&net.FlagUp) == net.FlagUp {
			p.Redisd.listen(name, "")
		}
	}
	return
}

func (p *common) redisdListen(dev *net.Interface) {
	if _, err := p.Redisd.listen(dev.Name, ""); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func (p *common) assign(key string, v interface{}) {
	p.Redisd.mutex.Lock()
	defer p.Redisd.mutex.Unlock()
	p.Redisd.assignments = p.Redisd.assignments.Insert(key, v)
	p.Redisd.flushKeyCache()
}

func (p *common) unassign(key string) {
	p.Redisd.mutex.Lock()
	defer p.Redisd.mutex.Unlock()
	p.Redisd.assignments = p.Redisd.assignments.Delete(key)
	p.Redisd.flushKeyCache()
}

// Use this to start the redis server listening on the given device.
func RedisdListen(dev *net.Interface) { Init.redisdListen(dev) }
