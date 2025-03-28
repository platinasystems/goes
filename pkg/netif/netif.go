// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/netip"
	"sort"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
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

type NetIfs []*NetIf

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
		if t+1+buf.Len() > 80 {
			fmt.Fprint(w, "\n\t")
			t = 8
		} else if t > 8 {
			fmt.Fprint(w, " ")
			t += 1
		}
		n, _ := w.Write(buf.Bytes())
		buf.Reset()
		t += n
	}
	bprintf := func(format string, args ...any) {
		fmt.Fprintf(buf, format, args...)
		wrap()
	}
	bprint := func(args ...any) {
		fmt.Fprint(buf, args...)
		wrap()
	}
	flush := func() {
		if t > 8 {
			fmt.Fprint(w, "\n\t")
			t = 8
		}
	}
	bprintf("flags=%04x<%s>", uint(nif.Flags), nif.Flags)
	bprint("type ", xnet.IFTName(nif.Type))
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
		flush()
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

func Name(ctx context.Context, i int) (s string, err error) {
	nif, err := Indexed(ctx, i)
	if err == nil {
		s = nif.Name
	}
	return
}

// Call function `f` with each interface but stops if `f` returns falses
func Range(ctx context.Context, f func(context.Context, *NetIf) bool) {
	nifs, err := List(ctx)
	if err == nil {
		for _, nif := range nifs {
			if !f(ctx, nif) {
				break
			}
		}
	}
}

func (nifs NetIfs) Indexed(i int) (*NetIf, error) {
	for _, nif := range nifs {
		if nif.Index == i {
			return nif, nil
		}
	}
	return nil, xerrors.NotFound(i)
}

func (nifs NetIfs) Named(s string) (*NetIf, error) {
	for _, nif := range nifs {
		if nif.Name == s {
			return nif, nil
		}
	}
	return nil, xerrors.NotFound(s)
}

// FIXME implement ioctl/sysctl and netlink versions of these nif fetches.

// Returns indexed interface which may be nil.
func Indexed(ctx context.Context, i int) (*NetIf, error) {
	nifs, err := List(ctx)
	if err != nil {
		return nil, err
	}
	return nifs.Indexed(i)
}

// Returns named interface which may be nil.
func Named(ctx context.Context, s string) (*NetIf, error) {
	nifs, err := List(ctx)
	if err != nil {
		return nil, err
	}
	return nifs.Named(s)
}
