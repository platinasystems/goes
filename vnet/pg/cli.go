// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pg

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"

	"fmt"
	"sort"
)

type default_stream struct {
	Stream
}

func (s *default_stream) PacketHeaders() []vnet.PacketHeader {
	return []vnet.PacketHeader{
		&vnet.IncrementingPayload{Count: s.MaxSize()},
	}
}

const (
	set_limit = 1 << iota
	set_size
	set_rate
	set_next
	set_node
	set_stream
	set_interface
	set_verbose
)

func (m *main) edit_streams(cmder cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	v := m.Vnet
	default_stream_config := stream_config{
		n_packets_limit: 0,
		min_size:        64,
		max_size:        64,
		next:            next_error,
	}
	var set_what uint
	enable, disable := true, false
	stream_name := ""
	c := default_stream_config
	c.si = m.nodes[c.node_index].Si()
	var r Streamer
	for !in.End() {
		var (
			x       float64
			sub_in  parse.Input
			comment parse.Comment
			index   uint
		)
		switch {
		case (in.Parse("c%*ount %f", &x) || in.Parse("%f", &x)) && x >= 0:
			c.n_packets_limit = uint64(x)
			set_what |= set_limit
		case in.Parse("p%*rint %f", &x):
			c.n_packets_per_print = uint64(x)
			set_what |= set_limit
		case in.Parse("si%*ze %d-%d", &c.min_size, &c.max_size):
			set_what |= set_size
		case in.Parse("si%*ze %d", &c.min_size):
			c.max_size = c.min_size
			set_what |= set_size
		case in.Parse("ra%*te %fbps", &x):
			set_what |= set_rate
			c.rate_bits_per_sec = x
			if set_what&set_limit == 0 {
				c.n_packets_limit = 0
			}
		case in.Parse("ra%*te %fpps", &x) || in.Parse("ra%*te %f", &x):
			set_what |= set_rate
			c.rate_packets_per_sec = x
			if set_what&set_limit == 0 {
				c.n_packets_limit = 0
			}
		case in.Parse("random"):
			c.random_size = true
			set_what |= set_size
		case in.Parse("ve%*rbose"):
			c.verbose = true
			set_what |= set_verbose
		case in.Parse("node %d", &c.node_index):
			set_what |= set_node
		case in.Parse("n%*ext %s", &c.next_name):
			set_what |= set_next
		case in.Parse("en%*able"):
			enable = true
		case in.Parse("dis%*able"):
			disable = true
		case in.Parse("na%*me %s", &stream_name):
		case in.Parse("in%*terface %v", &c.si, v):
			set_what |= set_interface
		case in.Parse("%v %v", &m.stream_type_map, &index, &sub_in):
			r, err = m.stream_types[index].ParseStream(&sub_in)
			if err != nil {
				return
			}
			r.get_stream().stream_config = default_stream_config
			set_what |= set_stream
		case in.Parse("%v", &comment):
		default:
			err = cli.ParseError
			return
		}
	}

	// If node was specified only apply changes to that node.
	// Otherwise apply changes to all nodes.
	for i := range m.nodes {
		n := &m.nodes[i]
		if set_what&set_node != 0 {
			if i > 0 {
				break
			}
			n = &m.nodes[c.node_index]
		}
		n.configure_streams(m, &c, w, stream_name, set_what, r)
		n.Activate(enable && !disable)
	}
	return
}

func (n *node) configure_streams(m *main, c *stream_config, w cli.Writer, stream_name string, set_what uint, r Streamer) {
	create := r != nil
	if create {
		n.new_stream(r, stream_name)
	}

	if r != nil {
		r := n.get_stream_by_name(stream_name)
		n.configure_stream(r, m, c, w, create, set_what)
	} else {
		n.stream_pool.Foreach(func(r Streamer) {
			n.configure_stream(r, m, c, w, create, set_what)
		})
	}
}

