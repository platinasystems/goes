// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
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
	packet_vector_max_len = 8
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
		v.m[i].msg_hdr.set(&v.a[i], v.p[i].iovs, 0)
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

func (n *rx_node) init(m *Main) {
	v := m.v
	if m.RxInjectNodeName == "" {
		m.RxInjectNodeName = "inject"
	}
	n.Next = []string{
		rx_node_next_error:     "error",
		rx_node_next_inject_ip: m.RxInjectNodeName,
	}
	n.Errors = []string{
		rx_error_drop:               "drops",
		rx_error_non_vnet_interface: "not vnet interface",
		rx_error_tun_not_ip4_or_ip6: "expected 4 or 6 for ip version",
	}
	v.RegisterInputNode(n, "unix-rx")
	n.buffer_pool = vnet.DefaultBufferPool
	v.AddBufferPool(n.buffer_pool)
	n.pv_pool = make(chan *rx_packet_vector, 2*vnet.MaxVectorLen)
	n.rv_pool = make(chan *rx_ref_vector, 2*vnet.MaxVectorLen)
	n.rv_input = make(chan *rx_ref_vector, 2*vnet.MaxVectorLen)
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
	pending_refs           []rx_pending_ref
	max_buffers_per_packet uint
	next_for_inject        rx_node_next
	next_by_si             elib.Uint32Vec
}

func (n *rx_node) set_next(si vnet.Si, next rx_node_next) {
	n.next_by_si.ValidateInit(uint(si), uint32(rx_node_next_error))
	n.next_by_si[si] = uint32(next)
}

func SetRxInject(v *vnet.Vnet, inject_node_name string) {
	m := GetMain(v)
	n := &m.rx_node
	n.next_for_inject = rx_node_next(v.AddNamedNext(n, inject_node_name))
}

type rx_node_next uint32

const (
	rx_node_next_error rx_node_next = iota
	rx_node_next_inject_ip
)

const (
	rx_error_drop = iota
	rx_error_non_vnet_interface
	rx_error_tun_not_ip4_or_ip6
)

func (v *rx_ref_vector) rx_packet(ns *net_namespace, p *rx_packet, rx *rx_node, i, n_bytes_in_packet uint, ifindex uint32) {
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
	if ns.m.m.verbose_packets {
		i := ns.interface_by_index[ifindex]
		ns.m.m.v.Logf("unix rx ns %s %s: %s\n", ns.name, i.name, ethernet.RefString(&ref))
	}
	if si, ok := ns.si_by_ifindex.get(ifindex); ok {
		ref.Si = si
		n := rx_node_next(rx.next_by_si[si])
		if n == rx_node_next_inject_ip {
			if ok := add_ip_ethernet_header(&ref); !ok {
				ref.SetError(&rx.Node, rx_error_tun_not_ip4_or_ip6)
				n = rx_node_next_error
			}
		}
		if n != rx_node_next_error && rx.next_for_inject != rx_node_next_error {
			n = rx.next_for_inject
		}
		v.nexts[i] = n
	} else {
		ref.Si = vnet.SiNil
		v.nexts[i] = rx_node_next_error
	}
	v.refs[i] = ref
	v.lens[i] = uint32(n_bytes_in_packet)
	return
}

// Add empty ethernet encapsulation for injection into switch.
// Switch uses 0 destination ethernet address for punt ports to mean packet is layer 3 packet.
func add_ip_ethernet_header(r *vnet.Ref) (ok bool) {
	b := r.DataSlice()[0]
	var h ethernet.Header
	switch b >> 4 {
	case 4:
		ok = true
		h.Type = ethernet.TYPE_IP4.FromHost()
	case 6:
		ok = true
		h.Type = ethernet.TYPE_IP6.FromHost()
	}
	if ok {
		r.Advance(-ethernet.SizeofHeader)
		*(*ethernet.Header)(r.DataOffset(0)) = h
	}
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
	case 0, syscall.EAGAIN, syscall.ENETDOWN:
	default:
		err = fmt.Errorf("%s: %s", tag, errno)
	}
	return
}

