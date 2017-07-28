// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip4

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"

	"fmt"
	"sort"
)

type fibShowUsageHook func(w cli.Writer)

//go:generate gentemplate -id FibShowUsageHook -d Package=ip4 -d DepsType=fibShowUsageHookVec -d Type=fibShowUsageHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl

type showFibConfig struct {
	detail      bool
	summary     bool
	unreachable bool
	showTable   string
}

func (m *Main) showIpFib(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	cf := showFibConfig{}
	for !in.End() {
		switch {
		case in.Parse("d%*etail"):
			cf.detail = true
		case in.Parse("s%*ummary"):
			cf.summary = true
		case in.Parse("un%*reachable"):
			cf.unreachable = true
		case in.Parse("t%*able %s", &cf.showTable):
		default:
			err = cli.ParseError
			return
		}
	}

	if cf.summary {
		m.showSummary(w)
		return
	}

	if cf.unreachable {
		m.showUnreachable(w, cf)
		return
	}

	m.showReachable(w, cf)
	return
}

func (m *Main) showReachable(w cli.Writer, cf showFibConfig) {
	// Sync adjacency stats with hardware.
	m.CallAdjSyncCounterHooks()

	type route struct {
		table  ip.FibIndex
		prefix Prefix
		r      mapFibResult
	}
	rs := []route{}
	for fi := range m.fibs {
		fib := m.fibs[fi]
		if fib != nil {
			t := ip.FibIndex(fi).Name(&m.Main)
			if cf.showTable != "" && t != cf.showTable {
				continue
			}
			fib.reachable.foreach(func(p *Prefix, r mapFibResult) {
				rs = append(rs, route{table: ip.FibIndex(fi), prefix: *p, r: r})
			})
		}
	}
	sort.Slice(rs, func(i, j int) bool {
		if cmp := int(rs[i].table) - int(rs[j].table); cmp != 0 {
			return cmp < 0
		}
		return rs[i].prefix.LessThan(&rs[j].prefix)
	})

	fmt.Fprintf(w, "%6s%30s%20s\n", "Table", "Destination", "Adjacency")
	for ri := range rs {
		r := &rs[ri]
		lines := m.adjLines(r.r.adj, cf.detail)
		for i := range lines {
			if i == 0 {
				fmt.Fprintf(w, "%12s%30s%s\n", r.table.Name(&m.Main), &r.prefix, lines[i])
			} else {
				fmt.Fprintf(w, "%12s%30s%s\n", "", "", lines[i])
			}
		}
		if cf.detail {
			fmt.Fprintf(w, "%s", &r.r.nh)
		}
	}
}

func (m *Main) showUnreachable(w cli.Writer, cf showFibConfig) {
	type unreachable struct {
		table ip.FibIndex
		p     Prefix
		nh    Address
		nhw   ip.NextHopWeight
	}
	us := []unreachable{}
	for fi := range m.fibs {
		fib := m.fibs[fi]
		if fib == nil {
			continue
		}
		t := ip.FibIndex(fi).Name(&m.Main)
		if cf.showTable != "" && t != cf.showTable {
			continue
		}
		fib.unreachable.foreach(func(p *Prefix, r mapFibResult) {
			u := unreachable{table: ip.FibIndex(fi)}
			for nh, ps := range r.nh {
				u.nh = nh
				for p, w := range ps {
					u.p = p
					u.nhw = w
					us = append(us, u)
				}
			}
		})
	}
	sort.Slice(us, func(i, j int) bool {
		if cmp := int(us[i].table) - int(us[j].table); cmp != 0 {
			return cmp < 0
		}
		return us[i].p.LessThan(&us[j].p)
	})

	fmt.Fprintf(w, "%6s%30s%20s\n", "Table", "Destination", "Next Hop")
	for ui := range us {
		u := &us[ui]
		nhs := u.nh.String()
		if u.nhw != 1 {
			nhs += fmt.Sprintf(", %d", u.nhw)
		}
		fmt.Fprintf(w, "%12s%30s%20s\n", u.table.Name(&m.Main), &u.p, nhs)
	}
}

func (m *Main) adjLines(adj ip.Adj, detail bool) (lines []string) {
	const initialSpace = "  "
	nhs := m.NextHopsForAdj(adj)
	adjs := m.GetAdj(adj)
	ai := 0
	for ni := range nhs {
		nh := &nhs[ni]
		line := fmt.Sprintf("%s%6d: ", initialSpace, int(adj)+ai)
		ss := []string{}
		adj_lines := adjs[ai].String(&m.Main)
		if nh.Weight != 1 || nh.Adj != adj {
			adj_lines[0] += fmt.Sprintf(" %d-%d, %d x %d", int(adj)+ai, int(adj)+ai+int(nh.Weight)-1, nh.Weight, nh.Adj)
		}
		// Indent subsequent lines like first line if more than 1 lines.
		for i := 1; i < len(adj_lines); i++ {
			adj_lines[i] = fmt.Sprintf("%*s%s", len(line), "", adj_lines[i])
		}
		ss = append(ss, adj_lines...)

		m.Main.ForeachAdjCounter(nh.Adj, ip.Adj(0), func(tag string, v vnet.CombinedCounter) {
			if v.Packets != 0 || detail {
				ss = append(ss, fmt.Sprintf("%s%spackets %16d", initialSpace, tag, v.Packets))
				ss = append(ss, fmt.Sprintf("%s%sbytes   %16d", initialSpace, tag, v.Bytes))
			}
		})

		for _, s := range ss {
			lines = append(lines, line+s)
			line = initialSpace
		}

		ai += int(nh.Weight)
	}

	return
}

func (m *Main) showSummary(w cli.Writer) {
	fmt.Fprintf(w, "%6s%12s\n", "Table", "Routes")
	for fi := range m.fibs {
		fib := m.fibs[fi]
		if fib != nil {
			fmt.Fprintf(w, "%12s%12d\n", ip.FibIndex(fi).Name(&m.Main), fib.Len())
		}
	}
	u := m.GetAdjacencyUsage()
	fmt.Fprintf(w, "Adjacencies: heap %d used, %d free\n", u.Used, u.Free)
	for i := range m.FibShowUsageHooks.hooks {
		m.FibShowUsageHooks.Get(i)(w)
	}
}

func (m *Main) cliInit(v *vnet.Vnet) {
	cmds := [...]cli.Command{
		cli.Command{
			Name:      "show ip fib",
			ShortHelp: "show ip4 forwarding table",
			Action:    m.showIpFib,
		},
	}
	for i := range cmds {
		v.CliAdd(&cmds[i])
	}
}