func (n *node) configure_stream(r Streamer, m *main, c *stream_config, w cli.Writer, create bool, set_what uint) {
	if set_what&set_next != 0 {
		c.next = m.Vnet.AddNamedNext(n, c.next_name)
	}

	s := r.get_stream()
	if create {
		s.stream_config = *c
	} else {
		if set_what&set_size != 0 {
			s.min_size = c.min_size
			s.max_size = c.max_size
			s.random_size = c.random_size
		}
		if set_what&set_limit != 0 {
			s.n_packets_limit = c.n_packets_limit
			s.n_packets_per_print = c.n_packets_per_print
			s.n_packets_sent = 0
		}
		if set_what&set_next != 0 {
			s.next = c.next
		}
		// Set nothing: repeat last run
		if set_what == 0 {
			s.n_packets_sent = 0
		}
		if set_what&set_rate != 0 {
			s.rate_bits_per_sec = c.rate_bits_per_sec
			s.rate_packets_per_sec = c.rate_packets_per_sec
		}
		if set_what&set_interface != 0 {
			s.stream_config.si = c.si
		}
		if set_what&set_verbose != 0 {
			s.stream_config.verbose = c.verbose
		}
	}

	s.last_time = cpu.TimeNow()
	s.credit_packets = 0
	s.n_packets_last_print = 0
	s.w = w
	ave_packet_bits := 8 * .5 * float64(s.min_size+s.max_size)
	if create || set_what&set_rate != 0 {
		if c.rate_bits_per_sec != 0 {
			s.rate_bits_per_sec = c.rate_bits_per_sec
			s.rate_packets_per_sec = s.rate_bits_per_sec / ave_packet_bits
		} else {
			s.rate_bits_per_sec = c.rate_packets_per_sec * ave_packet_bits
			s.rate_packets_per_sec = c.rate_packets_per_sec
		}
	}

	if set_what&(set_stream|set_size) != 0 || create {
		s.setData()
		n.setData(s)
	}
}

type limit uint64

func (l limit) String() string {
	if l == 0 {
		return ""
	} else {
		return fmt.Sprintf("%d", l)
	}
}

func (m *main) show_streams(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	type cli_stream struct {
		Node  uint   `format:"%d" align:"center"`
		Name  string `format:"%-30s" align:"left"`
		Limit string `format:"%16s" align:"right"`
		Sent  uint64 `format:"%16d" align:"right"`
	}
	cs := []cli_stream{}

	sort.Slice(cs, func(i, j int) bool { return cs[i].Name < cs[j].Name })

	for i := range m.nodes {
		n := &m.nodes[i]
		n.stream_pool.Foreach(func(r Streamer) {
			s := r.get_stream()
			cs = append(cs, cli_stream{
				Node:  n.index,
				Name:  s.name,
				Limit: limit(s.n_packets_limit).String(),
				Sent:  s.n_packets_sent,
			})
		})
	}
	elib.Tabulate(cs).Write(w)
	return
}

func (m *main) cli_init() {
	cmds := []cli.Command{
		cli.Command{
			Name:      "packet-generator",
			ShortHelp: "edit or create packet generator streams",
			Action:    m.edit_streams,
		},
		cli.Command{
			Name:      "show packet-generator",
			ShortHelp: "show packet generator streams",
			Action:    m.show_streams,
		},
	}
	for i := range cmds {
		m.Vnet.CliAdd(&cmds[i])
	}
}

type StreamType interface {
	Name() string
	ParseStream(in *parse.Input) (Streamer, error)
}

func AddStreamType(v *vnet.Vnet, name string, t StreamType) {
	m := GetMain(v)
	ti := uint(len(m.stream_types))
	m.stream_types = append(m.stream_types, t)
	m.stream_type_map.Set(name, ti)
}

func GetStreamType(v *vnet.Vnet, name string) (t StreamType) {
	m := GetMain(v)
	if ti, ok := m.stream_type_map[name]; ok {
		t = m.stream_types[ti]
	}
	return
}
