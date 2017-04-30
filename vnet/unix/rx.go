// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

var raw_sockaddr_ll_template = syscall.RawSockaddrLinklayer{
	Family: syscall.AF_PACKET,
}

type rx_packet struct {
	iovs  iovecVec
	refs  vnet.RefVec
	chain vnet.RefChain
}

func (p *rx_packet) alloc_refs(rx *rx_node, n uint) {
	rx.buffer_pool.AllocRefs(p.refs[:n])
	for i := uint(0); i < n; i++ {
		p.iovs[i].Base = (*byte)(p.refs[i].Data())
		p.iovs[i].Len = uint64(rx.buffer_pool.Size)
	}
}

func (p *rx_packet) rx_init(rx *rx_node) {
	n := rx.max_buffers_per_packet
	p.iovs.Validate(n - 1)
	p.refs.Validate(n - 1)
	p.iovs = p.iovs[:n]
	p.refs = p.refs[:n]
	p.alloc_refs(rx, n)
}

func (p *rx_packet) rx_free(rx *rx_node) { rx.buffer_pool.FreeRefs(&p.refs[0], p.refs.Len(), false) }

const (
	packet_vector_max_len = 64
	max_rx_packet_size    = 16 << 10
)

// Maximum sized packet vector.
type rx_packet_vector struct {
	i uint
	a [packet_vector_max_len]syscall.RawSockaddrLinklayer
	m [packet_vector_max_len]mmsghdr
	p [packet_vector_max_len]rx_packet
}

func (n *rx_node) new_packet_vector() (v *rx_packet_vector) {
	v = &rx_packet_vector{}
	for i := range v.p {
		v.p[i].rx_init(n)
		v.a[i] = raw_sockaddr_ll_template
		v.m[i].msg_hdr.set(&v.a[i], v.p[i].iovs)
	}
	return
}

func (n *rx_node) get_packet_vector() (v *rx_packet_vector) {
	select {
	case v = <-n.pv_pool:
	default:
		v = n.new_packet_vector()
	}
	return
}

func (n *rx_node) put_packet_vector(v *rx_packet_vector) { n.pv_pool <- v }

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
	n.pv_pool = make(chan *rx_packet_vector, 256)
	n.rv_pool = make(chan *rx_ref_vector, 256)
	n.rv_input = make(chan *rx_ref_vector, 256)
	n.max_buffers_per_packet = max_rx_packet_size / n.buffer_pool.Size
	if max_rx_packet_size%n.buffer_pool.Size != 0 {
		n.max_buffers_per_packet++
	}
}

type rx_node struct {
	vnet.InputNode
	buffer_pool            *vnet.BufferPool
	pv_pool                chan *rx_packet_vector
	rv_pool                chan *rx_ref_vector
	rv_input               chan *rx_ref_vector
	rv_pending             *rx_ref_vector
	max_buffers_per_packet uint
	active_lock            sync.Mutex
	active_count           int32
	next_by_si             elib.Uint32Vec
}

func (n *rx_node) set_next(si vnet.Si, next rx_node_next) {
	n.next_by_si.ValidateInit(uint(si), uint32(rx_node_next_error))
	n.next_by_si[si] = uint32(next)
}

type rx_node_next uint32

const (
	rx_node_next_error rx_node_next = iota
)

const (
	rx_error_drop = iota
	rx_error_non_vnet_interface
)

func (v *rx_ref_vector) rx_packet(ns *net_namespace, p *rx_packet, rx *rx_node, i, n_bytes_in_packet, ifindex uint) {
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
	if ns.m.m.verbosePackets {
		i := ns.interface_by_index[uint32(ifindex)]
		ns.m.m.v.Logf("unix rx ns %s %s: %s\n", ns.name, i.name, ethernet.RefString(&ref))
	}
	if si, ok := ns.si_by_ifindex[uint32(ifindex)]; ok {
		ref.Si = si
		v.nexts[i] = rx_node_next(rx.next_by_si[si])
		vnet.IfRxCounter.Add(rx.Vnet.GetIfThread(0), si, 1, n_bytes_in_packet)
	} else {
		ref.Si = vnet.SiNil
		v.nexts[i] = rx_node_next_error
	}
	v.refs[i] = ref
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

func errorForErrno(tag string, errno syscall.Errno) (err error) {
	// Ignore "network is down" errors.  Just silently drop packet.
	// These happen when interface is IFF_RUNNING (e.g. link up) but not yet IFF_UP (admin up).
	switch errno {
	case 0, syscall.ENETDOWN:
	default:
		err = fmt.Errorf("%s: %s", tag, errno)
	}
	return
}

// Write side is not used.  Per-namespace raw socket is used for tx path.
func (intf *tuntap_interface) WriteReady() (err error) { panic("shoudn't be called") }
func (intf *tuntap_interface) WriteAvailable() bool    { return false }

func (intf *tuntap_interface) ErrorReady() (err error) {
	var e int
	if e, err = syscall.GetsockoptInt(intf.Fd, syscall.SOL_SOCKET, syscall.SO_ERROR); err == nil {
		err = errorForErrno("error ready", syscall.Errno(e))
	}
	if err != nil {
		panic(err)
	}
	return
}

func (intf *tuntap_interface) ReadReady() (err error) {
	rx := &intf.m.rx_node
	v := rx.get_packet_vector()
	n_packets, errno := recvmmsg(intf.Fd, syscall.MSG_WAITFORONE, v.m[:])
	if errno != 0 {
		err = errorForErrno("recvmmsg", errno)
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
		rv.rx_packet(intf.namespace, p, rx, uint(i), uint(m.msg_len), uint(m.msg_hdr.ifindex()))
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
	if n_left < packet_vector_max_len {
		copy(p.refs[:n_left], rv.refs[i:])
		copy(p.lens[:n_left], rv.lens[i:])
		copy(p.nexts[:n_left], rv.nexts[i:])
		p.n_packets = n_left
	}
	rx.rv_pending = p
}

func (rx *rx_node) input_ref_vector(rv *rx_ref_vector, o *vnet.RefOut, n_doneʹ uint) (n_done uint, out_is_full bool) {
	n_done = n_doneʹ
	var i uint
	for i = 0; i < rv.n_packets; i++ {
		out := &o.Outs[rv.nexts[i]]
		out.BufferPool = rx.buffer_pool
		l := out.GetLen(rx.Vnet)
		if out_is_full = l == vnet.MaxVectorLen; out_is_full {
			rx.copy_pending(rv, i)
			break
		}
		r := &rv.refs[i]
		out.Refs[l] = *r
		out.SetLen(rx.Vnet, l+1)
		n_done++
	}
	if i >= rv.n_packets {
		if rv == rx.rv_pending { // clear pending
			rx.rv_pending = nil
		}
		rx.put_rx_ref_vector(rv)
	}
	return
}

func (rx *rx_node) NodeInput(out *vnet.RefOut) {
	n_done, out_is_full := uint(0), false
	if rx.rv_pending != nil {
		n_done, out_is_full = rx.input_ref_vector(rx.rv_pending, out, n_done)
	}
loop:
	for !out_is_full {
		select {
		case rv := <-rx.rv_input:
			n_done, out_is_full = rx.input_ref_vector(rv, out, n_done)
		default:
			break loop
		}
	}
	rx.active_lock.Lock()
	rx.Activate(atomic.AddInt32(&rx.active_count, -int32(n_done)) > 0)
	rx.active_lock.Unlock()
}
