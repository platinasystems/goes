// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/vnet"

	"syscall"
	"unsafe"
)

func (h *msghdr) set(a *syscall.RawSockaddrLinklayer, iovs iovecVec) {
	h.Name = (*byte)(unsafe.Pointer(a))
	h.Namelen = syscall.SizeofSockaddrLinklayer
	h.Iov = (*syscall.Iovec)(&iovs[0])
	h.Iovlen = uint64(len(iovs))
}

var raw_sockaddr_ll_template = syscall.RawSockaddrLinklayer{
	Family: syscall.AF_PACKET,
}

func (v *packet_vector) add_packet(p *packet, ifindex uint32) {
	i := v.i
	v.i++
	a := &v.a[i]
	*a = raw_sockaddr_ll_template
	a.Ifindex = int32(ifindex)
	v.mm[i].msg_hdr.set(a, p.iovs)
	v.p[i] = *p
}

const packet_vector_max_len = 64

// Maximum sized packet vector.
type packet_vector struct {
	i  uint
	a  [packet_vector_max_len]syscall.RawSockaddrLinklayer
	mm [packet_vector_max_len]mmsghdr
	p  [packet_vector_max_len]packet
}

func (m *net_namespace_main) get_packet_vector() (v *packet_vector) {
	select {
	case v = <-m.pv:
	default:
		v = &packet_vector{}
	}
	return
}

func (m *net_namespace_main) put_packet_vector(v *packet_vector) { m.pv <- v }

type rx_tx_node_main struct {
	pv chan *packet_vector
	// Common MTU for tuntap interfaces.
	mtu_bytes   uint
	buffer_pool *vnet.BufferPool
}

func (m *rx_tx_node_main) init(v *vnet.Vnet) {
	m.buffer_pool = vnet.DefaultBufferPool
	v.AddBufferPool(m.buffer_pool)
}

type rx_tx_node_common struct {
	m    *net_namespace_main
	ns   *net_namespace
	name string
}

func (n *rx_tx_node_common) add(m *net_namespace_main, ns *net_namespace, rx_tx string) {
	n.m = m
	n.ns = ns
	n.name = "unix-" + rx_tx
	if len(ns.name) > 0 {
		n.name += "-" + ns.name
	}
}

type rx_node struct {
	rx_tx_node_common
	vnet.InputNode
}

func (n *rx_node) add(m *net_namespace_main, ns *net_namespace) {
	n.rx_tx_node_common.add(m, ns, "rx")
	n.m.m.v.RegisterInputNode(n, n.name)
}

func (n *rx_node) NodeInput(out *vnet.RefOut) {
	panic("rx")
}

func (ns *net_namespace) ReadReady() (err error) {
	panic("rx")
	n := &ns.rx_node
	_ = n
	return
}
