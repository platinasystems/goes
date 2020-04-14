// Copyright 2016-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package daemons

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/external/atsock"
	"github.com/platinasystems/goes/external/log"
	"github.com/platinasystems/goes/internal/prog"
)

type Daemons struct {
	mutex sync.Mutex
	goes  *goes.Goes
	rpc   *atsock.RpcServer
	done  chan struct{}
	pids  []int
	log   daemonLog

	cmdsByPid map[int]*exec.Cmd
	stopping  bool
}

func sockname() string {
	return prog.Base() + "-daemons"
}

func (d *Daemons) init() {
	d.done = make(chan struct{})
	d.cmdsByPid = make(map[int]*exec.Cmd)
	d.log.init()
	log.Tee(&d.log)

}

func (d *Daemons) start(restarts int, args ...string) {
	if len(args) < 1 {
		return
	}
	rout, wout, err := os.Pipe()
	defer func(cs string) {
		if err != nil {
			log.Print("daemon", "err", cs, ": ", err)
		}
	}(strings.Join(args, " "))
	if err != nil {
		return
	}
	rerr, werr, err := os.Pipe()
	if err != nil {
		return
	}
	p := d.goes.Fork(args...)
	p.Stdin = nil
	p.Stdout = wout
	p.Stderr = werr
	p.Dir = "/"
	p.Env = []string{
		"PATH=" + prog.Path(),
		"TERM=linux",
	}
	if err = p.Start(); err != nil {
		return
	}
	log.Print("daemon", "info", "running ", p.Process.Pid, " ", args)
	id := fmt.Sprintf("%s.%s[%d]", prog.Base(), args[0], p.Process.Pid)
	d.mutex.Lock()
	d.pids = append(d.pids, p.Process.Pid)
	d.cmdsByPid[p.Process.Pid] = p
	d.mutex.Unlock()
	go log.LinesFrom(rout, id, "info")
	go log.LinesFrom(rerr, id, "err")
	go func(p *exec.Cmd, wout, werr *os.File, args ...string) {
		if err := p.Wait(); err != nil {
			fmt.Fprintln(werr, err)
		} else {
			fmt.Fprintln(wout, "done")
		}
		if d.cmd(p.Process.Pid) != nil {
			d.del(p.Process.Pid)
			if restarts == RestartLimit {
				fmt.Fprintln(werr, "to many restarts")
			} else {
				fmt.Fprintln(werr, "restart")
				defer d.start(restarts+1, args...)
			}
		}
		wout.Sync()
		werr.Sync()
		wout.Close()
		werr.Close()
	}(p, wout, werr, args...)
}

func (d *Daemons) List(args struct{}, reply *string) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	buf := &bytes.Buffer{}
	for _, pid := range d.pids {
		p := d.cmdsByPid[pid]
		fmt.Fprintf(buf, "%d: %v\n", pid, p.Args)
	}
	*reply = buf.String()
	return nil
}

func (d *Daemons) Log(args []string, reply *string) error {
	if len(args) > 0 {
		vargs := make([]interface{}, len(args))
		for i, arg := range args {
			vargs[i] = arg
		}
		log.Print(vargs...)
	}
	*reply = d.log.String()
	return nil
}

func (d *Daemons) Start(args []string, reply *struct{}) error {
	d.start(0, args...)
	return nil
}

func (d *Daemons) Stop(pids []int, reply *struct{}) error {
	if len(pids) == 0 {
		d.mutex.Lock()
		if d.stopping {
			d.mutex.Unlock()
			return syscall.EBUSY
		}
		d.stopping = true
		log.Print("daemon", "info", "stopping")
		defer close(d.done)
		// stop all in reverse order
		pids = make([]int, len(d.pids))
		for i, pid := range d.pids {
			pids[len(pids)-i-1] = pid
		}
		d.mutex.Unlock()
	}
	return d.stop(pids)
}

func (d *Daemons) Restart(pids []int, reply *struct{}) error {
	var pargs [][]string
	d.mutex.Lock()
	if len(pids) == 0 {
		// stop all in reverse order
		pids = make([]int, len(d.pids))
		for i, pid := range d.pids {
			pids[len(pids)-i-1] = pid
		}
		// but restart in original order
		pargs = make([][]string, len(pids))
		for i, pid := range d.pids {
			p := d.cmdsByPid[pid]
			pargs[i] = make([]string, len(p.Args))
			copy(pargs[i], p.Args)
		}
	} else {
		pargs = make([][]string, len(pids))
		for i, pid := range pids {
			p := d.cmdsByPid[pid]
			pargs[i] = make([]string, len(p.Args))
			copy(pargs[i], p.Args)
		}
	}
	d.mutex.Unlock()
	if err := d.stop(pids); err != nil {
		return err
	}
	for _, args := range pargs {
		log.Print("daemon", "info", "restarting: ", args)
		d.start(0, args...)
	}
	return nil
}

func (d *Daemons) cmd(pid int) *exec.Cmd {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.cmdsByPid[pid]
}

func (d *Daemons) del(pid int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.cmdsByPid, pid)
	for i, entry := range d.pids {
		if pid == entry {
			n := copy(d.pids[i:], d.pids[i+1:])
			d.pids = d.pids[:i+n]
			break
		}
	}
}

func (d *Daemons) stop(pids []int) error {
	for _, pid := range pids {
		if p := d.cmd(pid); p != nil {
			log.Print("daemon", "info", "stopping: ", p.Args)
			d.del(pid)
			p.Process.Signal(syscall.SIGTERM)
		} else {
			return fmt.Errorf("%d: not found", pid)
		}
	}
	have := func(dn string) bool {
		_, err := os.Stat(dn)
		return err == nil
	}
	for _, pid := range pids {
		procdn := fmt.Sprint("/proc/", pid)
		for t := 100 * time.Millisecond; have(procdn); t *= 2 {
			if t > 3*time.Second {
				log.Print("daemon", "info", "killing: ", pid)
				syscall.Kill(pid, syscall.SIGKILL)
			} else {
				time.Sleep(t)
			}
		}
	}
	return nil
}
