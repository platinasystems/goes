package ip4

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/pg"

	"fmt"
)

type pgStream struct {
	pg.Stream
	h []vnet.PacketHeader
}

func (s *pgStream) PacketHeaders() []vnet.PacketHeader { return s.h }

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
		switch {
		case in.Parse("%v", &h):
			s.h = append(s.h, &h)
			if t, ok := m.protocolMap[h.Protocol]; ok {
				var sub_r pg.Streamer
				sub_r, err = t.ParseStream(in)
				if err != nil {
					err = fmt.Errorf("ip4 %s: %s `%s'", t.Name(), err, in)
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
	pg.AddStreamType(v, "ip4", m)
}
