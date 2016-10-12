package pg

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"

	"fmt"
)

type Streamer interface {
	get_stream() *Stream
	Del()
	PacketHeaders() []vnet.PacketHeader
}

func (s *Stream) get_stream() *Stream { return s }
func (s *Stream) Del()                {}

type stream_config struct {
	random_size bool
	// Min, max packet size.
	min_size uint
	max_size uint
	// Number of packets to send or 0 for no limit.
	n_packets_limit uint64

	// Data rate in bits or packets per second.
	rate_bits_per_sec    float64
	rate_packets_per_sec float64

	// Next index relative to input node for this stream.
	next uint
}

type Stream struct {
	name  string
	index uint
	r     Streamer

	random_seed int64

	cur_size uint

	last_time            cpu.Time
	rate_packets_per_sec float64
	credit_packets       float64

	n_packets_sent uint64

	data         []byte
	buffer_types elib.Uint32Vec

	stream_config
}

//go:generate gentemplate -d Package=pg -id stream -d PoolType=stream_pool -d Type=Streamer -d Data=elts github.com/platinasystems/go/elib/pool.tmpl

func (s *Stream) GetSize() uint { return s.cur_size }
func (s *Stream) MaxSize() uint { return s.max_size }

func (s *Stream) next_size(cur, i uint) uint {
	if x := cur + 1 + i; x <= s.max_size {
		return x
	} else {
		return s.min_size + i
	}
}

func (s *Stream) SetData() {
	if s.max_size < s.min_size {
		s.max_size = s.min_size
	}
	s.cur_size = s.min_size
	h := s.r.PacketHeaders()

	// Add incrementing payload to pad to max size.
	l := uint(0)
	for i := range h {
		l += h[i].Len()
	}
	if l < s.MaxSize() {
		h = append(h, &vnet.IncrementingPayload{Count: s.MaxSize() - l})
	}

	s.data = vnet.MakePacket(h...)
}

func (n *node) get_stream(i uint) Streamer { return n.stream_pool.elts[i] }
func (n *node) get_stream_by_name(name string) (r Streamer) {
	if i, ok := n.stream_index_by_name[name]; ok {
		r = n.get_stream(i)
	}
	return
}

func (n *node) new_stream(r Streamer, format string, args ...interface{}) {
	name := fmt.Sprintf(format, args...)
	si, ok := n.stream_index_by_name[name]
	if ok {
		x := n.get_stream(si)
		x.Del()
	} else {
		si = n.stream_pool.GetIndex()
		n.stream_index_by_name.Set(name, si)
	}

	n.stream_pool.elts[si] = r
	s := r.get_stream()
	s.r = r
	s.index = si
	s.name = name
	return
}

func (s *Stream) clean() {
	s.data = nil
	s.name = ""
	s.r = nil
}

func (n *node) del_stream(r Streamer) {
	s := r.get_stream()
	n.stream_pool.PutIndex(s.index)
	delete(n.stream_index_by_name, s.name)
	s.index = ^uint(0)
	s.r.Del()
	s.clean()
}

func (n *node) GetHwInterfaceCounterNames() (nm vnet.InterfaceCounterNames)                      { return }
func (n *node) GetSwInterfaceCounterNames() (nm vnet.InterfaceCounterNames)                      { return }
func (n *node) GetHwInterfaceCounterValues(t *vnet.InterfaceThread)                              {}
func (n *node) FormatAddress() (s string)                                                        { return }
func (n *node) FormatRewrite(rw *vnet.Rewrite) (s string)                                        { return }
func (n *node) SetRewrite(v *vnet.Vnet, rw *vnet.Rewrite, packetType vnet.PacketType, da []byte) {}
func (n *node) ParseRewrite(rw *vnet.Rewrite, in *parse.Input)                                   {}
