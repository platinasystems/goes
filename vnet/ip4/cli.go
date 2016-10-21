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

type showIpFibRoute struct {
	table  ip.FibIndex
	prefix Prefix
	adj    ip.Adj
}

type showIpFibRoutes []showIpFibRoute

func (x showIpFibRoutes) Less(i, j int) bool {
	if cmp := int(x[i].table) - int(x[j].table); cmp != 0 {
		return cmp < 0
	}
	return x[i].prefix.LessThan(&x[j].prefix)
}

func (x showIpFibRoutes) Swap(i, j int) { x[i], x[j] = x[j], x[i] }
func (x showIpFibRoutes) Len() int      { return len(x) }

func (m *Main) showIpFib(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {

	detail := false
	for !in.End() {
		switch {
		case in.Parse("d%*etail"):
			detail = true
		default:
			err = cli.ParseError
			return
		}
	}

	// Sync adjacency stats with hardware.
	m.CallAdjSyncCounterHooks()

	rs := []showIpFibRoute{}
	for fi := range m.fibs {
		fib := m.fibs[fi]
		fib.foreach(func(p *Prefix, a ip.Adj) {
			rs = append(rs, showIpFibRoute{table: ip.FibIndex(fi), prefix: *p, adj: a})
		})
	}
	sort.Sort(showIpFibRoutes(rs))

	fmt.Fprintf(w, "%6s%30s%20s\n", "Table", "Destination", "Adjacency")
	for ri := range rs {
		r := &rs[ri]
		lines := []string{}
		adjs := m.GetAdj(r.adj)
		for ai := range adjs {
			initialSpace := "  "
			line := initialSpace
			if len(adjs) > 1 {
				line += fmt.Sprintf("%d: ", ai)
			}
			ss := adjs[ai].String(&m.Main)

			m.Main.ForeachAdjCounter(r.adj+ip.Adj(ai), func(tag string, v vnet.CombinedCounter) {
				if v.Packets != 0 || detail {
					ss = append(ss, fmt.Sprintf("%s%spackets %16d", initialSpace, tag, v.Packets))
					ss = append(ss, fmt.Sprintf("%s%sbytes   %16d", initialSpace, tag, v.Bytes))
				}
			})

			for _, s := range ss {
				lines = append(lines, line+s)
				line = initialSpace
			}
		}
		for i := range lines {
			if i == 0 {
				fmt.Fprintf(w, "%6d%30s%s\n", r.table, &r.prefix, lines[i])
			} else {
				fmt.Fprintf(w, "%6s%30s%s\n", "", "", lines[i])
			}
		}
	}

	return
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