func (intf *tuntap_interface) ErrorReady() (err error) {
	// Perform 0 byte read to get error from tuntap device.
	var b [0]byte
	_, err = syscall.Read(intf.Fd, b[:])
	err = nil
	return
}

func (intf *tuntap_interface) ReadReady() (err error) {
	rx := &intf.m.rx_node
	v := rx.get_packet_vector()
	n_packets := 0
	for i := range v.m {
		n, errno := readv(intf.Fd, v.p[i].iovs)
		if errno != 0 {
			err = errorForErrno("readv", errno)
			break
		}
		v.m[i].msg_len = uint32(n)
		n_packets++
	}
	if err != nil {
		rx.put_packet_vector(v)
		rx.CountError(rx_error_drop, 1)
	}
	if n_packets > 0 {
		rv := rx.get_rx_ref_vector()
		rv.n_packets = uint(n_packets)
		for i := 0; i < n_packets; i++ {
			p := &v.p[i]
			m := &v.m[i]
			rv.rx_packet(intf.namespace, p, rx, uint(i), uint(m.msg_len), intf.ifindex)
		}
		rx.rv_input <- rv
	}

	if elog.Enabled() {
		e := rx_tx_elog{
			kind:      rx_elog_ready,
			name:      intf.elog_name,
			n_packets: uint32(n_packets),
		}
		if err != nil {
			e.n_drops = 1
		}
		elog.Add(&e)
	}

	rx.AddDataActivity(int(n_packets))

	// Return packet vector for reuse.
	rx.put_packet_vector(v)

	return
}

type rx_pending_ref struct {
	ref  vnet.Ref
	len  uint32
	next rx_node_next
}

func (rx *rx_node) input_ref(ref vnet.Ref, len uint32, next rx_node_next,
	o *vnet.RefOut, n_doneʹ uint) (n_done uint, out_is_full bool) {
	n_done = n_doneʹ
	out := &o.Outs[next]
	out.BufferPool = rx.buffer_pool
	l := out.GetLen(rx.Vnet)
	if out_is_full = l == vnet.MaxVectorLen; out_is_full {
		return
	}
	if ref.Si != vnet.SiNil {
		vnet.IfRxCounter.Add(rx.GetIfThread(), ref.Si, 1, uint(len))
	}
	out.Refs[l] = ref
	out.SetLen(rx.Vnet, l+1)
	n_done++
	return
}

func (rx *rx_node) input_ref_vector(rv *rx_ref_vector, o *vnet.RefOut, n_doneʹ uint) (n_done uint, out_is_full bool) {
	n_done = n_doneʹ
	var i uint

	// First process pending packets.
	np := uint(len(rx.pending_refs))
	for i = 0; i < np; i++ {
		pr := &rx.pending_refs[i]
		n_done, out_is_full = rx.input_ref(pr.ref, pr.len, pr.next, o, n_done)
		if out_is_full {
			// Re-copy pending vector.
			copy(rx.pending_refs[:], rx.pending_refs[i:])
			break
		}
	}

	// Remove pending refs and reuse pending vector.
	if np > 0 {
		rx.pending_refs = rx.pending_refs[:np-i]
	}

	// Loop through given input packets.
	for i = 0; i < rv.n_packets; i++ {
		r, l, n := rv.refs[i], rv.lens[i], rv.nexts[i]
		if !out_is_full {
			n_done, out_is_full = rx.input_ref(r, l, n, o, n_done)
		}
		// A next node is or was full: add to pending vector.
		if out_is_full {
			r := rx_pending_ref{ref: r, len: l, next: n}
			rx.pending_refs = append(rx.pending_refs, r)
		}
	}
	rx.put_rx_ref_vector(rv)
	return
}

func (rx *rx_node) NodeInput(out *vnet.RefOut) {
	n_packets, out_is_full := uint(0), false
loop:
	for !out_is_full {
		select {
		case rv := <-rx.rv_input:
			n_packets, out_is_full = rx.input_ref_vector(rv, out, n_packets)
		default:
			break loop
		}
	}

	rx.AddDataActivity(-int(n_packets))

	if elog.Enabled() {
		e := rx_tx_elog{
			kind:      rx_elog_input,
			n_packets: uint32(n_packets),
		}
		elog.Add(&e)
	}
}
