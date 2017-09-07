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
	ai_src, ai_dst addressIncrement
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
	m.typeMap[TYPE_IP4.FromHost()] = pg.GetStreamType(m.v, "ip4")
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
					s.ai_src = addressIncrement{
						base: h.h.Src,
						min:  min,
						max:  max,
					}

				case in.Parse("dst %d-%d", &min, &max):
					s.ai_dst = addressIncrement{
						base: h.h.Dst,
						min:  min,
						max:  max,
					}

				default:
					break incLoop
				}
			}

			inner_type := h.h.Type
			if len(h.v) > 0 {
				vt := TYPE_VLAN.FromHost()
				h.h.Type = vt
				for i := range h.v {
					t := inner_type
					if i+1 < len(h.v) {
						t = vt
					}
					h.v[i].Type = t
				}
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
			} else {
				var x parse.HexString
				if !in.Parse("%v", &x) {
					in.ParseError()
				}
				y := vnet.GivenPayload{Payload: []byte(x)}
				s.AddHeader(&y)
			}

		default:
			in.ParseError()
		}
	}
	if err == nil {
		r = &s
	}
	return
}

type addressIncrement struct {
	base Address
	cur  uint64
	min  uint64
	max  uint64
}

func (ai *addressIncrement) valid() bool { return ai.max > ai.min }

func (ai *addressIncrement) do(dst []vnet.Ref, dataOffset uint, isSrc bool) {
	for i := range dst {
		h := (*Header)(dst[i].DataOffset(dataOffset))
		if isSrc {
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
func (s *pgStream) Finalize(r []vnet.Ref, do uint) (changed bool) {
	if s.ai_src.valid() {
		s.ai_src.do(r, do, true)
		changed = true
	}
	if s.ai_dst.valid() {
		s.ai_dst.do(r, do, false)
		changed = true
	}
	return
}

func (m *pgMain) pgInit(v *vnet.Vnet) {
	m.v = v
	pg.AddStreamType(v, "ethernet", m)
}
