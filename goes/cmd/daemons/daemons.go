// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package daemons starts redisd followed by all other configured daemons.
package daemons

import (
	"bytes"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/atsock"
	"github.com/platinasystems/log"
	"github.com/platinasystems/go/internal/prog"
)

type Command struct {
	// Machines list goes command + args for daemons that run from start,
	// including redisd.  Note that dependent daemons should wait on a
	// respective redis key, e.g.
	//	redis.Hwait(redis.DefaultHash, "redis.ready", "true", TIMEOUT)
	// or
	//	redis.IsReady()
	Init [][]string
	Daemons
}

type Daemons struct {
	mutex sync.Mutex
	goes  *goes.Goes
	rpc   *atsock.RpcServer
	done  chan struct{}

	cmdsByPid map[int]*exec.Cmd
}

func (*Command) String() string { return "goes-daemons" }

func (*Command) Usage() string {
	return "goes-daemons [OPTIONS]..."
}

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "start daemons and wait for their exit",
	}
}

func (c *Command) Goes(g *goes.Goes) { c.Daemons.goes = g }

func (*Command) Kind() cmd.Kind { return cmd.Hidden }

func (c *Command) Main(args ...string) error {
	if len(args) == 0 {
		return c.server()
	}
	cl, err := atsock.NewRpcClient("daemons")
	if err != nil {
		return err
	}
	defer cl.Close()
	empty := struct{}{}
	switch args[0] {
	case "list":
		var s string
		err = cl.Call("Daemons.List", struct{}{}, &s)
		if err == nil {
			os.Stdout.WriteString(s)
		}
	case "start":
		if len(args) < 2 {
			err = fmt.Errorf("missing COMMAND [ARG]...")
		} else {
			err = cl.Call("Daemons.Start", args[1:], &empty)
		}
	case "stop", "restart":
		var pid int
		if len(args) < 2 {
			err = fmt.Errorf("missing PID")
		} else if len(args) > 2 {
			err = fmt.Errorf("%v: unexpected", args[2:])
		} else if _, err = fmt.Sscan(args[1], &pid); err == nil {
			method := map[string]string{
				"stop":    "Daemons.Stop",
				"restart": "Daemons.Restart",
			}[args[0]]
			err = cl.Call(method, pid, &empty)
		}
	default:
		err = fmt.Errorf("%s: unknown", args[0])
	}
	return err
}

func (c *Command) server() (err error) {
	c.Daemons.done = make(chan struct{})
	c.Daemons.cmdsByPid = make(map[int]*exec.Cmd)

	signal.Ignore(syscall.SIGTERM)

	c.rpc, err = atsock.NewRpcServer("daemons")
	if err != nil {
		return
	}
	defer c.rpc.Close()

	for _, dargs := range c.Init {
		c.Daemons.start(dargs...)
	}

	rpc.Register(&c.Daemons)

	for n := -1; n != 0; {
		<-c.Daemons.done
		time.Sleep(5 * time.Second)
		c.Daemons.mutex.Lock()
		n = len(c.Daemons.cmdsByPid)
		c.Daemons.mutex.Unlock()
	}

	log.Print("daemon", "info", "done")
	return
}

func (daemons *Daemons) start(args ...string) {
	cs := strings.Join(args, " ")
	rout, wout, err := os.Pipe()
	defer func() {
		if err != nil {
			log.Print("daemon", "err", cs, ": ", err)
		}
	}()
	if err != nil {
		return
	}
	rerr, werr, err := os.Pipe()
	if err != nil {
		return
	}
	p := daemons.goes.Fork(args...)
	p.Stdin = nil
	p.Stdout = wout
	p.Stderr = werr
	p.Dir = "/"
	p.Env = []string{
		"PATH=" + prog.Path(),
		"TERM=linux",
	}
	err = p.Start()
	if err != nil {
		return
	}
	id := fmt.Sprintf("%s.%s[%d]", prog.Base(), args[0], p.Process.Pid)
	daemons.mutex.Lock()
	daemons.cmdsByPid[p.Process.Pid] = p
	daemons.mutex.Unlock()
	go log.LinesFrom(rout, id, "info")
	go log.LinesFrom(rerr, id, "err")
	go func(p *exec.Cmd, wout, werr *os.File) {
		if err := p.Wait(); err != nil {
			fmt.Fprintln(werr, err)
		} else {
			fmt.Fprintln(wout, "done")
		}
		wout.Sync()
		werr.Sync()
		wout.Close()
		werr.Close()
		daemons.mutex.Lock()
		delete(daemons.cmdsByPid, p.Process.Pid)
		daemons.mutex.Unlock()
		daemons.done <- struct{}{}
	}(p, wout, werr)
}

func (daemons *Daemons) List(args struct{}, reply *string) error {
	buf := &bytes.Buffer{}
	for k, v := range daemons.cmdsByPid {
		fmt.Fprintf(buf, "%d: %v\n", k, v.Args)
	}
	*reply = buf.String()
	return nil
}

func (daemons *Daemons) Start(args []string, reply *struct{}) error {
	daemons.start(args...)
	return nil
}

func (daemons *Daemons) Stop(pid int, reply *struct{}) error {
	daemons.mutex.Lock()
	defer daemons.mutex.Unlock()
	var err error
	if p, found := daemons.cmdsByPid[pid]; !found {
		err = fmt.Errorf("%d: not found", pid)
	} else {
		err = p.Process.Signal(syscall.SIGTERM)
		time.Sleep(1 * time.Second)
		err = p.Process.Kill()
		if err.Error() == "os: process already finished" {
			err = nil
		}
	}
	return err
}

func (daemons *Daemons) Restart(pid int, reply *struct{}) error {
	var err error
	daemons.mutex.Lock()
	p, found := daemons.cmdsByPid[pid]
	daemons.mutex.Unlock()
	if !found {
		err = fmt.Errorf("%d: not found", pid)
	} else {
		args := p.Args
		err = p.Process.Signal(syscall.SIGTERM)
		time.Sleep(1 * time.Second)
		err = p.Process.Kill()
		if err.Error() == "os: process already finished" {
			err = nil
		}
		if err == nil {
			fmt.Fprintf(os.Stderr, "Calling daemons.start: %v\n",
				args)
			daemons.start(args...)
		}
	}
	return err
}
