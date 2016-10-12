// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package internal

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"syscall"
)

type daemons struct {
	mutex sync.Mutex
	byPid map[int]string
}

type Killaller interface {
	Killall()
}

func (p *daemons) Hdel(key, subkey string, subkeys ...string) (int, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	killed := 0
	if p.byPid == nil {
		return killed, nil
	}
	for _, spid := range append([]string{subkey}, subkeys...) {
		var pid int
		_, err := fmt.Sscan(spid, &pid)
		if err != nil {
			return killed, err
		}
		if arg0, found := p.byPid[pid]; found {
			err = syscall.Kill(pid, syscall.SIGTERM)
			if err != nil {
				return killed, err
			}
			killed += 1
			fmt.Println(arg0, "[", pid, "]: killed")
		}
	}
	return killed, nil
}

func (p *daemons) Hget(key, subkey string) ([]byte, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	var pid int
	if p.byPid == nil {
		return nil, nil
	}
	_, err := fmt.Sscan(subkey, &pid)
	if err != nil {
		return nil, err
	}
	if arg0, found := p.byPid[pid]; found {
		return []byte(arg0), nil
	}
	return nil, nil
}

func (p *daemons) Hgetall(key string) ([][]byte, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	pids := make([]int, 0, len(p.byPid))
	reply := make([][]byte, 0, 2*len(pids))
	for pid := range p.byPid {
		pids = append(pids, pid)
	}
	sort.Ints(pids)
	for _, pid := range pids {
		reply = append(reply, []byte(fmt.Sprint(pid)))
		reply = append(reply, []byte(p.byPid[pid]))
	}
	pids = pids[:0]
	return reply, nil
}

func (p *daemons) Hkeys(key string) ([][]byte, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	pids := make([]int, 0, len(p.byPid))
	for pid := range p.byPid {
		pids = append(pids, pid)
	}
	sort.Ints(pids)
	reply := make([][]byte, 0, len(pids))
	for _, pid := range pids {
		reply = append(reply, []byte(fmt.Sprint(pid)))
	}
	return reply, nil
}

func (p *daemons) Hset(key, subkey string, value []byte) (int, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.byPid == nil {
		p.byPid = make(map[int]string)
	}

	var pid int
	if _, err := fmt.Sscan(subkey, &pid); err != nil {
		return -1, err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return -1, err
	}
	ret := 1
	if _, found := p.byPid[pid]; found {
		ret = 0
	}
	arg0 := string(value)
	p.byPid[pid] = arg0
	fmt.Println("registered ", arg0, "[", pid, "]")
	if os.Getpid() == 1 {
		go p.wait(arg0, proc)
	}
	return ret, nil
}

func (p *daemons) Killall() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.byPid == nil {
		return
	}
	for pid, arg0 := range p.byPid {
		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			fmt.Fprintln(os.Stderr, "kill ", arg0, "[", pid, "]: ",
				err)
		} else {
			fmt.Println("killed ", arg0, "[", pid, "]")
		}
		delete(p.byPid, pid)
	}
}

func (p *daemons) wait(arg0 string, proc *os.Process) {
	pid := proc.Pid
	_, err := proc.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, arg0, "[", pid, "]: ", err)
	} else {
		fmt.Println(arg0, "[", pid, "]: terminated")
	}

	p.mutex.Lock()
	delete(p.byPid, pid)
	p.mutex.Unlock()
}
