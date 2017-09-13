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
	"sync/atomic"
	"syscall"
)

type tx_node struct {
	vnet.OutputNode
	m       *net_namespace_main
	pv_pool chan *tx_packet_vector
	p_pool  chan *tx_packet
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
	intf        *tuntap_interface
	ri          *vnet.RefIn
	buffer_pool *vnet.BufferPool
	a           [packet_vector_max_len]syscall.RawSockaddrLinklayer
	m           [packet_vector_max_len]mmsghdr
	p           [packet_vector_max_len]tx_packet
	r           [packet_vector_max_len]vnet.Ref
}

func (n *tx_node) get_packet_vector(p *vnet.BufferPool, intf *tuntap_interface) (v *tx_packet_vector) {
	select {
	case v = <-n.pv_pool:
	default:
		v = &tx_packet_vector{}
	}
	v.n_packets = 0
	v.buffer_pool = p
	v.intf = intf
	return
}
func (n *tx_node) put_packet_vector(v *tx_packet_vector) { n.pv_pool <- v }

func (v *tx_packet_vector) add_packet(n *tx_node, r *vnet.Ref, intf *tuntap_interface) {
	ifindex := intf.ifindex
	i := v.n_packets
	v.n_packets++

	// For tun interfaces strip ethernet header.
	// First 4 bits of ip header will indicate ip4/ip6 packet type.
	if intf.isTun {
		r.Advance(ethernet.SizeofHeader)
	}

	p := &v.p[i]
	l := p.add_ref(r, ifindex)
	v.r[i] = *r

	a := &v.a[i]
	*a = raw_sockaddr_ll_template
	a.Ifindex = int32(p.ifindex)
	if i > 0 {
		v.m[i-1].msg_hdr.Flags |= syscall.MSG_MORE
	}
	v.m[i].msg_hdr.set(a, p.iovs, 0)
	v.m[i].msg_len = uint32(l)
}

const (
	tx_error_none = iota
	tx_error_unknown_interface
	tx_error_interface_down
	tx_error_packet_too_large
	tx_error_drop
)

func (n *tx_node) init(m *net_namespace_main) {
	n.m = m
	n.Errors = []string{
		tx_error_unknown_interface: "unknown interface",
		tx_error_interface_down:    "interface is down",
		tx_error_packet_too_large:  "packet too large",
		tx_error_drop:              "error drops",
	}
	m.m.v.RegisterOutputNode(n, "punt")
	n.pv_pool = make(chan *tx_packet_vector, 2*vnet.MaxVectorLen)
}

func (n *tx_node) NodeOutput(out *vnet.RefIn) {
	elog.F1u("unix tx output %d packets", uint64(out.InLen()))
	var (
		pv      *tx_packet_vector
		pv_intf *tuntap_interface
	)
	for i := uint(0); i < out.InLen(); i++ {
		ref := &out.Refs[i]
		intf, ok := n.m.vnet_tuntap_interface_by_si[ref.Si]
		if !ok {
			out.BufferPool.FreeRefs(ref, 1, true)
			n.CountError(tx_error_unknown_interface, 1)
			continue
		}

		if intf != pv_intf {
			if pv != nil {
				pv.tx_queue(n, out)
				pv = nil
			}
			pv_intf = intf
		}
		if pv == nil {
			pv = n.get_packet_vector(out.BufferPool, intf)
		}
		pv.add_packet(n, ref, intf)
		if pv.n_packets >= packet_vector_max_len {
			pv.tx_queue(n, out)
			pv = nil
			pv_intf = nil
		}
	}
	if pv != nil {
		pv.tx_queue(n, out)
	}
}

func (v *tx_packet_vector) tx_queue(n *tx_node, ri *vnet.RefIn) {
	np, intf := v.n_packets, v.intf
	v.ri = ri
	n.AddSuspendActivity(ri, int(np))
	x := atomic.AddInt32(&intf.active_count, int32(np))
	iomux.Update(intf)
	select {
	case intf.to_tx <- v:
		if elog.Enabled() {
			e := rx_tx_elog{
				kind:      tx_elog_queue_write,
				name:      intf.elog_name,
				n_packets: uint32(np),
				active:    x,
			}
			elog.Add(&e)
		}
		return
	default:
		// Should never happen since to_tx has MaxOutstandingTxRefs entries.
		panic("tx full")
	}
}

