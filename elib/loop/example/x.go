// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/arp"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"

	"fmt"
)

type myNode struct {
	loop.Node
	myErr [n_error]loop.ErrorRef
	pool  loop.BufferPool
}

var MyNode = &myNode{}

func init() { loop.Register(MyNode, "my-node") }

type out struct {
	loop.Out
	Outs []loop.RefIn
}

func (n *myNode) MakeLoopOut() loop.LooperOut { return &out{} }

const (
	error_one = iota
	error_two
	n_error
)

var errorStrings = [...]string{
	error_one: "error one",
	error_two: "error two",
}

func (n *myNode) LoopInit(l *loop.Loop) {
	l.AddNext(n, loop.ErrorNode)
	for i := range errorStrings {
		n.myErr[i] = n.NewError(errorStrings[i])
	}
	t := &n.pool.BufferTemplate
	*t = *loop.DefaultBufferTemplate
	t.Size = 2048
	if false {
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
	} else {
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
					L3Type:          ethernet.IP4.FromHost(),
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
	n.pool.Init()
}

func (n *myNode) LoopInput(l *loop.Loop, lo loop.LooperOut) {
	o := lo.(*out)
	toErr := &o.Outs[0]
	toErr.AllocPoolRefs(&n.pool)
	rs := toErr.Refs[:]
	for i := range rs {
		r := &rs[i]
		if true {
			eh := ethernet.GetPacketHeader(r)
			l.Logf("%s %d: %s\n", n.NodeName(), i, eh)
			r.Advance(ethernet.HeaderBytes)
			if false {
				ih := ip4.GetHeader(r)
				l.Logf("%d: %s\n", i, ih)
			} else {
				ah := arp.GetHeader(r)
				l.Logf("%d: %s\n", i, ah)
			}
		}
		r.Err = n.myErr[i%n_error]
	}
	toErr.SetLen(l, uint(len(toErr.Refs)))
}

func init() {
	loop.CliAdd(&cli.Command{
		Name:      "a",
		ShortHelp: "a short help",
		Action: func(c cli.Commander, w cli.Writer, s *cli.Scanner) {
			n := uint(1)
			if s.Peek() != cli.EOF {
				if err := s.Parse("%d", &n); err != nil {
					fmt.Fprintln(w, "parse error")
					return
				}
			}
			if n == 0 {
				MyNode.Activate(true)
			} else {
				MyNode.ActivateCount(n)
			}
		},
	})
}

func main() {
	loop.Run()
}
