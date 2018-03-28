// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/iomux"
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

func (p *rx_packet) set_iovs(rx *rx_node, n uint) {
	for i := uint(0); i < n; i++ {
		p.iovs[i].Base = (*byte)(p.refs[i].Data())
		p.iovs[i].Len = uint64(rx.buffer_pool.Size)
	}
}

func (p *rx_packet) alloc_refs(rx *rx_node, n uint) {
	rx.buffer_pool.AllocRefs(p.refs[:n])
	p.set_iovs(rx, n)
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
	rx_packet_vector_max_len = 16
	tx_packet_vector_max_len = vnet.MaxVectorLen
	max_rx_packet_size       = 16 << 10
)

// Maximum sized packet vector.
type rx_packet_vector struct {
	i             uint
	a             [rx_packet_vector_max_len]syscall.RawSockaddrLinklayer
	m             [rx_packet_vector_max_len]mmsghdr
	p             [rx_packet_vector_max_len]rx_packet
	nrefs         [rx_packet_vector_max_len]uint32
	n_refill_refs uint
	alloc_refs    vnet.RefVec
}

func (v *rx_packet_vector) refill_refs(rx *rx_node, n_packets uint) {
	n := v.n_refill_refs
	if n == 0 {
		return
	}
	v.n_refill_refs = 0
	v.alloc_refs.ValidateLen(n)
	rx.buffer_pool.AllocRefs(v.alloc_refs[:n])
	j := uint(0)
	for i := uint(0); i < n_packets; i++ {
		nr := uint(v.nrefs[i])
		p := &v.p[i]
		copy(p.refs[:nr], v.alloc_refs[j:])
		p.set_iovs(rx, nr)
		j += nr
	}
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

func (n *rx_node) get_rx_ref_vector(intf *tuntap_interface) (v *rx_ref_vector) {
	select {
	case v = <-n.rv_pool:
	default:
		v = &rx_ref_vector{}
	}
	v.intf = intf
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
	n.pv_pool = make(chan *rx_packet_vector, 8*vnet.MaxVectorLen)
	n.rv_pool = make(chan *rx_ref_vector, 8*vnet.MaxVectorLen)
	n.rv_input = make(chan *rx_ref_vector, 8*vnet.MaxVectorLen)
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
	pending_intfs          []*tuntap_interface
	max_buffers_per_packet uint
	next_for_inject        rx_node_next
	next_by_si             elib.Uint32Vec
}

type tuntap_interface_rx_node struct {
	// Interface is on pending_intfs slice.
	is_pending bool
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

func (rv *rx_ref_vector) rx_packet(rx *rx_node, ns *net_namespace, pv *rx_packet_vector, pvi, rvi, n_bytes_in_packet uint, ifindex uint32) (n_refs uint) {
	//avoid corner case panic when interface is down and removed but a transaction is in progress
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("rx.go: rx_packet recover(), pvi=%d, ifindex=%d, %v\n", pvi, ifindex, x)
		}
	}()
	//
	size := rx.buffer_pool.Size
	n_left := n_bytes_in_packet
	p := &pv.p[pvi]
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
	pv.nrefs[pvi] = uint32(n_refs)
	pv.n_refill_refs += n_refs
	ref := p.chain.Done()
	ref.SetError(&rx.Node, rx_error_non_vnet_interface)
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
		rv.nexts[rvi] = n
	} else {
		ref.Si = vnet.SiNil
		rv.nexts[rvi] = rx_node_next_error
	}
	rv.refs[rvi] = ref
	rv.nrefs[rvi] = uint32(n_refs)
	rv.lens[rvi] = uint32(n_bytes_in_packet)
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
	n_refs    uint
	xi        uint
	intf      *tuntap_interface
	refs      [vnet.MaxVectorLen]vnet.Ref
	nrefs     [vnet.MaxVectorLen]uint32
	lens      [vnet.MaxVectorLen]uint32
	nexts     [vnet.MaxVectorLen]rx_node_next
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

func (intf *tuntap_interface) ReadAvailable() bool {
	rx := &intf.m.rx_node
	return rx.ActiveCount() < vnet.MaxOutstandingTxRefs
}

