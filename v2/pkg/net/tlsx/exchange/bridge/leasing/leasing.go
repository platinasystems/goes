// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package leasing

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
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
	network netip.Prefix
	base,
	next netip.Addr
	address map[string]netip.Addr
	tenant  map[netip.Addr]string
}

func (t *T) Configure(args []string) ([]string, error) {
	switch len(args) {
	case 0:
		return args, ErrNoNetwork
	case 1:
		return args, ErrNoBase
	}
	if args[0] == "-h" {
		return args, flag.ErrHelp
	}
	var err error
	t.network, err = netip.ParsePrefix(args[0])
	if err != nil {
		return args[1:], fmt.Errorf("<network>: %w", err)
	}
	t.base, err = netip.ParseAddr(args[1])
	if err != nil {
		return args[2:], fmt.Errorf("<base>: %w", err)
	}
	if !t.network.Contains(t.base) {
		return args[2:], fmt.Errorf("%v doesn't contain %v",
			t.network, t.base)
	}
	t.next = t.base
	t.address = make(map[string]netip.Addr)
	t.tenant = make(map[netip.Addr]string)
	t.Enabled = true
	return args[2:], nil
}

func (t *T) Lease(tenant string) netip.Prefix {
	t.Lock()
	defer t.Unlock()
	addr, ok := t.address[tenant]
	if !ok {
		addr = t.next
		t.next = t.next.Next()
		t.address[tenant] = addr
		t.tenant[addr] = tenant
	}
	return netip.PrefixFrom(addr, t.network.Bits())
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
	if !t.network.Contains(addr) {
		return fmt.Errorf("network %v doesn't contain %v",
			t.network, addr)
	}
	if addr.Compare(t.base) >= 0 {
		return fmt.Errorf("%v >= base @ %v", addr, t.base)
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
