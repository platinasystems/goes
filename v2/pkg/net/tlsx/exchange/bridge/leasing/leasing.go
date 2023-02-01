// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package leasing

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sync"
)

var (
	ErrNoNetwork = errors.New("missing <network>")
	ErrNoBase    = errors.New("missing <base> lease address")
)

type T struct {
	Enabled bool
	sync.RWMutex
	prefix  netip.Prefix
	next    netip.Addr
	address map[string]netip.Addr
	tenant  map[netip.Addr]string
}

func (t *T) Configure(prefix netip.Prefix) {
	t.prefix = prefix
	t.next = prefix.Addr().Next()
	t.address = make(map[string]netip.Addr)
	t.tenant = make(map[netip.Addr]string)
	t.Enabled = true
}

func (t *T) Lease(tenant string) netip.Prefix {
	t.Lock()
	defer t.Unlock()
	addr, ok := t.address[tenant]
	if !ok {
		for {
			addr = t.next
			t.next = t.next.Next()
			if _, occupied := t.tenant[addr]; !occupied {
				break
			}
			if t.prefix.Contains(t.next) {
				t.next = t.prefix.Addr().Next()
			}
		}
		t.address[tenant] = addr
		t.tenant[addr] = tenant
	}
	return netip.PrefixFrom(addr, t.prefix.Bits())
}

func (t *T) MarshalJSON() ([]byte, error) {
	t.RLock()
	defer t.RUnlock()
	return json.Marshal(t.address)
}

func (t *T) MarshalText() ([]byte, error) {
	t.RLock()
	defer t.RUnlock()
	buf := new(bytes.Buffer)
	for k, v := range t.address {
		fmt.Fprint(buf, k, ": ", v, "\n")
	}
	return buf.Bytes(), nil
}

func (t *T) Occupy(tenant, arg string) error {
	t.Lock()
	defer t.Unlock()
	addr, err := netip.ParseAddr(arg)
	if err != nil {
		return err
	}
	if !t.prefix.Contains(addr) {
		return fmt.Errorf("network %v doesn't contain %v",
			t.prefix, addr)
	}
	if occupant, occupied := t.tenant[addr]; occupied {
		if occupant != tenant {
			return fmt.Errorf("%v occupied by %s", addr, occupant)
		}
	} else {
		t.address[tenant] = addr
		t.tenant[addr] = tenant
	}
	return nil
}
