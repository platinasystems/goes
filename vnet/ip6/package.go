// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip6

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"

	"fmt"
)

var packageIndex uint

func Init(v *vnet.Vnet) *ip.Main {
	m := &Main{}
	packageIndex = v.AddPackage("ip6", m)
	cf := ip.FamilyConfig{
		Family:          ip.Ip6,
		AddressStringer: ipAddressStringer,
		RewriteNode:     &m.rewriteNode,
		PacketType:      vnet.IP6,
	}
	m.Main.PackageInit(v, cf)
	return &m.Main
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

func ipAddressStringer(a *ip.Address) string { return IpAddress(a).String() }

type Main struct {
	vnet.Package
	ip.Main
	nodeMain
}

func RegisterLayer(v *vnet.Vnet, t ip.Protocol, l vnet.Layer) {
	m := GetMain(v)
	m.RegisterLayer(v, t, l)
}
func UnregisterLayer(v *vnet.Vnet, t ip.Protocol) (ok bool) {
	m := GetMain(v)
	ok = m.UnregisterLayer(v, t)
	return
}

func (m *Main) FormatLayer(b []byte) (lines []string) {
	h := (*Header)(vnet.Pointer(b))
	lines = append(lines, h.String())
	n := SizeofHeader
	if n < len(b) {
		if l, ok := m.GetLayer(h.Protocol); ok {
			lines = append(lines, l.FormatLayer(b[n:])...)
		} else {
			panic(fmt.Errorf("no formatter for protocol %v", h.Protocol))
		}
	}
	return
}

func (m *Main) ParseLayer(b []byte, in *parse.Input) (n uint) {
	h := (*Header)(vnet.Pointer(b))
	h.Parse(in)
	n = SizeofHeader
	if !in.End() {
		if l, ok := m.GetLayer(h.Protocol); ok {
			n += l.ParseLayer(b[n:], in)
		} else {
			panic(fmt.Errorf("no parser for protocol %v", h.Protocol))
		}
	}
	return
}

func (m *Main) Init() (err error) {
	v := m.Vnet
	m.Main.Init(v)
	m.nodeInit(v)
	RegisterLayer(v, ip.IP6_IN_IP, m)
	ethernet.RegisterLayer(v, ethernet.TYPE_IP6, m)
	return
}
