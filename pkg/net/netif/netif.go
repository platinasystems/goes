// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/netip"
	"os/signal"
	"sort"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/net/ift"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
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
)

type NetIf struct {
	net.Interface
	Type int
	// IP & IPv6
	Prefixes   []netip.Prefix
	Multicasts []netip.Addr
	// Extra attributes
	Extra  map[string]any
	Rx, Tx struct {
		Packets, Bytes, Drops, Errors uint64
	}
	Collisions uint64
}

func (nif *NetIf) Addrs() (addrs []net.Addr, err error) {
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

func (nif *NetIf) MulticastAddrs() (addrs []net.Addr, err error) {
	for _, addr := range nif.Multicasts {
		addrs = append(addrs, &net.IPAddr{
			IP:   net.IP(addr.AsSlice()),
			Zone: addr.Zone(),
		})
	}
	return
}

func (nif *NetIf) Format(w fmt.State, verb rune) {
	buf := new(bytes.Buffer)
	t, _ := fmt.Fprintf(w, "%s[%d]:", nif.Name, nif.Index)
	wrap := func() {
		if t+buf.Len() > 80 {
			fmt.Fprint(w, "\n\t")
			t = 8
		} else {
			fmt.Fprint(w, " ")
			t += 1
		}
		n, _ := w.Write(buf.Bytes())
		t += n
	}
	bprintf := func(format string, args ...any) {
		buf.Reset()
		fmt.Fprintf(buf, format, args...)
		wrap()
	}
	bprint := func(args ...any) {
		buf.Reset()
		fmt.Fprint(buf, args...)
		wrap()
	}
	bprintf("flags=%04x<%s>", uint(nif.Flags), nif.Flags)
	bprint("type ", ift.Name(nif.Type))
	bprintf("mtu %d", nif.MTU)
	if len(nif.HardwareAddr) == 6 && nif.HardwareAddr[0] != 0 {
		bprint("mac ", nif.HardwareAddr)
	}
	keys := make([]string, 0, len(nif.Extra))
	for k := range nif.Extra {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		bprint(k, " ", nif.Extra[k])
	}
	for _, prefix := range nif.Prefixes {
		if prefix.Addr().Is6() {
			bprint("inet6 ", prefix)
		} else {
			bprint("inet ", prefix)
		}
	}
	if t > 8 {
		fmt.Fprintln(w)
	}
}

func Indexed(i int) *NetIf {
	return cache().byIndex[i]
}

func Interfaces() []*NetIf {
	return cache().list
}

func Name(i int) string {
	if nif := Indexed(i); nif != nil {
		return nif.Name
	}
	return fmt.Sprintf("%d", i)
}

func Named(s string) *NetIf {
	return cache().byName[s]
}

func Range(f func(*NetIf) bool) {
	for _, nif := range Interfaces() {
		if !f(nif) {
			break
		}
	}
}

var cache = sync.OnceValue(func() (nifs struct {
	list    []*NetIf
	byIndex map[int]*NetIf
	byName  map[string]*NetIf
}) {

	var err error

	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, termination.Signals...)
	defer stop()

	if nifs.list, err = List(ctx); err != nil {
		panic(err)
	}
	nifs.byIndex = make(map[int]*NetIf)
	nifs.byName = make(map[string]*NetIf)
	for _, nif := range nifs.list {
		nifs.byIndex[nif.Index] = nif
		nifs.byName[nif.Name] = nif
	}
	return nifs
})
