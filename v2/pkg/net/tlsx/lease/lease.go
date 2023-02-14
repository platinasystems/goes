// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package lease

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

var (
	ErrInvalid     = errors.New("invalid")
	ErrOccupied    = errors.New("occupied")
	ErrUnavailable = errors.New("unavailable")
)

var (
	mutex     sync.RWMutex
	prefix, z netip.Prefix
	next      netip.Addr
	bits      int
	addressOf = make(map[string]netip.Addr)
	tenantAt  = make(map[netip.Addr]string)
)

func Enable(p netip.Prefix) {
	mutex.Lock()
	defer mutex.Unlock()
	prefix = p
	next = prefix.Addr().Next()
	bits = prefix.Bits()
}

func Contract(tenant string, args ...string) (netip.Prefix, error) {
	mutex.Lock()
	defer mutex.Unlock()
	if !prefix.IsValid() {
		return z, egress.Marked(ErrInvalid)
	}
	if len(args) == 0 {
		if addr, ok := addressOf[tenant]; ok {
			return netip.PrefixFrom(addr, bits), nil
		} else {
			for pass := 0; pass < 2; {
				addr = next
				next = next.Next()
				if _, occupied := tenantAt[addr]; !occupied {
					addressOf[tenant] = addr
					tenantAt[addr] = tenant
					return netip.PrefixFrom(addr, bits), nil
				}
				if !prefix.Contains(next) {
					next = prefix.Addr().Next()
					pass += 1
				}
			}
			return z, egress.Marked(ErrUnavailable)
		}
	} else if addr, err := netip.ParseAddr(args[0]); err != nil {
		return z, egress.Marked(err)
	} else if !prefix.Contains(addr) {
		return z, egress.Marked(ErrInvalid)
	} else if occupant, occupied := tenantAt[addr]; !occupied {
		addressOf[tenant] = addr
		tenantAt[addr] = tenant
		return netip.PrefixFrom(addr, bits), nil
	} else if occupant == tenant {
		return netip.PrefixFrom(addr, bits), nil
	} else {
		return z, egress.Marked(ErrOccupied)
	}
}

func MarshalJSON() ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	return json.Marshal(addressOf)
}

func MarshalText() ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	buf := new(bytes.Buffer)
	for k, v := range addressOf {
		fmt.Fprint(buf, k, ": ", v, "\n")
	}
	return buf.Bytes(), nil
}
