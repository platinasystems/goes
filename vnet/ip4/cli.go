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

	// Sync adjacency stats with hardware.
	m.CallAdjSyncCounterHooks()

	type unreachable struct {
		valid       bool
		via         Address
		viaFibIndex ip.FibIndex
		weight      ip.NextHopWeight
	}
	type route struct {
		prefixFibIndex ip.FibIndex
		prefix         Prefix
		r              mapFibResult
		u              unreachable
	}
	rs := []route{}
	for fi := range m.fibs {
		fib := m.fibs[fi]
		if fib == nil {
			continue
		}
		t := ip.FibIndex(fi).Name(&m.Main)
		if cf.showTable != "" && t != cf.showTable {
			continue
		}
		fib.reachable.foreach(func(p *Prefix, r mapFibResult) {
			rt := route{prefixFibIndex: ip.FibIndex(fi), prefix: *p, r: r}
			rs = append(rs, rt)
		})
		fib.unreachable.foreach(func(p *Prefix, r mapFibResult) {
			rt := route{prefix: *p, r: r}
			u := unreachable{
				valid: true,
			}
			for nh, ps := range r.nh {
				u.via = nh.a
				u.viaFibIndex = nh.i
				for pi, nher := range ps {
					rt.prefix = pi.p
					rt.prefixFibIndex = pi.i
					u.weight = nher.NextHopWeight()
					rt.u = u
					rs = append(rs, rt)
				}
			}
		})
	}
	sort.Slice(rs, func(i, j int) bool {
		if cmp := int(rs[i].prefixFibIndex) - int(rs[j].prefixFibIndex); cmp != 0 {
			return cmp < 0
		}
		return rs[i].prefix.LessThan(&rs[j].prefix)
	})

	fmt.Fprintf(w, "%6s%30s%40s\n", "Table", "Destination", "Adjacency")
	for ri := range rs {
		r := &rs[ri]
		var lines []string
		if r.u.valid {
			nhs := fmt.Sprintf("%10sunreachable via %v", "", &r.u.via)
			if r.u.viaFibIndex != r.prefixFibIndex {
				nhs += ", table " + r.u.viaFibIndex.Name(&m.Main)
			}
			if r.u.weight != 1 {
				nhs += fmt.Sprintf(", weight %d", r.u.weight)
			}
			lines = []string{nhs}
		} else {
			lines = m.adjLines(r.r.adj, cf.detail)
		}
		for i := range lines {
			if i == 0 {
				fmt.Fprintf(w, "%12s%30s%s\n", r.prefixFibIndex.Name(&m.Main), &r.prefix, lines[i])
			} else {
				fmt.Fprintf(w, "%12s%30s%s\n", "", "", lines[i])
			}
		}
		if cf.detail {
			fmt.Fprintf(w, "%s", &r.r.nh)
		}
	}

	return
}

func (m *Main) adjLines(baseAdj ip.Adj, detail bool) (lines []string) {
	const initialSpace = "  "
	nhs := m.NextHopsForAdj(baseAdj)
	adjs := m.GetAdj(baseAdj)
	ai := ip.Adj(0)
	for ni := range nhs {
		nh := &nhs[ni]
		adj := baseAdj + ai
		line := fmt.Sprintf("%s%6d: ", initialSpace, adj)
		ss := []string{}
		adj_lines := adjs[ai].String(&m.Main)
		if nh.Weight != 1 || nh.Adj != baseAdj {
			adj_lines[0] += fmt.Sprintf(" %d-%d, %d x %d", adj, adj+ip.Adj(nh.Weight)-1, nh.Weight, nh.Adj)
		}
		// Indent subsequent lines like first line if more than 1 lines.
		for i := 1; i < len(adj_lines); i++ {
			adj_lines[i] = fmt.Sprintf("%*s%s", len(line), "", adj_lines[i])
		}
		ss = append(ss, adj_lines...)

		counterAdj := nh.Adj
		if !m.EqualAdj(adj, nh.Adj) {
			counterAdj = adj
		}

		m.Main.ForeachAdjCounter(counterAdj, func(tag string, v vnet.CombinedCounter) {
			if v.Packets != 0 || detail {
				ss = append(ss, fmt.Sprintf("%s%spackets %16d", initialSpace, tag, v.Packets))
				ss = append(ss, fmt.Sprintf("%s%sbytes   %16d", initialSpace, tag, v.Bytes))
			}
		})

		for _, s := range ss {
			lines = append(lines, line+s)
			line = initialSpace
		}

		ai += ip.Adj(nh.Weight)
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
