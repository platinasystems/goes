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

var defaultHeader = Header{
	Type: IP4.FromHost(),
	Src:  Address{0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5},
	Dst:  Address{0xea, 0xeb, 0xec, 0xed, 0xee, 0xef},
}

func (m *pgMain) ParseStream(in *parse.Input) (r pg.Streamer, err error) {
	m.initTypes()

	var s pgStream
	h := defaultHeader
	for !in.End() {
		switch {
		case in.Parse("%v", &h):
			s.h = append(s.h, &h)
			if t, ok := m.typeMap[h.Type]; ok {
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
