// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

type Parameter uint8

const (
	UnknownParameter Parameter = iota
	TrueFlagParameter
	FalseFlagParameter
	TrueAttrParameter
	FalseAttrParameter
	Int16Parameter
	Uint16Parameter
	Int32Parameter
	Uint32Parameter
	Int64Parameter
	Uint64Parameter
	BytesParameter
	HardwareAddrParameter
	StringParameter
	LinkModeParameter
	StateParameter
	PrefixParameter
	AddrParameter
)

type Netif struct {
	net.Interface
	Type fmt.Stringer
	// IP & IPv6
	Prefixes []netip.Prefix
	// Extra attributes
	Extra  map[string]any
	Rx, Tx struct {
		Packets, Bytes, Drops, Errors uint64
	}
	Collisions uint64
}

func (nif *Netif) Addrs() (addrs []net.Addr, err error) {
	for _, prefix := range nif.Prefixes {
		if addr := prefix.Addr(); !addr.IsMulticast() {
			addrs = append(addrs, &net.IPAddr{
				IP:   net.IP(addr.AsSlice()),
				Zone: addr.Zone(),
			})
		}
	}
	return
}

func (nif *Netif) MulticastAddrs() (addrs []net.Addr, err error) {
	for _, prefix := range nif.Prefixes {
		if addr := prefix.Addr(); addr.IsMulticast() {
			addrs = append(addrs, &net.IPAddr{
				IP:   net.IP(addr.AsSlice()),
				Zone: addr.Zone(),
			})
		}
	}
	return
}

func (nif *Netif) Format(w fmt.State, verb rune) {
	var n, t int
	sb := new(strings.Builder)
	n, _ = fmt.Fprintf(w, "%s[%d]:", nif.Name, nif.Index)
	t += n
	n, _ = fmt.Fprintf(w, " flags=%04x<%s>", uint(nif.Flags), nif.Flags)
	t += n
	n, _ = fmt.Fprintf(w, " mtu %d", nif.MTU)
	t += n
	for k, v := range nif.Extra {
		sb.Reset()
		fmt.Fprint(sb, k, " ", v)
		if t+sb.Len() > 80 {
			fmt.Fprint(w, "\n\t")
			t = 8
		} else {
			fmt.Fprint(w, " ")
			t += 1
		}
		n, _ = fmt.Fprint(w, sb)
		t += n
	}
	fmt.Fprint(w, "\n\t", nif.Type)
	if len(nif.HardwareAddr) == 6 && nif.HardwareAddr[0] != 0 {
		fmt.Fprint(w, " ", nif.HardwareAddr)
	}
	for _, prefix := range nif.Prefixes {
		fmt.Fprint(w, "\n\tinet")
		if prefix.Addr().Is6() {
			fmt.Fprint(w, "6")
		}
		fmt.Fprint(w, " ", prefix)
	}
	fmt.Fprintln(w)
}

func (nif *Netif) parseIFF(iff IFF) {
	if iff.Has(IFF_UP) {
		nif.Flags |= net.FlagUp
	} else {
		nif.Flags &^= net.FlagUp
	}
	if iff.Has(IFF_BROADCAST) {
		nif.Flags |= net.FlagBroadcast
	} else {
		nif.Flags &^= net.FlagBroadcast
	}
	if iff.Has(IFF_LOOPBACK) {
		nif.Flags |= net.FlagLoopback
	} else {
		nif.Flags &^= net.FlagLoopback
	}
	if iff.Has(IFF_POINTOPOINT) {
		nif.Flags |= net.FlagPointToPoint
	} else {
		nif.Flags &^= net.FlagPointToPoint
	}
	if iff.Has(IFF_MULTICAST) {
		nif.Flags |= net.FlagMulticast
	} else {
		nif.Flags &^= net.FlagMulticast
	}
	if iff.Has(IFF_RUNNING) {
		nif.Flags |= net.FlagRunning
	} else {
		nif.Flags &^= net.FlagRunning
	}
}
