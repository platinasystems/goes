// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/iomux"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

type nodeMain struct {
	v            *vnet.Vnet
	rxPacketPool chan *packet
	rxRefsFree   chan []rxRef
	txPacketPool chan *packet
	puntNode     puntNode
}

func (nm *nodeMain) Init(m *Main) {
	nm.rxPacketPool = make(chan *packet, 64)
	nm.txPacketPool = make(chan *packet, 64)
	nm.rxRefsFree = make(chan []rxRef, 64)
	nm.puntNode.Errors = []string{
		puntErrorNonUnix:       "non-unix interface",
		puntErrorInterfaceDown: "interface is down",
	}
	nm.puntNode.Next = []string{
		puntNextError: "error",
	}
	m.v.RegisterInOutNode(&nm.puntNode, "punt")
}

type node struct {
	ethernet.Interface
	vnet.InterfaceNode
	i             *tuntap_interface
	rxRefsLock    sync.Mutex
	rxRefs        []rxRef
	rxPending     []rxRef
	rxDrops       int32
	rxActiveLock  sync.Mutex
	rxActiveCount int32
	txRefIns      chan txRefIn
	txRefIn       txRefIn
	txAvailable   int32
	txIovecs      iovecVec
}

type rxNext int

const (
	rxNextTx rxNext = iota
)

func (intf *tuntap_interface) interfaceNodeInit(m *Main) {
	ifName := intf.Name()
	vnetName := ifName + "-unix"
	n := &intf.node
	n.i = intf
	n.txRefIns = make(chan txRefIn, 64)
	n.Next = []string{
		rxNextTx: ifName,
	}
	m.v.RegisterHwInterface(n, vnetName)
	m.v.RegisterInterfaceNode(n, n.Hi(), vnetName)
	ni := m.v.AddNamedNext(&m.puntNode, vnetName)
	m.puntNode.setNext(intf.si, ni)

	// Use /dev/net/tun file descriptor for input/output.
	intf.Fd = intf.namespace.dev_net_tun_fd
	iomux.Add(intf)
}

func (n *node) DriverName() string                                          { return "tuntap" }
func (n *node) GetHwInterfaceCounterNames() (nm vnet.InterfaceCounterNames) { return }
func (n *node) GetHwInterfaceCounterValues(t *vnet.InterfaceThread)         {}
func (n *node) ValidateSpeed(speed vnet.Bandwidth) (err error)              { return }

type rxRef struct {
	ref vnet.Ref
	len uint
}

func (h *msghdr) set(a *syscall.RawSockaddrLinklayer, iovs iovecVec) {
	h.Name = (*byte)(unsafe.Pointer(a))
	h.Namelen = syscall.SizeofSockaddrLinklayer
	h.Iov = (*syscall.Iovec)(&iovs[0])
	h.Iovlen = uint64(len(iovs))
}

// Maximum sized packet vector.
type packet_vector struct {
	a  [vnet.MaxVectorLen]syscall.RawSockaddrLinklayer
	mm [vnet.MaxVectorLen]mmsghdr
}

var raw_sockaddr_ll_template = syscall.RawSockaddrLinklayer{
	Family: syscall.AF_PACKET,
}

func (v *packet_vector) tx_add(p *packet, ifindex uint32, i uint) {
	a := &v.a[i]
	*a = raw_sockaddr_ll_template
	a.Ifindex = int32(ifindex)
	v.mm[i].msg_hdr.set(a, p.iovs)
}

type packet struct {
	iovs  iovecVec
	refs  vnet.RefVec
	chain vnet.RefChain
}

func (p *packet) allocRefs(m *Main, n uint) {
	m.bufferPool.AllocRefs(p.refs[:n])
	for i := uint(0); i < n; i++ {
		p.iovs[i].Base = (*byte)(p.refs[i].Data())
		p.iovs[i].Len = uint64(m.bufferPool.Size)
	}
}