func (pv *tx_packet_vector) advance(n *tx_node, intf *tuntap_interface, i uint) {
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
		n.put_packet_vector(pv)
		intf.pv = nil
	}
}

func (intf *tuntap_interface) WriteAvailable() bool { return intf.active_count > 0 }
func (intf *tuntap_interface) WriteReady() (err error) {
	if intf.pv == nil {
		intf.pv = <-intf.to_tx
	}
	pv := intf.pv
	ns := intf.namespace
	n := &ns.m.tx_node

	n_packets, n_drops := 0, 0
loop:
	for i := uint(0); i < pv.n_packets; i++ {
		var errno syscall.Errno
		for {
			// First try sendmsg.
			if !ns.m.tuntap_sendmsg_recvmsg_disable {
				// sendmsg/sendmmsg does yet not work on /dev/net/tun sockets.  ENOTSOCK
				_, errno = sendmsg(intf.Fd, 0, &pv.m[i].msg_hdr)
				ns.m.tuntap_sendmsg_recvmsg_disable = errno == syscall.ENOTSOCK
				if !ns.m.tuntap_sendmsg_recvmsg_disable {
					break
				}
			} else {
				// Use writev since sendmsg failed.
				_, errno = writev(intf.Fd, pv.p[i].iovs)
				break
			}
		}

		switch errno {
		case syscall.EWOULDBLOCK:
			break loop
		case syscall.EIO:
			// Signaled by tun.c in kernel and means that interface is down.
			n.CountError(tx_error_interface_down, 1)
			n_drops++
		case syscall.EMSGSIZE:
			n.CountError(tx_error_packet_too_large, 1)
			n_drops++
		default:
			if errno != 0 {
				err = fmt.Errorf("writev: %s", errno)
				n.CountError(tx_error_drop, 1)
				n_drops++
				break loop
			}
		}
		n_packets++
	}
	if n.m.m.verbose_packets {
		for i := 0; i < n_packets; i++ {
			r := &pv.r[i]
			n.m.m.v.Logf("unix tx %s: %s\n", intf.Name(), ethernet.RefString(r))
		}
	}
	// Advance to next packet in error case.
	np := n_packets + n_drops
	if np > 0 {
		pv.advance(n, intf, uint(np))
	}
	x := atomic.AddInt32(&intf.active_count, -int32(np))
	n.AddSuspendActivity(pv.ri, -int(np))

	if elog.Enabled() {
		e := rx_tx_elog{
			kind:      tx_elog_write,
			name:      intf.elog_name,
			n_packets: uint32(n_packets),
			n_drops:   uint32(n_drops),
			active:    x,
		}
		elog.Add(&e)
	}
	iomux.Update(intf)

	// Count punts and drops on this interface.
	{
		th := n.Vnet.GetIfThread(0)
		vnet.IfPunts.Add(th, intf.si, uint(n_packets))
		vnet.IfDrops.Add(th, intf.si, uint(n_drops))
	}
	return
}

const (
	tx_elog_queue_write = iota
	tx_elog_write
	rx_elog_input
	rx_elog_read
)

type rx_tx_elog_kind uint32

func (k rx_tx_elog_kind) String() string {
	t := [...]string{
		tx_elog_queue_write: "tx queue-write",
		tx_elog_write:       "tx write",
		rx_elog_input:       "rx input",
		rx_elog_read:        "rx read",
	}
	return elib.StringerHex(t[:], int(k))
}

type rx_tx_elog struct {
	name      elog.StringRef
	n_packets uint32
	n_drops   uint32
	active    int32
	kind      rx_tx_elog_kind
}

func (e *rx_tx_elog) Elog(l *elog.Log) {
	switch e.kind {
	case tx_elog_queue_write, tx_elog_write:
		if e.n_drops != 0 {
			l.Logf("unix %s %s %d packets, %d drops, active %d", e.kind, e.name, e.n_packets, e.n_drops, e.active)
		} else {
			l.Logf("unix %s %s %d packets, active %d", e.kind, e.name, e.n_packets, e.active)
		}
	case rx_elog_read:
		if e.n_drops != 0 {
			l.Logf("unix %s %s %d packets, %d drops", e.kind, e.name, e.n_packets, e.n_drops)
		} else {
			l.Logf("unix %s %s %d packets", e.kind, e.name, e.n_packets)
		}
	case rx_elog_input: // no interface name for input
		l.Logf("unix %s %d packets", e.kind, e.n_packets)
	}
}
