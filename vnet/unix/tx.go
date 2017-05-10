// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/iomux"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
	"sync/atomic"
	"syscall"
)

type tx_node struct {
	vnet.OutputNode
	m       *net_namespace_main
	pv_pool chan *tx_packet_vector
	p_pool  chan *tx_packet
}

type net_namespace_tx_node struct {
	n *tx_node

	// Raw socket for transmit path.
	tx_raw_socket_fd int
	iomux.File

	active_count int32
	to_tx        chan *tx_packet_vector
	pv           *tx_packet_vector

	suspend_saved_out *vnet.RefIn
}

func (n *net_namespace_tx_node) init(m *net_namespace_main) {
	n.to_tx = make(chan *tx_packet_vector, vnet.MaxVectorLen)
	n.n = &m.tx_node
}

type tx_packet struct {
	ifindex uint32
	iovs    iovecVec
}

func (p *tx_packet) add_one_ref(r *vnet.Ref, iʹ, lʹ uint) (i, l uint) {
	i, l = iʹ, lʹ
	d := r.DataLen()
	p.iovs.Validate(i)
	p.iovs[i].Base = (*byte)(r.Data())
	p.iovs[i].Len = uint64(d)
	return i + 1, l + d
}

func (p *tx_packet) add_ref(r *vnet.Ref, ifindex uint32) (l uint) {
	p.ifindex = ifindex
	i := uint(0)
	for {
		i, l = p.add_one_ref(r, i, l)
		if r.NextValidFlag() == 0 {
			break
		}
		r = r.NextRef()
	}
	p.iovs = p.iovs[:i]
	return
}

type tx_packet_vector struct {
	n_packets   uint
	ns          *net_namespace
	buffer_pool *vnet.BufferPool
	a           [packet_vector_max_len]syscall.RawSockaddrLinklayer
	m           [packet_vector_max_len]mmsghdr
	p           [packet_vector_max_len]tx_packet
	r           [packet_vector_max_len]vnet.Ref
}

func (n *tx_node) get_packet_vector(p *vnet.BufferPool, ns *net_namespace) (v *tx_packet_vector) {
	select {
	case v = <-n.pv_pool:
	default:
		v = &tx_packet_vector{}
	}
	v.n_packets = 0
	v.buffer_pool = p
	v.ns = ns
	return
}
func (n *tx_node) put_packet_vector(v *tx_packet_vector) { n.pv_pool <- v }

func (v *tx_packet_vector) add_packet(n *tx_node, r *vnet.Ref, ifindex uint32) {
	i := v.n_packets
	v.n_packets++

	p := &v.p[i]
	l := p.add_ref(r, ifindex)
	v.r[i] = *r

	a := &v.a[i]
	*a = raw_sockaddr_ll_template
	a.Ifindex = int32(p.ifindex)
	v.m[i].msg_hdr.set(a, p.iovs)
	v.m[i].msg_len = uint32(l)
}

const (
	tx_error_unknown_interface = iota
	tx_error_interface_down
)

func (n *tx_node) init(m *net_namespace_main) {
	n.m = m
	n.Errors = []string{
		tx_error_unknown_interface: "unknown interface",
		tx_error_interface_down:    "interface is down",
	}
	m.m.v.RegisterOutputNode(n, "punt")
	n.pv_pool = make(chan *tx_packet_vector, 2*vnet.MaxVectorLen)
}

func (n *tx_node) NodeOutput(out *vnet.RefIn) {
	elog.GenEventf("unix-tx output %d", out.InLen())
	var (
		pv        *tx_packet_vector
		ns        *net_namespace
		n_unknown uint
	)
	for i := uint(0); i < out.InLen(); i++ {
		r := &out.Refs[i]
		if intf, ok := n.m.interface_by_si[r.Si]; ok {
			if intf.namespace != ns {
				if pv != nil {
					pv.tx(n, out)
				}
				ns = intf.namespace
			}
			if pv == nil {
				pv = n.get_packet_vector(out.BufferPool, ns)
			}
			pv.add_packet(n, r, intf.ifindex)
			if pv.n_packets >= packet_vector_max_len {
				pv.tx(n, out)
				pv = nil
			}
		} else {
			out.BufferPool.FreeRefs(r, 1, true)
			n_unknown++
		}
	}
	n.CountError(tx_error_unknown_interface, n_unknown)
	if pv != nil {
		pv.tx(n, out)
	}
}

func (v *tx_packet_vector) tx(n *tx_node, out *vnet.RefIn) {
	np, ns := v.n_packets, v.ns
	elog.GenEventf("unix-tx %d", np)
	atomic.AddInt32(&ns.active_count, int32(np))
	iomux.Update(ns)
	for {
		select {
		case ns.to_tx <- v:
			return
		default:
			ns.suspend_saved_out = out
			n.Suspend(out)
		}
	}
}

func (pv *tx_packet_vector) advance(n *net_namespace_tx_node, i uint) {
	np := pv.n_packets
	n_left := np - i
	pv.buffer_pool.FreeRefs(&pv.r[0], i, true)
	if n_left > 0 {
		copy(pv.a[:n_left], pv.a[i:])
		copy(pv.m[:n_left], pv.m[i:])
		copy(pv.r[:n_left], pv.r[i:])
		// For packets swap them to avoid leaking iovecs.
		for j := uint(0); j < n_left; j++ {
			pv.p[j], pv.p[i+j] = pv.p[i+j], pv.p[j]
		}
		pv.n_packets = n_left
	} else {
		n.n.put_packet_vector(pv)
		n.pv = nil
	}
}

func (n *net_namespace) WriteAvailable() bool { return n.active_count > 0 }

func (ns *net_namespace) WriteReady() (err error) {
	n := &ns.net_namespace_tx_node
	if n.pv == nil {
		n.pv = <-n.to_tx
	}
	pv := n.pv
	n_packets, errno := sendmmsg(n.Fd, 0, pv.m[:pv.n_packets])
	switch {
	case errno == syscall.EWOULDBLOCK:
		return
	case errno == syscall.EIO:
		// Signaled by tun.c in kernel and means that interface is down.
		n.n.CountError(tx_error_interface_down, 1)
	case errno != 0:
		err = fmt.Errorf("sendmmsg: %s", errno)
		return
	default:
		if n.n.m.m.verbose_packets {
			for i := 0; i < n_packets; i++ {
				r := &pv.r[i]
				intf := ns.interface_by_index[pv.p[i].ifindex]
				n.n.m.m.v.Logf("unix tx ns %s %s: %s\n", ns, intf.name, ethernet.RefString(r))
			}
		}
		elog.GenEventf("unix write-ready tx %d", n_packets)
		pv.advance(n, uint(n_packets))
		atomic.AddInt32(&ns.active_count, int32(-n_packets))
		iomux.Update(ns)
		if out := ns.suspend_saved_out; out != nil && len(n.to_tx) < cap(n.to_tx)/2 {
			ns.suspend_saved_out = nil
			n.n.Resume(out)
		}
	}
	return
}

// Namespace iomux.File is write only so ReadReady should never be called.
func (ns *net_namespace) ReadReady() (err error) {
	panic("not used")
	return
}
