// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/arp"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"

	"fmt"
	"math/rand"
	"os"
	"time"
)

type stream struct {
	random      bool
	random_size bool
	random_seed int64
	next        uint
	n_packets   uint
	cur_size    uint
	min_size    uint
	max_size    uint
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
	pool   hw.BufferPool
	isUnix bool
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

func (n *myNode) Configure(in *parse.Input) {}

func (n *myNode) register(v *vnet.Vnet) {
	n.Next = []string{
		0: "bcm-cpu",
		1: "error",
	}
	portMap := make(map[string]uint)
	n.next = 0
	for i := 0; i < 32; i++ {
		name := fmt.Sprintf("en-%02d-0", i)
		portMap[name] = uint(len(n.Next))
		n.Next = append(n.Next, name)
	}
	n.Errors = []string{
		0: "error 1",
	}

	v.RegisterInterfaceNode(n, n.HwIf.Hi(), "pg-rx")

	v.CliAdd(&cli.Command{
		Name:      "a",
		ShortHelp: "a short help",
		Action: func(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
			n.n_packets = 1
			n.next = 0
			n.min_size = 64
			n.max_size = 64
			for !in.End() {
				var v uint
				switch {
				case in.Parse("%v", parse.StringMap(portMap), &v):
					n.next = v
				case in.Parse("random"):
					n.random = true
				case in.Parse("size %d %d", &n.min_size, &n.max_size):
				case in.Parse("size %d", &n.min_size):
					n.max_size = n.min_size
				case in.Parse("%d", &n.n_packets):
				default:
					err = cli.ParseError
					return
				}
			}
			n.validate_size()
			n.Activate(true)
			return
		},
	})
}

func ip4Template(t *hw.BufferTemplate) {
	t.Data = vnet.MakePacket(
		&ethernet.Header{
			Type: ethernet.IP4.FromHost(),
			// Type: ethernet.VLAN.FromHost(),
			Src: ethernet.Address{0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5},
			Dst: ethernet.Address{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0x82},
		},
		// &ethernet.VlanHeader{
		// 	Type:                ethernet.IP4.FromHost(),
		// 	Priority_cfi_and_id: vnet.Uint16(0x1).FromHost(),
		// },
		&ip4.Header{
			Protocol: ip.ICMP,
			Src:      ip4.Address{0x1, 0x2, 0x3, 0x4},
			Dst:      ip4.Address{0x5, 0x6, 0x7, 0x8},
			Tos:      0,
			Ttl:      64,
			Ip_version_and_header_length: 0x45,
			Fragment_id:                  vnet.Uint16(0x1234).FromHost(),
			Flags_and_fragment_offset:    ip4.DontFragment.FromHost(),
		},
		&vnet.IncrementingPayload{Count: t.Size - ethernet.HeaderBytes - ip4.HeaderBytes},
	)
}

func arpTemplate(t *hw.BufferTemplate, isReply bool) {
	e := &ethernet.Header{
		Type: ethernet.ARP.FromHost(),
		Src:  ethernet.Address{0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5},
		Dst:  ethernet.BroadcastAddr,
	}
	a := &arp.HeaderEthernetIp4{
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
	}

	if isReply {
		a.Header.Opcode = arp.Reply.FromHost()
		e.Dst = ethernet.Address{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0x82}
	}

	t.Data = vnet.MakePacket(e, a)
}

func (n *myNode) Init() (err error) {
	v := n.Vnet
	n.register(v)
	v.RegisterHwInterface(n, "my-node")

	// Link is always up for packet generator.
	n.SetLinkUp(true)
	n.SetAdminUp(true)

	t := &n.pool.BufferTemplate
	*t = *hw.DefaultBufferTemplate
	t.Size = max_buffer_size
	if true {
		ip4Template(t)
	} else {
		arpTemplate(t, false)
	}
	n.pool.Name = "my-node"
	v.AddBufferPool((*vnet.BufferPool)(&n.pool))

	// dumps pcap file for import into wireshark.
	if false {
		f, _ := os.Create("/tmp/x.pcap")
		w := pcapgo.NewWriter(f)
		w.WriteFileHeader(65536, layers.LinkTypeEthernet) // new file, must do this.
		w.WritePacket(gopacket.CaptureInfo{
			CaptureLength: len(t.Data),
			Length:        len(t.Data),
		}, t.Data)
		f.Close()
	}
	return
}

func (n *myNode) validate(r *vnet.Ref) {
	if elib.Debug {
		if r.DataLen() != uint(len(n.pool.Data)) {
			panic("size")
		}
		if d := r.DataSlice(); string(d) != string(n.pool.Data) {
			panic("data")
		}
	}
}

func (n *myNode) InterfaceInput(o *vnet.RefOut) {
	toOut := &o.Outs[n.next]
	toOut.BufferPool = (*vnet.BufferPool)(&n.pool)
	nPak := n.n_packets
	if nPak > 1 && n.random {
		nPak = uint(1 + rand.Intn(int(nPak-1)))
	}
	if nPak > uint(len(toOut.Refs)) {
		nPak = uint(len(toOut.Refs))
	}
	n.n_packets -= nPak
	n.pool.AllocRefs(&toOut.Refs[0].RefHeader, nPak)
	t := n.GetIfThread()
	rs := toOut.Refs[:nPak]
	nBytes := uint(0)
	for i := range rs {
		r := &rs[i]
		// n.validate(r)
		n.SetError(r, 0)
		r.SetDataLen(n.cur_size)
		nBytes += r.DataLen()
		n.cur_size++
		if n.cur_size > n.max_size {
			n.cur_size = n.min_size
		}
	}
	vnet.IfRxCounter.Add(t, n.Si(), nPak, nBytes)
	toOut.SetLen(n.Vnet, nPak)
	n.Activate(n.n_packets > 0)
}

func (n *myNode) InterfaceOutput(i *vnet.TxRefVecIn) {
	if false {
		// Enable to test poller suspend/resume.
		// Send my-node output to my-node (send to self).
		time.Sleep(1 * time.Second)
	}
	n.CountError(tx_packets_dropped, i.NPackets())
	n.Vnet.FreeTxRefIn(i)
}

func (n *myNode) SetRewrite(v *vnet.Vnet, rw *vnet.Rewrite, packetType vnet.PacketType, da []byte) {}
func (n *myNode) FormatRewrite(rw *vnet.Rewrite) string                                            { return "" }
func (n *myNode) ValidateSpeed(speed vnet.Bandwidth) (err error)                                   { return }
func (n *myNode) GetHwInterfaceCounters(nm *vnet.InterfaceCounterNames, t *vnet.InterfaceThread)   {}
