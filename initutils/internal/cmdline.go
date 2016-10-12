// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package internal

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"sort"
	"sync"

	redis_server "github.com/platinasystems/go-redis-server"
)

type cmdline struct {
	sync.Mutex
	m redis_server.HashValue
	k []string
}

func (p *cmdline) load() error {
	if p.m != nil {
		return nil
	}
	bf, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return err
	}
	clre, err := regexp.Compile("\\S+='.+'|\\S+=\".+\"|\\S+=\\S+|\\S+")
	if err != nil {
		return err
	}
	p.m = make(redis_server.HashValue)
	for _, bl := range bytes.Split(bf, []byte{'\n'}) {
		if len(bl) == 0 {
			continue
		}
		for _, b := range clre.FindAll(bl, -1) {
			eq := bytes.Index(b, []byte{'='})
			if eq <= 1 {
				p.m[string(b)] = []byte("true")
				continue
			}
			name := string(b[:eq])
			var value []byte
			if b[eq+1] == '\'' || b[eq+1] == '"' {
				value = b[eq+2 : len(b)-1]
			} else {
				value = b[eq+1:]
			}
			p.m[name] = value
		}
	}
	p.k = make([]string, 0, len(p.m))
	for k := range p.m {
		p.k = append(p.k, k)
	}
	sort.Strings(p.k)
	return nil
}

func (p *cmdline) Hexists(key, subkey string) (int, error) {
	p.Lock()
	defer p.Unlock()
	ret := 0
	if err := p.load(); err != nil {
		return ret, err
	}
	if _, found := p.m[subkey]; found {
		ret = 1
	}
	return ret, nil
}

func (p *cmdline) Hget(key, subkey string) ([]byte, error) {
	p.Lock()
	defer p.Unlock()
	if err := p.load(); err != nil {
		return nil, err
	}
	if b, found := p.m[subkey]; found {
		return b, nil
	}
	return nil, nil
}

func (p *cmdline) Hgetall(key string) ([][]byte, error) {
	p.Lock()
	defer p.Unlock()
	if err := p.load(); err != nil {
		return nil, err
	}
	reply := make([][]byte, 0, 2*len(p.m))
	for k, v := range p.m {
		reply = append(reply, []byte(k))
		reply = append(reply, v)
	}
	return reply, nil
}

func (p *cmdline) Hkeys(key string) ([][]byte, error) {
	p.Lock()
	defer p.Unlock()
	if err := p.load(); err != nil {
		return nil, err
	}
	reply := make([][]byte, 0, len(p.k))
	for _, k := range p.k {
		reply = append(reply, []byte(k))
	}
	return reply, nil
}
