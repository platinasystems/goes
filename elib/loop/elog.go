// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package loop

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
)

type poller_elog_event_type byte

const (
	poller_start poller_elog_event_type = iota + 1
	poller_done
	poller_suspend
	poller_resume
	poller_resumed
	poller_activate
)

var poller_elog_event_type_names = [...]string{
	poller_start:    "start",
	poller_done:     "done",
	poller_suspend:  "suspend",
	poller_resume:   "resume",
	poller_resumed:  "resumed",
	poller_activate: "activate",
}

func (t poller_elog_event_type) String() string {
	return elib.Stringer(poller_elog_event_type_names[:], int(t))
}

type pollerElogEvent struct {
	poller_index byte
	event_type   poller_elog_event_type
	flags        byte
	node_name    elog.StringRef
}

func (n *Node) pollerElog(t poller_elog_event_type, f node_flags) {
	if elog.Enabled() {
		c := elog.GetCaller(elog.PointerToFirstArg(&n))
		le := pollerElogEvent{
			event_type:   t,
			poller_index: byte(n.activePollerIndex),
			flags:        byte(f),
			node_name:    n.elogNodeName,
		}
		elog.Addc(&le, c)
	}
}

func (e *pollerElogEvent) Format(x *elog.Context, f elog.Format) string {
	return f("loop%d %v %s %s", e.poller_index, e.event_type, x.GetString(e.node_name), node_flags(e.flags))
}
func (e *pollerElogEvent) SetData(x *elog.Context, p elog.Pointer) { *(*pollerElogEvent)(p) = *e }

type callEvent struct {
	active_index uint32
	n_vectors    uint32
	node_name    elog.StringRef
}

func (e *callEvent) Format(x *elog.Context, f elog.Format) string {
	return f("loop%d %s(%d)", e.active_index, x.GetString(e.node_name), e.n_vectors)
}
func (e *callEvent) SetData(x *elog.Context, p elog.Pointer) { *(*callEvent)(p) = *e }

func (e *callEvent) Strings(x *elog.Context) []string {
	return []string{}
}
func (e *callEvent) Encode(_ *elog.Context, b []byte) (i int) {
	i += elog.EncodeUint32(b[i:], e.active_index)
	i += elog.EncodeUint32(b[i:], uint32(e.node_name))
	i += elog.EncodeUint32(b[i:], e.n_vectors)
	return
}
func (e *callEvent) Decode(_ *elog.Context, b []byte) (i int) {
	e.active_index, i = elog.DecodeUint32(b, i)
	var x uint32
	x, i = elog.DecodeUint32(b, i)
	e.node_name = elog.StringRef(x)
	e.n_vectors, i = elog.DecodeUint32(b, i)
	return
}
