// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip4

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/icmp4"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/pg"

	"fmt"
	"math/rand"
)

type pgStream struct {
	pg.Stream
	ai_src, ai_dst addressIncrement
}

type pgMain struct {
	v           *vnet.Vnet
	protocolMap map[ip.Protocol]pg.StreamType
	icmpMain
}

func (m *pgMain) initProtocolMap() {
	if m.protocolMap != nil {
		return
	}
	m.protocolMap = make(map[ip.Protocol]pg.StreamType)
	m.protocolMap[ip.ICMP] = pg.GetStreamType(m.v, "icmp4")
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
	ai := addressIncrement{
		min:      min,
		max:      max,
		cur:      min,
		isRandom: isRandom,
	}
	if isSrc {
		ai.base = h.Src
		s.ai_src = ai
	} else {
		ai.base = h.Dst
		s.ai_dst = ai
	}
}

type addressIncrement struct {
	base     Address
	cur      uint64
	min      uint64
	max      uint64
	isRandom bool
}

func (ai *addressIncrement) valid() bool { return ai.max != ai.min }

func (ai *addressIncrement) do(dst []vnet.Ref, dataOffset uint, isSrc bool) {
	for i := range dst {
		h := (*Header)(dst[i].DataOffset(dataOffset))
		v := ai.cur
		if ai.isRandom {
			v = uint64(rand.Intn(int(1 + ai.max - ai.min)))
		}
		a := &h.Dst
		if isSrc {
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

func (s *pgStream) Finalize(r []vnet.Ref, data_offset uint) (changed bool) {
	if s.ai_src.valid() {
		s.ai_src.do(r, data_offset, true)
		changed = true
	}
	if s.ai_dst.valid() {
		s.ai_dst.do(r, data_offset, false)
		changed = true
	}
	if s.IsVariableSize() {
		s.setLength(r, data_offset)
		changed = true
	}
	return
}

func (s *pgStream) setLength(dst []vnet.Ref, dataOffset uint) {
	for i := range dst {
		r := &dst[i]
		h := (*Header)(r.DataOffset(dataOffset))
		h.Length.Set(r.ChainLen() - dataOffset)
		h.Checksum = h.ComputeChecksum()
	}
}

type icmpMain struct {
}

type icmpStream struct {
	pg.Stream
}

func (m *icmpMain) Name() string { return "icmp4" }

func (m *icmpMain) ParseStream(in *parse.Input) (r pg.Streamer, err error) {
	var s icmpStream
	h := icmp4.Header{}
	for !in.End() {
		switch {
		case in.Parse("%v", &h):
			s.AddHeader(&h)
			switch h.Type {
			case icmp4.Echo_request, icmp4.Echo_reply:
				s.AddHeader(&icmp4.EchoRequest{})
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

func (s *icmpStream) Finalize(dst []vnet.Ref, do uint) (changed bool) {
	if changed = s.IsVariableSize(); !changed {
		return
	}
	for i := range dst {
		r := &dst[i]
		h := (*icmp4.Header)(r.DataOffset(do))
		h.Checksum = 0
		sum := ip.Checksum(0).AddRef(r, do)
		h.Checksum = ^sum.Fold()
	}
	return
}

func (m *pgMain) pgInit(v *vnet.Vnet) {
	m.v = v
	pg.AddStreamType(v, "ip4", m)
	pg.AddStreamType(v, "icmp4", &m.icmpMain)
}
