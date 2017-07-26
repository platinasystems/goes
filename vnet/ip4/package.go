// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip4

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
	packageIndex = v.AddPackage("ip4", m)
	cf := ip.FamilyConfig{
		Family:           ip.Ip4,
		AddressStringer:  ipAddressStringer,
		RewriteNode:      &m.rewriteNode,
		PacketType:       vnet.IP4,
		GetRoute:         m.getRoute,
		GetRouteFibIndex: m.getRouteFibIndex,
		AddDelRoute:      m.addDelRoute,
		RemapAdjacency:   m.remapAdjacency,
	}
	m.Main.PackageInit(v, cf)
	v.RegisterSwIfAdminUpDownHook(m.swIfAdminUpDown)
	return &m.Main
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

func ipAddressStringer(a *ip.Address) string { return IpAddress(a).String() }

type Main struct {
	vnet.Package
	ip.Main
	fibMain
	nodeMain
	pgMain
	ifAddrAddDelHooks IfAddrAddDelHookVec
	FibShowUsageHooks fibShowUsageHookVec
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
			panic(fmt.Errorf("no formatter for protocol %s", h.Protocol))
		}
	}
	return
}

func (m *Main) ParseLayer(b []byte, in *parse.Input) (n uint) {
	h := (*Header)(vnet.Pointer(b))
	h.Parse(in)
	h.Checksum = h.ComputeChecksum()
	n = SizeofHeader
	if !in.End() {
		if l, ok := m.GetLayer(h.Protocol); ok {
			n += l.ParseLayer(b[n:], in)
		} else {
			panic(fmt.Errorf("no parser for protocol %s", h.Protocol))
		}
	}
	return
}

func (m *Main) Init() (err error) {
	v := m.Vnet
	m.Main.Init(v)
	m.nodeInit(v)
	m.pgInit(v)
	m.cliInit(v)
	RegisterLayer(v, ip.IP_IN_IP, m)
	ethernet.RegisterLayer(v, ethernet.TYPE_IP4, m)
	return
}