func (p *packet) initForRx(m *Main, intf *tuntap_interface) {
	n := intf.mtuBuffers
	p.iovs.Validate(n - 1)
	p.refs.Validate(n - 1)
	p.iovs = p.iovs[:n]
	p.refs = p.refs[:n]
	p.allocRefs(m, n)
}

func (p *packet) free(m *Main) {
	m.bufferPool.FreeRefs(&p.refs[0], p.refs.Len(), false)
}

func (m *Main) getRxPacket(intf *tuntap_interface) (p *packet) {
	select {
	case p = <-m.rxPacketPool:
	default:
		p = &packet{}
		p.initForRx(m, intf)
	}
	return
}

func (m *Main) putRxPacket(p *packet) {
	select {
	case m.rxPacketPool <- p:
	default:
		p.free(m)
	}
}

func (m *Main) rxRefsGet() (r []rxRef) {
	select {
	case r = <-m.rxRefsFree:
		r = r[:0]
	default:
	}
	return
}

func (m *Main) rxRefsPut(r []rxRef) {
	select {
	case m.rxRefsFree <- r:
	default:
	}
}

func (n *node) InterfaceInput(o *vnet.RefOut) {
	m := n.i.m
	toTx := &o.Outs[rxNextTx]
	toTx.BufferPool = m.bufferPool

	n.rxRefsLock.Lock()
	refs := n.rxRefs
	n.rxRefs = m.rxRefsGet()
	n.rxRefsLock.Unlock()
	if n.rxPending == nil {
		n.rxPending = refs
	} else {
		n.rxPending = append(n.rxPending, refs...)
		m.rxRefsPut(refs)
	}

	nDrops := n.rxDrops
	atomic.AddInt32(&n.rxDrops, -nDrops)

	nPending := uint(len(n.rxPending))
	nPackets, nBytes := nPending, uint(0)
	if nPackets >= vnet.MaxVectorLen {
		nPackets = vnet.MaxVectorLen
	}

	for i := uint(0); i < nPackets; i++ {
		r := &n.rxPending[i]
		nBytes += r.len
		toTx.Refs[i] = r.ref
		if m.verbosePackets {
			m.v.Logf("unix rx %d: %x\n", r.len, r.ref.DataSlice())
		}
	}
	if nPackets < nPending {
		copy(n.rxPending[0:], n.rxPending[nPackets:])
	}
	n.rxPending = n.rxPending[:nPending-nPackets]

	t := n.GetIfThread()
	vnet.IfRxCounter.Add(t, n.Si(), nPackets, nBytes)
	vnet.IfDrops.Add(t, n.Si(), uint(nDrops))
	toTx.SetLen(m.v, nPackets)
	n.rxActiveLock.Lock()
	n.Activate(atomic.AddInt32(&n.rxActiveCount, -int32(nPackets)) > 0)
	n.rxActiveLock.Unlock()
}

func (intf *tuntap_interface) ReadReady() (err error) {
	m, n := intf.m, &intf.node
	p := m.getRxPacket(intf)
	var (
		nRead int
		errno syscall.Errno
	)
	nRead, errno = readv(intf.Fd, p.iovs)
	if errno != 0 {
		err = errorForErrno("readv", errno)
		m.putRxPacket(p)
		atomic.AddInt32(&n.rxDrops, 1)
		return
	}
	size := m.bufferPool.Size
	nLeft := uint(nRead)
	var nRefs uint
	for nRefs = 0; nLeft > 0; nRefs++ {
		l := size
		if nLeft < l {
			l = nLeft
		}
		r := &p.refs[nRefs]
		r.SetDataLen(l)
		if r.NextValidFlag() != 0 {
			panic("next")
		}
		p.chain.Append(r)
		nLeft -= l
	}

	// Send packet to input node.
	var r rxRef
	r.len = p.chain.Len()
	r.ref = p.chain.Done()
	// use xxx-unix interface as receive interface.
	r.ref.Si = n.Si()
	n.rxActiveLock.Lock()
	atomic.AddInt32(&n.rxActiveCount, 1)
	n.Activate(true)
	n.rxActiveLock.Unlock()
	n.rxRefsLock.Lock()
	n.rxRefs = append(n.rxRefs, r)
	n.rxRefsLock.Unlock()

	// Refill packet with new buffers & return for re-use.
	p.allocRefs(m, nRefs)
	m.putRxPacket(p)
	return
}

