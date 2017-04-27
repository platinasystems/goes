// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/vnet"

	"fmt"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

func (h *msghdr) set(a *syscall.RawSockaddrLinklayer, iovs []iovec) {
	h.Name = (*byte)(unsafe.Pointer(a))
	h.Namelen = syscall.SizeofSockaddrLinklayer
	h.Iov = (*syscall.Iovec)(&iovs[0])
	h.Iovlen = uint64(len(iovs))
}

var raw_sockaddr_ll_template = syscall.RawSockaddrLinklayer{
	Family: syscall.AF_PACKET,
}

func (v *packet_vector) add_packet(iovs []iovec, ifindex uint32) {
	i := v.i
	v.i++
	a := &v.a[i]
	*a = raw_sockaddr_ll_template
	a.Ifindex = int32(ifindex)
	v.m[i].msg_hdr.set(a, iovs)
}

func (p *packet) alloc_refs(rx *rx_node, n uint) {
	rx.buffer_pool.AllocRefs(p.refs[:n])
	for i := uint(0); i < n; i++ {
		p.iovs[i].Base = (*byte)(p.refs[i].Data())
		p.iovs[i].Len = uint64(rx.buffer_pool.Size)
	}
}

func (p *packet) rx_init(rx *rx_node) {
	n := rx.max_buffers_per_packet
	p.iovs.Validate(n - 1)
	p.refs.Validate(n - 1)
	p.iovs = p.iovs[:n]
	p.refs = p.refs[:n]
	p.alloc_refs(rx, n)
}

func (p *packet) rx_free(rx *rx_node) { rx.buffer_pool.FreeRefs(&p.refs[0], p.refs.Len(), false) }

const (
	packet_vector_max_len = 64
	max_rx_packet_size    = 16 << 10
)

// Maximum sized packet vector.
type packet_vector struct {
	i uint
	a [packet_vector_max_len]syscall.RawSockaddrLinklayer
	m [packet_vector_max_len]mmsghdr
	p [packet_vector_max_len]packet
}

func (n *rx_node) new_packet_vector() (v *packet_vector) {
	v = &packet_vector{}
	for i := range v.p {
		v.p[i].rx_init(n)
	}
	return
}

func (n *rx_node) get_packet_vector() (v *packet_vector) {
	select {
	case v = <-n.pv_pool:
	default:
		v = n.new_packet_vector()
	}
	return
}

func (n *rx_node) put_packet_vector(v *packet_vector) { n.pv_pool <- v }

func (n *rx_node) get_rx_ref_vector() (v *rx_ref_vector) {
	select {
	case v = <-n.rv_pool:
	default:
		v = &rx_ref_vector{}
	}
	return
}

func (n *rx_node) put_rx_ref_vector(v *rx_ref_vector) { n.rv_pool <- v }

func (n *rx_node) init(v *vnet.Vnet) {
	n.Next = []string{
		rx_node_next_error: "error",
	}
	n.Errors = []string{
		rx_error_drop:               "drops",
		rx_error_non_vnet_interface: "not vnet interface",
	}
	v.RegisterInputNode(n, "unix-rx")
	n.buffer_pool = vnet.DefaultBufferPool
	v.AddBufferPool(n.buffer_pool)
	n.pv_pool = make(chan *packet_vector, 64)
	n.rv_pool = make(chan *rx_ref_vector, 64)
	n.rv_input = make(chan *rx_ref_vector, 64)
	n.max_buffers_per_packet = max_rx_packet_size / n.buffer_pool.Size
	if max_rx_packet_size%n.buffer_pool.Size != 0 {
		n.max_buffers_per_packet++
	}
}

type rx_node struct {
	vnet.InputNode
	buffer_pool            *vnet.BufferPool
	pv_pool                chan *packet_vector
	rv_pool                chan *rx_ref_vector
	rv_input               chan *rx_ref_vector
	rv_pending             *rx_ref_vector
	max_buffers_per_packet uint
	active_lock            sync.Mutex
	active_count           int32
}

type rx_node_next uint32

const (
	rx_node_next_error rx_node_next = iota
)

const (
	rx_error_drop = iota
	rx_error_non_vnet_interface
)

