// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip4

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/pg"

	"fmt"
	"math/rand"
)

type pgStream struct {
	pg.Stream
}

type pgMain struct {
	v           *vnet.Vnet
	protocolMap map[ip.Protocol]pg.StreamType
}

func (m *pgMain) initProtocolMap() {
	if m.protocolMap != nil {
		return
	}
	m.protocolMap = make(map[ip.Protocol]pg.StreamType)
	// FIXME: not yet
	// m.protocol[ICMP] = pg.GetStreamType(m.v, "icmp")
}

func (m *pgMain) Name() string { return "ip4" }

var defaultHeader = Header{
	Protocol: ip.UDP,
	Src:      Address{0x1, 0x2, 0x3, 0x4},
	Dst:      Address{0x5, 0x6, 0x7, 0x8},
	Tos:      0,
	Ttl:      255,
	Ip_version_and_header_length: 0x45,
	Fragment_id:                  vnet.Uint16(0x1234).FromHost(),
	Flags_and_fragment_offset:    DontFragment.FromHost(),
}

func (m *pgMain) ParseStream(in *parse.Input) (r pg.Streamer, err error) {
	m.initProtocolMap()
	var s pgStream
	h := defaultHeader
	for !in.End() {
		var min, max uint64
		switch {
		case in.Parse("%v", &h):
		incLoop:
			for {
				switch {
				case in.Parse("src %v-%v", &min, &max):
					s.addInc(true, false, &h, min, max)
				case in.Parse("src %v", &max):
					s.addInc(true, false, &h, 0, max-1)
				case in.Parse("dst %v-%v", &min, &max):
					s.addInc(false, false, &h, min, max)
				case in.Parse("dst %v", &max):
					s.addInc(false, false, &h, 0, max-1)
				case in.Parse("rand%*om src %v-%v", &min, &max):
					s.addInc(true, true, &h, min, max)
				case in.Parse("rand%*om dst %v-%v", &min, &max):
					s.addInc(false, true, &h, min, max)
				default:
					break incLoop
				}
			}
			s.AddHeader(&h)
			if t, ok := m.protocolMap[h.Protocol]; ok {
				var sub_r pg.Streamer
				sub_r, err = t.ParseStream(in)
				if err != nil {
					err = fmt.Errorf("ip4 %s: %s `%s'", t.Name(), err, in)
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

func (s *pgStream) addInc(isSrc, isRandom bool, h *Header, min, max uint64) {
	if max < min {
		max = min
	}
	ai := &addressIncrement{
		s:        s,
		min:      min,
		max:      max,
		cur:      min,
		isSrc:    isSrc,
		isRandom: isRandom,
	}
	if isSrc {
		ai.base = h.Src
	} else {
		ai.base = h.Dst
	}
	s.DataHooks.Add(ai.Do)
}

type addressIncrement struct {
	s        *pgStream
	base     Address
	cur      uint64
	min      uint64
	max      uint64
	isSrc    bool
	isRandom bool
}

func (ai *addressIncrement) Do(dst []vnet.Ref, dataOffset uint) {
	for i := range dst {
		h := (*Header)(dst[i].DataOffset(dataOffset))
		v := ai.cur
		if ai.isRandom {
			v = uint64(rand.Intn(int(1 + ai.max - ai.min)))
		}
		a := &h.Dst
		if ai.isSrc {
			a = &h.Src
		}
		*a = ai.base
		a.Add(v)
		h.Checksum = h.ComputeChecksum()
		ai.cur++
		if ai.cur > ai.max {
			ai.cur = ai.min
		}
	}
}

func (m *pgMain) pgInit(v *vnet.Vnet) {
	m.v = v
	pg.AddStreamType(v, "ip4", m)
}
