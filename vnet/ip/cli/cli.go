// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"

	"fmt"
)

var packageIndex uint

type Main struct {
	vnet.Package
}

func Init(v *vnet.Vnet) {
	m := &Main{}
	packageIndex = v.AddPackage("ip-cli", m)
	m.DependsOn("ip4", "ip6")
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

func (m *Main) ip_route(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	type add_del struct {
		is_del     bool
		is_ip6     bool
		ip4_prefix ip4.Prefix
		count      uint
		ip4_nhs    []ip4.NextHop
		adjs       []ip.Adjacency
		fib_index  ip.FibIndex
	}
	var x add_del

	switch {
	case in.Parse("add"):
		x.is_del = false
	case in.Parse("del"):
		x.is_del = true
	}

	if !in.Parse("%v", &x.ip4_prefix) {
		err = fmt.Errorf("looking for prefix, got `%s'", in)
		return
	}

	m4 := ip4.GetMain(m.Vnet)

	x.count = 1
loop:
	for !in.End() {
		switch {
		case in.Parse("c%*ount %d", &x.count):
		case in.Parse("t%*able %d", &x.fib_index):
		default:
			break loop
		}
	}

	var (
		adj ip.Adjacency
		nh4 ip4.NextHop
	)
	switch {
	case in.Parse("via %v", &nh4, m.Vnet):
		x.ip4_nhs = append(x.ip4_nhs, nh4)
	case in.Parse("%v", &adj, &m4.Main):
		x.adjs = append(x.adjs, adj)
	default:
		err = fmt.Errorf("looking for via NEXT-HOP or adjacency, got `%s'", in)
		return
	}

	for i := uint(0); i < x.count; i++ {
		p := x.ip4_prefix.Add(i)

		for i := range x.ip4_nhs {
			err = m4.AddDelRouteNextHop(&p, &x.ip4_nhs[i], x.is_del)
			if err != nil {
				return
			}
		}

		if len(x.adjs) > 0 {
			pi := p.ToIpPrefix()
			for i := range x.adjs {
				var (
					ai ip.Adj
					as []ip.Adjacency
					ok bool
				)
				if x.is_del {
					ai, ok = m4.GetRouteFibIndex(&pi, x.fib_index)
					if !ok {
						err = fmt.Errorf("%s not found", &pi)
						return
					}
				} else {
					ai, as = m4.NewAdj(1)
					as[0] = x.adjs[i]
					m4.CallAdjAddHooks(ai)
				}
				if _, err = m4.AddDelRoute(&pi, x.fib_index, ai, x.is_del); err != nil {
					return
				}
				if x.is_del {
					m4.CallAdjDelHooks(ai)
					m4.DelAdj(ai)
				}
			}
		}
	}

	return
}

func (m *Main) ip_interface(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var (
		si vnet.Si
		i  int
	)
	v := m.Vnet
	m4 := ip4.GetMain(m.Vnet)
	var p ip4.Prefix
	switch {
	case in.Parse("fib %v %d", &si, v, &i):
		m4.SetFibIndexForSi(si, ip.FibIndex(i))
	case in.Parse("a%*ddress %v %v", &si, v, &p):
		m4.AddDelInterfaceAddress(si, &p, false)
	case in.Parse("d%*elete %v %v", &si, v, &p):
		m4.AddDelInterfaceAddress(si, &p, true)
	default:
		err = cli.ParseError
	}
	return
}

func (m *Main) Init() (err error) {
	v := m.Vnet

	cmds := []cli.Command{
		cli.Command{
			Name:      "ip route",
			ShortHelp: "add/delete ip4/ip6 routes",
			Action:    m.ip_route,
		},
		cli.Command{
			Name:      "ip interface",
			ShortHelp: "ip interface commands",
			Action:    m.ip_interface,
		},
	}
	for i := range cmds {
		v.CliAdd(&cmds[i])
	}
	return
}
