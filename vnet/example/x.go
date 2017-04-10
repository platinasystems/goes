// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/arp"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	ipcli "github.com/platinasystems/go/vnet/ip/cli"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"

	"fmt"
	"os"
	"time"
)

type stream struct {
	random          bool
	random_size     bool
	random_seed     int64
	cur_size        uint
	min_size        uint
	max_size        uint
	n_packets_limit uint
	n_packets_sent  uint
	next            uint
}

const max_buffer_size = 16 << 10

func (s *stream) validate_size() {
	if s.min_size > max_buffer_size {
		s.min_size = max_buffer_size
	}
	if s.max_size < s.min_size {
		s.max_size = s.min_size
	}
	if s.max_size > max_buffer_size {
		s.max_size = max_buffer_size
	}
	s.cur_size = s.min_size
}

type myNode struct {
	vnet.InterfaceNode
	ethernet.Interface
	vnet.Package
	pool           vnet.BufferPool
	isUnix         bool
	verbose_output bool
	stream
}

var (
	MyNode        = &myNode{}
	myNodePackage uint
)

const (
	error_one = iota
	error_two
	tx_packets_dropped
	n_error
)

const (
	next_error = iota
	next_punt
	n_next
)

func init() {
	vnet.AddInit(func(v *vnet.Vnet) {
		MyNode.Errors = []string{
			error_one:          "error one",
			error_two:          "error two",
			tx_packets_dropped: "tx packets dropped",
		}
		MyNode.Next = []string{
			next_error: "error",
			next_punt:  "punt",
		}

		v.RegisterHwInterface(MyNode, "my-node")
		v.RegisterInterfaceNode(MyNode, MyNode.Hi(), "my-node")
		MyNode.stream = stream{n_packets_limit: 1, min_size: 64, max_size: 64, next: next_error}

		if false {
			v.CliAdd(&cli.Command{
				Name:      "a",
				ShortHelp: "a short help",
				Action: func(c cli.Commander, w cli.Writer, in *cli.Input) error {
					n := MyNode
					n.n_packets_sent = 0 // reset
					for !in.End() {
						var (
							next_name string
							count     float64
						)
						switch {
						case (in.Parse("c%*ount %f", &count) || in.Parse("%f", &count)) && count >= 0:
							n.n_packets_limit = uint(count)
						case in.Parse("s%*ize %d %d", &n.min_size, &n.max_size):
						case in.Parse("s%*ize %d", &n.min_size):
							n.max_size = n.min_size
						case in.Parse("n%*ext %s", &next_name):
							n.next = v.AddNamedNext(n, next_name)
						case in.Parse("r%*andom"):
							n.random = true
						default:
							return cli.ParseError
						}
					}
					n.validate_size()
					n.Activate(true)
					return nil
				},
			})
		}
	})
}

func (n *myNode) ValidateSpeed(speed vnet.Bandwidth) (err error) { return }

const (
	s1_counter vnet.HwIfCounterKind = iota
	s2_counter
)

const (
	c1_counter vnet.HwIfCombinedCounterKind = iota
	c2_counter
)

func (n *myNode) GetHwInterfaceCounterNames() (nm vnet.InterfaceCounterNames) {
	nm.Single = []string{
		s1_counter: "s1",
		s2_counter: "s2",
	}
	nm.Combined = []string{
		c1_counter: "c1",
		c2_counter: "c2",
	}
	return
}

func (n *myNode) GetHwInterfaceCounterValues(t *vnet.InterfaceThread) { return }

func ip4Template(t *hw.BufferTemplate) {
	t.Data = vnet.MakePacket(
		&ethernet.Header{
			Type: ethernet.IP4.FromHost(),
			Src:  ethernet.Address{0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5},
			Dst:  ethernet.Address{0xea, 0xeb, 0xec, 0xed, 0xee, 0xef},
		},
		&ip4.Header{
			Protocol: ip.UDP,
			Src:      ip4.Address{0x1, 0x2, 0x3, 0x4},
			Dst:      ip4.Address{0x5, 0x6, 0x7, 0x8},
			Tos:      0,
			Ttl:      255,
			Ip_version_and_header_length: 0x45,
			Fragment_id:                  vnet.Uint16(0x1234).FromHost(),
			Flags_and_fragment_offset:    ip4.DontFragment.FromHost(),
		},
		&vnet.IncrementingPayload{Count: t.Size - ethernet.HeaderBytes - ip4.HeaderBytes},
	)
}