func (v *rx_ref_vector) rx(p *packet, rx *rx_node, i, n_bytes_in_packet, ifindex uint) {
	size := rx.buffer_pool.Size
	n_left := n_bytes_in_packet
	var n_refs uint
	for n_refs = 0; n_left > 0; n_refs++ {
		l := size
		if n_left < l {
			l = n_left
		}
		r := &p.refs[n_refs]
		r.SetDataLen(l)
		if r.NextValidFlag() != 0 {
			panic("next")
		}
		p.chain.Append(r)
		n_left -= l
	}
	p.alloc_refs(rx, n_refs)
	ref := p.chain.Done()
	ref.SetError(&rx.Node, rx_error_non_vnet_interface)
	ref.Si = vnet.SiNil // fixme
	v.refs[i] = ref
	v.nexts[i] = rx_node_next_error
	v.lens[i] = uint32(n_bytes_in_packet)
	return
}

func (m *msghdr) ifindex() uint32 {
	a := (*syscall.RawSockaddrLinklayer)(unsafe.Pointer(m.Name))
	return uint32(a.Ifindex)
}

type rx_ref_vector struct {
	n_packets uint
	refs      [packet_vector_max_len]vnet.Ref
	lens      [packet_vector_max_len]uint32
	nexts     [packet_vector_max_len]rx_node_next
}

func (ns *net_namespace) ReadReady() (err error) {
	rx := &ns.m.rx_node
	v := rx.get_packet_vector()
	n_packets, errno := recvmmsg(ns.Fd, syscall.MSG_WAITFORONE, v.m[:])
	if errno != 0 {
		err = errorForErrno("readv", errno)
		if n_packets != 0 {
			panic(fmt.Errorf("ReadReady error %s but n packets %d > 0", err, n_packets))
		}
		rx.put_packet_vector(v)
		rx.CountError(rx_error_drop, 1)
		return
	}
	rv := rx.get_rx_ref_vector()
	rv.n_packets = uint(n_packets)
	for i := 0; i < n_packets; i++ {
		p := &v.p[i]
		m := &v.m[i]
		rv.rx(p, rx, uint(i), uint(m.msg_len), uint(m.msg_hdr.ifindex()))
	}

	rx.rv_input <- rv
	rx.active_lock.Lock()
	atomic.AddInt32(&rx.active_count, int32(n_packets))
	rx.Activate(true)
	rx.active_lock.Unlock()

	// Return packet vector for reuse.
	rx.put_packet_vector(v)

	return
}

func (rx *rx_node) copy_pending(rv *rx_ref_vector, i uint) {
	p := rx.rv_pending
	if p == nil {
		p = rv
	}
	n := rv.n_packets
	n_left := n - i
	copy(p.refs[:n_left], rv.refs[i:])
	copy(p.lens[:n_left], rv.lens[i:])
	copy(p.nexts[:n_left], rv.nexts[i:])
	p.n_packets = n_left
	rx.rv_pending = p
}

func (rx *rx_node) input_ref_vector(rv *rx_ref_vector, o *vnet.RefOut, n_packets, n_bytes *uint) (done bool) {
	np, nb := *n_packets, *n_bytes
	var i uint
	for i = 0; i < rv.n_packets; i++ {
		out := &o.Outs[rv.nexts[i]]
		l := out.Len()
		if done = l == vnet.MaxVectorLen; done {
			rx.copy_pending(rv, i)
			break
		}
		out.Refs[l] = rv.refs[i]
		out.SetLen(rx.Vnet, l+1)
		np++
		nb += uint(rv.lens[i])
	}
	if i >= rv.n_packets {
		rx.put_rx_ref_vector(rv)
	}
	*n_packets, *n_bytes = np, nb
	return
}

func (rx *rx_node) NodeInput(out *vnet.RefOut) {
	n_packets, n_bytes := uint(0), uint(0)

	done := false
	if rx.rv_pending != nil {
		done = rx.input_ref_vector(rx.rv_pending, out, &n_packets, &n_bytes)
	}
	for !done {
		select {
		case rv := <-rx.rv_input:
			done = rx.input_ref_vector(rv, out, &n_packets, &n_bytes)
		default:
			done = true
		}
	}

	rx.active_lock.Lock()
	rx.Activate(atomic.AddInt32(&rx.active_count, -int32(n_packets)) > 0)
	rx.active_lock.Unlock()
}
