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
}

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
		var min, max uint64
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

		incLoop:
			for {
				switch {
				case in.Parse("src %d-%d", &min, &max):
					ai := &addressIncrement{
						s:     &s,
						base:  h.h.Src,
						isSrc: true,
						min:   min,
						max:   max,
					}
					s.DataHooks.Add(ai.Do)

				case in.Parse("dst %d-%d", &min, &max):
					ai := &addressIncrement{
						s:     &s,
						base:  h.h.Dst,
						isSrc: false,
						min:   min,
						max:   max,
					}
					s.DataHooks.Add(ai.Do)

				default:
					break incLoop
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

			s.AddHeader(&h.h)
			for i := range h.v {
				s.AddHeader(&h.v[i])
			}
			if t, ok := m.typeMap[inner_type]; ok {
				var sub_r pg.Streamer
				sub_r, err = t.ParseStream(in)
				if err != nil {
					err = fmt.Errorf("ethernet %s: %s `%s'", t.Name(), err, in)
					return
				}
				s.AddStreamer(sub_r)
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

type addressIncrement struct {
	s     *pgStream
	base  Address
	cur   uint64
	min   uint64
	max   uint64
	isSrc bool
}

func (ai *addressIncrement) Do(dst []vnet.Ref, dataOffset uint) {
	for i := range dst {
		h := GetHeader(&dst[i])
		if ai.isSrc {
			h.Src = ai.base
			h.Src.Add(ai.cur)
		} else {
			h.Dst = ai.base
			h.Dst.Add(ai.cur)
		}
		ai.cur++
		if ai.cur > ai.max {
			ai.cur = ai.min
		}
	}
}

func (m *pgMain) pgInit(v *vnet.Vnet) {
	m.v = v
	pg.AddStreamType(v, "ethernet", m)
}