type txRefIn struct {
	in *vnet.TxRefVecIn
	i  uint
}

func (n *node) InterfaceOutput(i *vnet.TxRefVecIn) {
	intf := n.i
	n.txRefIns <- txRefIn{in: i}
	atomic.AddInt32(&n.txAvailable, 1)
	iomux.Update(intf)
}

func (intf *tuntap_interface) WriteAvailable() bool { return intf.node.txAvailable > 0 }

func (intf *tuntap_interface) WriteReady() (err error) {
	n := &intf.node
	ri := &n.txRefIn
	for {
		l := uint(0)
		if ri.in != nil {
			l = ri.in.Refs.Len()
		}
		if ri.i >= l {
			if ri.in != nil {
				intf.m.Vnet.FreeTxRefIn(ri.in)
				ri.in = nil
				atomic.AddInt32(&n.txAvailable, -1)
				iomux.Update(intf)
			}
			select {
			case *ri = <-n.txRefIns:
				ri.i = 0
			default:
				return
			}
		}

		// Convert vnet buffer references for a single packet into iovecs for writing to kernel.
		nIovecs, nWriteLeft := uint(0), uint(0)
		for i := ri.i; i < ri.in.Refs.Len(); i++ {
			n.txIovecs.Validate(nIovecs)
			r := &ri.in.Refs[i]
			n.txIovecs[nIovecs] = iovec{
				Base: (*byte)(r.Data()),
				Len:  uint64(r.DataLen()),
			}
			nWriteLeft += r.DataLen()
			nIovecs++
			if !r.NextIsValid() {
				break
			}
		}

		// Inject packet into kernel tun/tap devices.
		if nIovecs > 0 {
			nWrite, errno := writev(intf.Fd, n.txIovecs[:nIovecs])
			switch {
			case errno == syscall.EWOULDBLOCK:
				return
			case errno == syscall.EIO:
				// Signaled by tun.c in kernel and means that interface is down.
				intf.m.puntNode.CountError(puntErrorInterfaceDown, 1)
			case errno != 0:
				err = fmt.Errorf("writev: %s", errno)
				return
			default:
				if uint(nWrite) != nWriteLeft {
					panic("partial packet write")
				}
				if intf.m.verbosePackets {
					intf.m.v.Logf("unix tx %d: %x\n", nWrite, ri.in.Refs[ri.i].DataSlice())
				}
			}
			ri.i += nIovecs
		}
	}

	return
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

const (
	puntErrorNonUnix uint = iota
	puntErrorInterfaceDown
)
const (
	puntNextError uint = iota
)

type puntNode struct {
	vnet.InOutNode
	nextBySi elib.Uint32Vec
}

func (n *puntNode) setNext(si vnet.Si, next uint) {
	n.nextBySi.ValidateInit(uint(si), uint32(next))
	n.nextBySi[si] = uint32(next)
}

func (n *puntNode) NodeInput(in *vnet.RefIn, o *vnet.RefOut) {
	for i := uint(0); i < in.Len(); i++ {
		r := &in.Refs[i]
		x := n.nextBySi[r.Si]
		n.SetError(r, puntErrorNonUnix)
		o.Outs[x].BufferPool = in.BufferPool
		no := o.Outs[x].AddLen(n.Vnet)
		o.Outs[x].Refs[no] = *r
	}
}
