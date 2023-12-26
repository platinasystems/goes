// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/netip"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

var lease struct {
	sync.RWMutex
	prefix, z netip.Prefix
	next      netip.Addr
	bits      int
}

var (
	leaseAddressOf = make(map[string]netip.Addr)
	leaseTenantAt  = make(map[netip.Addr]string)
)

func leaseEnable(p netip.Prefix) {
	lease.Lock()
	defer lease.Unlock()
	lease.prefix = p
	lease.next = p.Addr().Next()
	lease.bits = p.Bits()
}

func leaseContract(tenant string, args []string) (netip.Prefix, error) {
	lease.Lock()
	defer lease.Unlock()
	if !lease.prefix.IsValid() {
		return lease.z, egress.Mark(ErrInvalid)
	}
	if len(args) == 0 {
		if addr, ok := leaseAddressOf[tenant]; ok {
			return netip.PrefixFrom(addr, lease.bits), nil
		} else {
			for pass := 0; pass < 2; {
				addr = lease.next
				lease.next = lease.next.Next()
				if _, occupied := leaseTenantAt[addr]; !occupied {
					leaseAddressOf[tenant] = addr
					leaseTenantAt[addr] = tenant
					return netip.PrefixFrom(addr, lease.bits), nil
				}
				if !lease.prefix.Contains(lease.next) {
					lease.next = lease.prefix.Addr().Next()
					pass += 1
				}
			}
			return lease.z, egress.Mark(ErrUnavailable)
		}
	} else if addr, err := netip.ParseAddr(args[0]); err != nil {
		return lease.z, egress.Mark(err)
	} else if !lease.prefix.Contains(addr) {
		return lease.z, egress.Mark(ErrInvalid)
	} else if occupant, occupied := leaseTenantAt[addr]; !occupied {
		leaseAddressOf[tenant] = addr
		leaseTenantAt[addr] = tenant
		return netip.PrefixFrom(addr, lease.bits), nil
	} else if occupant == tenant {
		return netip.PrefixFrom(addr, lease.bits), nil
	} else {
		return lease.z, egress.Mark(ErrOccupied)
	}
}

func leaseMarshalJSON() ([]byte, error) {
	lease.RLock()
	defer lease.RUnlock()
	return json.Marshal(leaseAddressOf)
}

func leaseMarshalText() ([]byte, error) {
	lease.RLock()
	defer lease.RUnlock()
	buf := new(bytes.Buffer)
	for k, v := range leaseAddressOf {
		fmt.Fprint(buf, k, ": ", v, "\n")
	}
	return buf.Bytes(), nil
}
