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
	provider *tls.Conn
	consumer chan *tls.Conn
}

var rendezvousPool = &sync.Pool{
	New: func() any {
		return &rendezvous{
			consumer: make(chan *tls.Conn),
		}
	},
}

var providers = struct {
	mutex      sync.RWMutex
	rendezvous map[string]chan *rendezvous
}{
	rendezvous: make(map[string]chan *rendezvous),
}

func newRendezvous(provider *tls.Conn) *rendezvous {
	r := rendezvousPool.Get().(*rendezvous)
	r.provider = provider
	return r
}

func (r *rendezvous) Free() {
	rendezvousPool.Put(r)
}

func WaitForConsumer(ctx context.Context, c *tls.Conn, ski string) (
	*tls.Conn, error,
) {
	r := newRendezvous(c)
	providers.mutex.RLock()
	ch, ok := providers.rendezvous[ski]
	providers.mutex.RUnlock()
	if !ok {
		providers.mutex.Lock()
		if ch, ok = providers.rendezvous[ski]; !ok {
			ch = make(chan *rendezvous, PendingRendezvous)
			providers.rendezvous[ski] = ch
		}
		providers.mutex.Unlock()
	}
	ch <- r
	select {
	case consumer := <-r.consumer:
		r.Free()
		return consumer, nil
	case <-ctx.Done():
		return nil, context.Canceled
	}
}

func WaitForProvider(ctx context.Context, c *tls.Conn, ski string) (
	*tls.Conn, error,
) {
	providers.mutex.RLock()
	ch, ok := providers.rendezvous[ski]
	providers.mutex.RUnlock()
	if !ok {
		providers.mutex.Lock()
		if providers.rendezvous == nil {
			providers.rendezvous = make(map[string]chan *rendezvous)
		}
		if ch, ok = providers.rendezvous[ski]; !ok {
			ch = make(chan *rendezvous, PendingRendezvous)
			providers.rendezvous[ski] = ch
		}
		providers.mutex.Unlock()
	}
	select {
	case r := <-ch:
		provider := r.provider
		r.consumer <- c
		return provider, nil
	case <-ctx.Done():
		return nil, context.Canceled
	}
}