func (intf *tuntap_interface) ReadReady() (err error) {
	var n_packets, n_refs, n_rv uint
	rx := &intf.m.rx_node
	eagain := false

	if rx.IsSuspended() {
		return
	}

	pv := rx.get_packet_vector()
	rv := rx.get_rx_ref_vector(intf)

	is_suspended := false
	n_refs_this_ref_vector := uint(0)
	const max_refs = vnet.MaxOutstandingTxRefs
	for !eagain && n_refs < max_refs && !is_suspended {
		n_packets_this_packet_vector := uint(0)
		for pvi := range pv.m {
			n, errno := readv(intf.Fd, pv.p[pvi].iovs)
			if errno != 0 {
				eagain = errno == syscall.EAGAIN
				err = errorForErrno("readv", errno)
				break
			}
			pv.m[pvi].msg_len = uint32(n)
			n_refs_this_packet := rv.rx_packet(rx, intf.namespace, pv, uint(pvi), n_rv, uint(n), intf.ifindex)
			n_refs_this_ref_vector += n_refs_this_packet
			n_rv++
			if n_rv >= uint(len(rv.refs)) {
				if elog.Enabled() {
					e := rx_tx_elog{
						kind:      rx_elog_read,
						name:      intf.elog_name,
						n_packets: uint32(n_rv),
						n_refs:    uint32(n_refs_this_ref_vector),
					}
					elog.Add(&e)
				}
				rv.n_packets = n_rv
				rv.n_refs = n_refs_this_ref_vector
				rx.AddDataActivity(int(n_refs_this_ref_vector))
				rx.rv_input <- rv
				n_rv = 0
				n_refs_this_ref_vector = 0
				rv = rx.get_rx_ref_vector(intf)
			}
			n_packets++
			n_refs += n_refs_this_packet
			n_packets_this_packet_vector++
			if is_suspended = rx.IsSuspended(); is_suspended {
				break
			}
		}
		pv.refill_refs(rx, n_packets_this_packet_vector)
		if err != nil {
			rx.CountError(rx_error_drop, 1)
			break
		}
	}
	if n_rv > 0 {
		if elog.Enabled() {
			e := rx_tx_elog{
				kind:        rx_elog_read,
				name:        intf.elog_name,
				n_packets:   uint32(n_rv),
				n_refs:      uint32(n_refs_this_ref_vector),
				would_block: eagain,
			}
			if err != nil {
				e.n_drops = 1
			}
			elog.Add(&e)
		}
		rv.n_packets = n_rv
		rv.n_refs = n_refs_this_ref_vector
		rx.AddDataActivity(int(n_refs_this_ref_vector))
		rx.rv_input <- rv
	} else {
		// Return unused ref vector.
		rx.put_rx_ref_vector(rv)
	}
	// Return packet vector for reuse.
	rx.put_packet_vector(pv)
	if intf.Fd != -1 {
		iomux.Update(intf)
	}
	return
}

type rx_pending_ref struct {
	intf  *tuntap_interface
	ref   vnet.Ref
	n_ref uint32
	len   uint32
	next  rx_node_next
}

func (rx *rx_node) input_ref(intf *tuntap_interface, ref vnet.Ref, len uint32, next rx_node_next, o *vnet.RefOut) (out_is_full bool) {
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
	if !intf.is_pending {
		intf.is_pending = true
		rx.pending_intfs = append(rx.pending_intfs, intf)
	}
	return
}

func (rx *rx_node) input_ref_vector(rv *rx_ref_vector, o *vnet.RefOut, n_packets0, n_refs0 uint) (n_packets uint, n_refs uint, out_is_full bool) {
	n_packets, n_refs = n_packets0, n_refs0
	var i uint

	// First process pending packets.
	np := uint(len(rx.pending_refs))
	for i = 0; i < np; i++ {
		pr := &rx.pending_refs[i]
		out_is_full = rx.input_ref(pr.intf, pr.ref, pr.len, pr.next, o)
		if out_is_full {
			// Re-copy pending vector.
			copy(rx.pending_refs[:], rx.pending_refs[i:])
			break
		}
		n_packets++
		n_refs += uint(pr.n_ref)
	}

	// Remove pending refs and reuse pending vector.
	if np > 0 {
		rx.pending_refs = rx.pending_refs[:np-i]
	}

	if rv != nil {
		// Loop through given input packets.
		for i = 0; i < rv.n_packets; i++ {
			r, nr, l, n := rv.refs[i], rv.nrefs[i], rv.lens[i], rv.nexts[i]
			if !out_is_full {
				out_is_full = rx.input_ref(rv.intf, r, l, n, o)
			}
			// A next node is or was full: add to pending vector.
			if out_is_full {
				r := rx_pending_ref{intf: rv.intf, ref: r, n_ref: nr, len: l, next: n}
				rx.pending_refs = append(rx.pending_refs, r)
			} else {
				n_packets++
				n_refs += uint(nr)
			}
		}
		rx.put_rx_ref_vector(rv)
	}
	return
}

func (rx *rx_node) NodeInput(out *vnet.RefOut) {
	n_packets, n_refs, out_is_full := uint(0), uint(0), false
	if rx.pending_intfs != nil {
		rx.pending_intfs = rx.pending_intfs[:0]
	}
loop:
	for !out_is_full {
		select {
		case rv := <-rx.rv_input:
			n_packets, n_refs, out_is_full = rx.input_ref_vector(rv, out, n_packets, n_refs)
		default:
			if len(rx.pending_refs) > 0 {
				n_packets, n_refs, out_is_full = rx.input_ref_vector(nil, out, n_packets, n_refs)
			}
			break loop
		}
	}

	if elog.Enabled() {
		e := rx_tx_elog{
			kind:      rx_elog_input,
			n_packets: uint32(n_packets),
			n_refs:    uint32(n_refs),
		}
		elog.Add(&e)
	}

	rx.AddDataActivity(-int(n_refs))
	for _, intf := range rx.pending_intfs {
		intf.is_pending = false
		if intf.Fd != -1 {
			iomux.Update(intf)
		}
	}
}
