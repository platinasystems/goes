// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"sync"
)

const PendingRendezvous = 4

type rendezvous struct {
	host  *tls.Conn
	guest chan *tls.Conn
}

var rendezvousPool = &sync.Pool{
	New: func() any {
		return &rendezvous{
			guest: make(chan *tls.Conn),
		}
	},
}

var hosts = struct {
	mutex      sync.RWMutex
	rendezvous map[string]chan *rendezvous
}{
	rendezvous: make(map[string]chan *rendezvous),
}

func newRendezvous(host *tls.Conn) *rendezvous {
	r := rendezvousPool.Get().(*rendezvous)
	r.host = host
	return r
}

func (r *rendezvous) Free() {
	rendezvousPool.Put(r)
}

func waitForGuest(ctx context.Context, c *tls.Conn, ski string) (
	*tls.Conn, error,
) {
	r := newRendezvous(c)
	hosts.mutex.RLock()
	ch, ok := hosts.rendezvous[ski]
	hosts.mutex.RUnlock()
	if !ok {
		hosts.mutex.Lock()
		if ch, ok = hosts.rendezvous[ski]; !ok {
			ch = make(chan *rendezvous, PendingRendezvous)
			hosts.rendezvous[ski] = ch
		}
		hosts.mutex.Unlock()
	}
	ch <- r
	select {
	case guest := <-r.guest:
		r.Free()
		return guest, nil
	case <-ctx.Done():
		return nil, context.Canceled
	}
}

func waitForHost(ctx context.Context, c *tls.Conn, ski string) (
	*tls.Conn, error,
) {
	hosts.mutex.RLock()
	ch, ok := hosts.rendezvous[ski]
	hosts.mutex.RUnlock()
	if !ok {
		hosts.mutex.Lock()
		if hosts.rendezvous == nil {
			hosts.rendezvous = make(map[string]chan *rendezvous)
		}
		if ch, ok = hosts.rendezvous[ski]; !ok {
			ch = make(chan *rendezvous, PendingRendezvous)
			hosts.rendezvous[ski] = ch
		}
		hosts.mutex.Unlock()
	}
	select {
	case r := <-ch:
		host := r.host
		r.guest <- c
		return host, nil
	case <-ctx.Done():
		return nil, context.Canceled
	}
}
