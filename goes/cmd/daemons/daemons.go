// Copyright 2016-2016 Platina Systems, Inc. All rights reserved.
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

	"github.com/platinasystems/atsock"
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/internal/prog"
	"github.com/platinasystems/log"
)

const sockname = "goes-daemons"

type Daemons struct {
	mutex sync.Mutex
	goes  *goes.Goes
	rpc   *atsock.RpcServer
	done  chan struct{}
	pids  []int

	cmdsByPid map[int]*exec.Cmd
	stopping  bool
}

func (daemons *Daemons) start(args ...string) {
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
	daemons.pids = append(daemons.pids, p.Process.Pid)
	daemons.cmdsByPid[p.Process.Pid] = p
	daemons.mutex.Unlock()
	go log.LinesFrom(rout, id, "info")
	go log.LinesFrom(rerr, id, "err")
	go func(p *exec.Cmd, wout, werr *os.File, args ...string) {
		if err := p.Wait(); err != nil {
			fmt.Fprintln(werr, err)
		} else {
			fmt.Fprintln(wout, "done")
		}
		if daemons.cmd(p.Process.Pid) != nil {
			fmt.Fprintln(werr, "restart")
			daemons.del(p.Process.Pid)
			defer daemons.start(args...)
		}
		wout.Sync()
		werr.Sync()
		wout.Close()
		werr.Close()
	}(p, wout, werr, args...)
}

func (daemons *Daemons) List(args struct{}, reply *string) error {
	daemons.mutex.Lock()
	defer daemons.mutex.Unlock()
	buf := &bytes.Buffer{}
	for _, pid := range daemons.pids {
		p := daemons.cmdsByPid[pid]
		fmt.Fprintf(buf, "%d: %v\n", pid, p.Args)
	}
	*reply = buf.String()
	return nil
}

func (daemons *Daemons) Start(args []string, reply *struct{}) error {
	daemons.start(args...)
	return nil
}

func (daemons *Daemons) Stop(pids []int, reply *struct{}) error {
	if len(pids) == 0 {
		daemons.mutex.Lock()
		if daemons.stopping {
			daemons.mutex.Unlock()
			return syscall.EBUSY
		}
		daemons.stopping = true
		defer func() {
			log.Print("daemon", "info", "signalling done")
			close(daemons.done)
		}()
		// stop all in reverse order
		pids = make([]int, len(daemons.pids))
		for i, pid := range daemons.pids {
			pids[len(pids)-i-1] = pid
		}
		daemons.mutex.Unlock()
	}
	return daemons.stop(pids)
}

func (daemons *Daemons) Restart(pids []int, reply *struct{}) error {
	var pargs [][]string
	daemons.mutex.Lock()
	if len(pids) == 0 {
		// stop all in reverse order
		pids = make([]int, len(daemons.pids))
		for i, pid := range daemons.pids {
			pids[len(pids)-i-1] = pid
		}
		// but restart in original order
		pargs = make([][]string, len(pids))
		for i, pid := range daemons.pids {
			p := daemons.cmdsByPid[pid]
			pargs[i] = make([]string, len(p.Args))
			copy(pargs[i], p.Args)
		}
	} else {
		pargs = make([][]string, len(pids))
		for i, pid := range pids {
			p := daemons.cmdsByPid[pid]
			pargs[i] = make([]string, len(p.Args))
			copy(pargs[i], p.Args)
		}
	}
	daemons.mutex.Unlock()
	if err := daemons.stop(pids); err != nil {
		return err
	}
	for _, args := range pargs {
		log.Print("daemon", "info", "restarting: ", args)
		daemons.start(args...)
	}
	return nil
}

func (daemons *Daemons) cmd(pid int) *exec.Cmd {
	daemons.mutex.Lock()
	defer daemons.mutex.Unlock()
	return daemons.cmdsByPid[pid]
}

func (daemons *Daemons) del(pid int) {
	daemons.mutex.Lock()
	defer daemons.mutex.Unlock()
	delete(daemons.cmdsByPid, pid)
	for i, entry := range daemons.pids {
		if pid == entry {
			n := copy(daemons.pids[i:], daemons.pids[i+1:])
			daemons.pids = daemons.pids[:i+n]
			break
		}
	}
}

func (daemons *Daemons) stop(pids []int) error {
	for _, pid := range pids {
		if p := daemons.cmd(pid); p != nil {
			log.Print("daemon", "info", "stopping: ", p.Args)
			daemons.del(pid)
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
