// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/pg"

	"fmt"
)

type pgStream struct {
	pg.Stream
	h []vnet.PacketHeader
}

func (s *pgStream) PacketHeaders() []vnet.PacketHeader { return s.h }

type pgMain struct {
	v       *vnet.Vnet
	typeMap map[Type]pg.StreamType
}

func (m *pgMain) Name() string { return "ethernet" }

func (m *pgMain) initTypes() {
	if m.typeMap != nil {
		return
	}
	m.typeMap = make(map[Type]pg.StreamType)
	m.typeMap[IP4.FromHost()] = pg.GetStreamType(m.v, "ip4")
}

func (m *pgMain) ParseStream(in *parse.Input) (r pg.Streamer, err error) {
	m.initTypes()

	var s pgStream
	for !in.End() {
		var h struct {
			v []VlanHeader
			h Header
		}
		switch {
		case in.Parse("%v", &h.h):
			for {
				var v VlanHeader
				if in.Parse("%v", &v) {
					h.v = append(h.v, v)
				} else {
					break
				}
			}

			inner_type := h.h.Type
			switch len(h.v) {
			case 0:
			case 1:
				h.v[0].Type = inner_type
				h.h.Type = VLAN.FromHost()
			case 2:
				h.v[1].Type = inner_type
				h.v[0].Type = VLAN.FromHost()
				h.h.Type = VLAN_IN_VLAN.FromHost()
			case 3:
				err = fmt.Errorf("number of vlans must be <= 2, given %d", len(h.v))
				return
			}

			s.h = append(s.h, &h.h)
			for i := range h.v {
				s.h = append(s.h, &h.v[i])
			}
			if t, ok := m.typeMap[inner_type]; ok {
				var sub_r pg.Streamer
				sub_r, err = t.ParseStream(in)
				if err != nil {
					err = fmt.Errorf("ethernet %s: %s `%s'", t.Name(), err, in)
					return
				}
				s.h = append(s.h, sub_r.PacketHeaders()...)
			}
		default:
			err = parse.ErrInput
			return
		}
	}
	if err == nil {
		r = &s
	}
	return
}

func (m *pgMain) pgInit(v *vnet.Vnet) {
	m.v = v
	pg.AddStreamType(v, "ethernet", m)
}
