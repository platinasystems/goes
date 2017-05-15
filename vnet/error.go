// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/loop"

	"fmt"
	"sort"
	"sync/atomic"
)

type ErrorRef uint32

type BufferError struct {
	nodeIndex  uint32
	errorIndex uint16
}

type errorThread struct {
	counts          elib.Uint64Vec
	countsLastClear elib.Uint64Vec
	cache           ErrorRef
}

//go:generate gentemplate -d Package=vnet -id errorThread -d VecType=errorThreadVec -d Type=*errorThread github.com/platinasystems/go/elib/vec.tmpl
//go:generate gentemplate -d Package=vnet -id error -d VecType=errVec -d Type=err github.com/platinasystems/go/elib/vec.tmpl

type errorNode struct {
	OutputNode
	threads errorThreadVec
	errs    errVec
}

func (n *errorNode) getThread(id uint) (t *errorThread) {
	n.threads.Validate(uint(id))
	if t = n.threads[id]; t == nil {
		t = &errorThread{}
		n.threads[id] = t
	}
	i := n.errs.Len()
	if i > 0 {
		t.counts.Validate(i - 1)
		t.countsLastClear.Validate(i - 1)
	}
	return
}

const poisonErrorRef = 0xfeedface

func (t *errorThread) count(e ErrorRef, n uint64) {
	if elib.Debug {
		if e == poisonErrorRef {
			panic("error ref not set")
		}
	}
	t.counts[e] += n
}

func (n *errorNode) MakeLoopIn() loop.LooperIn { return &RefIn{} }

var ErrorNode = &errorNode{}

func init() {
	AddInit(func(v *Vnet) {
		v.RegisterOutputNode(ErrorNode, "error")
	})
}

func (en *errorNode) NodeOutput(ri *RefIn) {
	ts := en.getThread(ri.ThreadId())

	cache := ts.cache
	cacheCount := uint64(0)
	i, n := uint(0), ri.InLen()
	for i+4 <= n {
		e0, e1, e2, e3 := ErrorRef(ri.Refs[i+0].Aux), ErrorRef(ri.Refs[i+1].Aux), ErrorRef(ri.Refs[i+2].Aux), ErrorRef(ri.Refs[i+3].Aux)
		cacheCount += 4
		i += 4
		if e0 == cache && e1 == cache && e2 == cache && e3 == cache {
			continue
		}
		cacheCount -= 4
		ts.count(e0, 1)
		ts.count(e1, 1)
		ts.count(e2, 1)
		ts.count(e3, 1)
		if e0 == e1 && e2 == e3 && e0 == e2 {
			ts.counts[cache] += cacheCount
			cache, cacheCount = e0, 0
		}
	}

	for i < n {
		ts.count(ErrorRef(ri.Refs[i+0].Aux), 1)
		i++
	}

	ts.count(cache, cacheCount)
	ts.cache = cache
	ri.FreeRefs(n)
}

type err struct {
	nodeIndex uint32
	str       string
}

func (n *Node) NewError(s string) (r ErrorRef) {
	e := err{nodeIndex: uint32(n.Index()), str: s}
	en := ErrorNode
	r = ErrorRef(len(en.errs))
	en.errs = append(en.errs, e)
	return
}

func (r *RefOpaque) SetError(n *Node, i uint) { r.Aux = uint32(n.errorRefs[i]) }
func (n *Node) SetError(r *Ref, i uint)       { r.SetError(n, i) }
func (n *Node) CountError(i, count uint) {
	ts := ErrorNode.getThread(0)
	atomic.AddUint64(&ts.counts[n.errorRefs[i]], uint64(count))
}

type errNode struct {
	Node  string `format:"%-30s"`
	Error string
	Count uint64 `format:"%16d"`
}
type errNodes []errNode

func (ns errNodes) Less(i, j int) bool {
	if ns[i].Node == ns[j].Node {
		return ns[i].Error < ns[j].Error
	}
	return ns[i].Node < ns[j].Node
}
func (ns errNodes) Swap(i, j int) { ns[i], ns[j] = ns[j], ns[i] }
func (ns errNodes) Len() int      { return len(ns) }

func (v *Vnet) showErrors(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	en := ErrorNode
	ns := []errNode{}
	for i := range en.errs {
		e := &en.errs[i]
		c := uint64(0)
		for _, t := range en.threads {
			if t != nil {
				c += t.counts[i]
				if i < len(t.countsLastClear) {
					c -= t.countsLastClear[i]
				}
			}
		}
		if c > 0 {
			n := v.loop.DataNodes[e.nodeIndex].GetNode()
			ns = append(ns, errNode{
				Node:  n.Name(),
				Error: e.str,
				Count: c,
			})
		}
	}
	if len(ns) > 1 {
		sort.Sort(errNodes(ns))
	}
	if len(ns) > 0 {
		elib.TabulateWrite(w, ns)
	} else {
		fmt.Fprintln(w, "No errors since last clear.")
	}
	return
}

func (v *Vnet) clearErrors(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	for _, t := range ErrorNode.threads {
		if t != nil {
			copy(t.countsLastClear, t.counts)
		}
	}
	return
}

func init() {
	AddInit(func(v *Vnet) {
		v.CliAdd(&cli.Command{
			Name:      "show errors",
			ShortHelp: "show error counters",
			Action:    v.showErrors,
		})
		v.CliAdd(&cli.Command{
			Name:      "clear errors",
			ShortHelp: "clear error counters",
			Action:    v.clearErrors,
		})
	})
}
