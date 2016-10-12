// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package internal

import (
	"sort"
	"sync"

	"github.com/platinasystems/go/redis"
)

type standby struct {
	sync.Mutex
	m map[string]struct{}
}

func (p *standby) Hdel(key, subkey string, subkeys ...string) (int, error) {
	p.Lock()
	defer p.Unlock()

	removed := 0
	if p.m == nil {
		return removed, nil
	}
	for _, k := range append([]string{subkey}, subkeys...) {
		if _, found := p.m[k]; found {
			delete(p.m, k)
			removed++
		}
	}
	return removed, nil
}

func (p *standby) Hexists(key, subkey string) (int, error) {
	p.Lock()
	defer p.Unlock()
	ret := 0
	if p.m == nil {
		return ret, nil
	}
	if _, found := p.m[subkey]; found {
		ret = 1
	}
	return ret, nil
}

func (p *standby) Hkeys(key string) ([][]byte, error) {
	p.Lock()
	defer p.Unlock()
	if p.m == nil {
		return nil, nil
	}
	reply := redis.MakeBB(len(p.m))
	for k := range p.m {
		reply = append(reply, []byte(k))
	}
	sort.Sort(reply)
	return reply.Redis(), nil
}

func (p *standby) Hset(key, subkey string, value []byte) (int, error) {
	p.Lock()
	defer p.Unlock()

	if p.m == nil {
		p.m = make(map[string]struct{})
	}
	p.m[subkey] = struct{}{}
	return 1, nil
}
