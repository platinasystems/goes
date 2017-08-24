// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/gre"
	ipcli "github.com/platinasystems/go/vnet/ip/cli"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	"github.com/platinasystems/go/vnet/unix"

	"fmt"
	"os"
	"time"
)

type myInterface struct {
	vnet.InterfaceNode
	ethernet.Interface
	n *myNode
}

type myNode struct {
	intfs []myInterface
	vnet.Package
	isUnix                bool
	verbose_output        bool
	interface_name_format string
	interface_count       uint
	inject_node           inject_node
}

var (
	MyNode = &myNode{
		interface_name_format: "e%d",
		interface_count:       1,
	}
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

func (n *myNode) ValidateSpeed(speed vnet.Bandwidth) (err error) { return }

const (
	s1_counter vnet.HwIfCounterKind = iota
	s2_counter
)

const (
	c1_counter vnet.HwIfCombinedCounterKind = iota
	c2_counter
)

func (n *myInterface) GetHwInterfaceCounterNames() (nm vnet.InterfaceCounterNames) {
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

func (n *myInterface) GetHwInterfaceCounterValues(t *vnet.InterfaceThread) { return }

func (n *myNode) Init() (err error) {
	v := n.Vnet
	v.RegisterOutputNode(&n.inject_node, "example-inject")
	n.intfs = make([]myInterface, n.interface_count)
	for i := range n.intfs {
		intf := &n.intfs[i]
		intf.n = n
		config := &ethernet.InterfaceConfig{
			Address: ethernet.Address{0xfe, 0xdc, 0xba, 0, 0, 0},
		}
		config.Address.Add(uint64(i))
		ethernet.RegisterInterface(v, intf, config, n.interface_name_format, i)

		intf.Errors = []string{
			error_one:          "error one",
			error_two:          "error two",
			tx_packets_dropped: "tx packets dropped",
		}
		intf.Next = []string{
			next_error: "error",
			next_punt:  "punt",
		}
		v.RegisterInterfaceNode(intf, intf.Hi(), n.interface_name_format, i)

		// Admin/link is always up.
		if err = intf.SetLinkUp(true); err != nil {
			return
		}
		if err = intf.SetAdminUp(true); err != nil {
			return
		}

		// Enable for event test
		if false {
			intf.AddTimedEvent(&myEvent{}, 1)
		}
	}
	return
}

func (n *myNode) Configure(in *parse.Input) {
	for !in.End() {
		switch {
		case in.Parse("unix"):
			n.isUnix = true
		case in.Parse("count %d", &n.interface_count):
		case in.Parse("name %s", &n.interface_name_format):
		case in.Parse("verbose"):
			n.verbose_output = true
		default:
			panic(parse.ErrInput)
		}
	}
}

func (n *myInterface) IsUnix() bool       { return n.n.isUnix }
func (i *myInterface) DriverName() string { return "example" }

func (n *myInterface) InterfaceInput(o *vnet.RefOut) {
	panic("ga")
}

func (n *myInterface) InterfaceOutput(in *vnet.TxRefVecIn) {
	if false {
		// Enable to test poller suspend/resume.
		time.Sleep(1 * time.Second)
	}
	if n.n.verbose_output {
		for i := range in.Refs {
			fmt.Printf("%s: %x\n", n.Name(), in.Refs[i].DataSlice())
		}
	}
	n.CountError(tx_packets_dropped, in.NPackets())
	n.Vnet.FreeTxRefIn(in)
}

type inject_node struct {
	vnet.OutputNode
}

func (n *inject_node) NodeOutput(in *vnet.RefIn) {
	l := in.InLen()
	for i := uint(0); i < l; i++ {
		r := in.Refs[i]
		fmt.Printf("%s %s: %x\n", n.Name(), r.Si.Name(n.Vnet), r.DataSlice())
	}
	in.FreeRefs(l)
}

func main() {
	v := &vnet.Vnet{}

	// Select packages we want to run with.
	unix.Init(v, unix.Config{RxInjectNodeName: "example-inject"})
	m4 := ip4.Init(v)
	m6 := ip6.Init(v)
	ethernet.Init(v, m4, m6)
	gre.Init(v)
	ixge.Init(v)
	pg.Init(v)
	ipcli.Init(v)
	myNodePackage = v.AddPackage("example", MyNode)

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