func arpTemplate(t *hw.BufferTemplate) {
	t.Data = vnet.MakePacket(
		&ethernet.Header{
			Type: ethernet.ARP.FromHost(),
			Src:  ethernet.Address{0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5},
			Dst:  ethernet.BroadcastAddr,
		},
		&arp.HeaderEthernetIp4{
			Header: arp.Header{
				Opcode:          arp.Request.FromHost(),
				L2Type:          arp.L2TypeEthernet.FromHost(),
				L3Type:          vnet.Uint16(ethernet.IP4.FromHost()),
				NL2AddressBytes: ethernet.AddressBytes,
				NL3AddressBytes: ip4.AddressBytes,
			},
			Addrs: [2]arp.EthernetIp4Addr{
				arp.EthernetIp4Addr{
					Ethernet: ethernet.Address{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5},
					Ip4:      ip4.Address{10, 11, 12, 13},
				},
				arp.EthernetIp4Addr{
					Ethernet: ethernet.Address{0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5},
					Ip4:      ip4.Address{20, 21, 22, 23},
				},
			},
		},
	)
}

func (n *myNode) Init() (err error) {
	v := n.Vnet
	config := &ethernet.InterfaceConfig{
		Address: ethernet.Address{0, 1, 2, 3, 4, 5},
	}
	ethernet.RegisterInterface(v, MyNode, config, "my-node")

	// Link is always up for packet generator.
	n.SetLinkUp(true)
	n.SetAdminUp(true)

	t := &n.pool.BufferTemplate
	*t = vnet.DefaultBufferPool.BufferTemplate
	t.Size = max_buffer_size
	if true {
		ip4Template(t)
	} else {
		arpTemplate(t)
	}
	n.pool.Name = n.Name()
	v.AddBufferPool(&n.pool)

	// Enable for event test
	if false {
		n.AddTimedEvent(&myEvent{}, 1)
	}
	return
}

func (n *myNode) Configure(in *parse.Input) {
	switch {
	case in.Parse("tuntap"):
		n.isUnix = true
	case in.Parse("verbose"):
		n.verbose_output = true
	default:
		panic(parse.ErrInput)
	}
}

func (n *myNode) IsUnix() bool       { return n.isUnix }
func (n *myNode) DriverName() string { return "my-node" }

func (n *myNode) InterfaceInput(o *vnet.RefOut) {
	out := &o.Outs[n.next]
	out.BufferPool = &n.pool
	t := n.GetIfThread()

	cap := out.Cap()
	np := cap
	if n.n_packets_limit != 0 {
		np = 0
		if n.n_packets_sent < n.n_packets_limit {
			np = n.n_packets_limit - n.n_packets_sent
			if np > cap {
				np = cap
			}
		}
	}

	out.AllocPoolRefs(&n.pool, np)
	rs := out.Refs[:]
	nBytes := uint(0)
	for i := uint(0); i < np; i++ {
		r := &rs[i]
		n.SetError(r, uint(i%n_error))

		r.SetDataLen(n.cur_size)
		n.cur_size++
		if n.cur_size > n.max_size {
			n.cur_size = n.min_size
		}
		nBytes += r.DataLen()
	}
	vnet.IfRxCounter.Add(t, n.Si(), np, nBytes)
	c1_counter.Add(t, n.Hi(), np, nBytes)
	s1_counter.Add(t, n.Hi(), np)
	out.SetLen(n.Vnet, np)
	n.n_packets_sent += np
	if n.n_packets_limit != 0 {
		n.Activate(n.n_packets_sent < n.n_packets_limit)
	}
}

func (n *myNode) InterfaceOutput(in *vnet.TxRefVecIn) {
	if false {
		// Enable to test poller suspend/resume.
		// Send my-node output to my-node (send to self).
		time.Sleep(1 * time.Second)
	}
	if n.verbose_output {
		// Enable to echo ethernet/ip4.
		for i := range in.Refs {
			eh := (*ethernet.Header)(in.Refs[i].Data())
			ih := (*ip4.Header)(in.Refs[i].DataOffset(14))
			fmt.Printf("%s %s\n", eh, ih)
		}
	}
	n.CountError(tx_packets_dropped, in.NPackets())
	n.Vnet.FreeTxRefIn(in)
}

func main() {
	v := &vnet.Vnet{}

	// Select packages we want to run with.
	unix.Init(v)
	ethernet.Init(v)
	ip4.Init(v)
	ip6.Init(v)
	ixge.Init(v)
	pg.Init(v)
	ipcli.Init(v)
	myNodePackage = v.AddPackage("my-node", MyNode)

	var in parse.Input
	in.Add(os.Args[1:]...)
	err := v.Run(&in)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}

type myEvent struct {
	vnet.Event
	x int
}

func (e *myEvent) EventAction() {
	e.x++
	e.AddTimedEvent(e, 1)
}

func (e *myEvent) String() string { return fmt.Sprintf("my-event %d", e.x) }
