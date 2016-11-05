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
	"sync/atomic"
	"syscall"
	"unsafe"
)

type nodeMain struct {
	v            *vnet.Vnet
	rxPacketPool chan *packet
	txPacketPool chan *packet
	puntNode     puntNode
}

func (nm *nodeMain) Init(m *Main) {
	nm.rxPacketPool = make(chan *packet, 64)
	nm.txPacketPool = make(chan *packet, 64)
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
	i           *Interface
	rxRefs      chan rxRef
	txRefIns    chan txRefIn
	txRefIn     txRefIn
	txAvailable int32
	txIovecs    iovecVec
}

type rxNext int

const (
	rxNextTx rxNext = iota
)

func (intf *Interface) interfaceNodeInit(m *Main) {
	ifName := intf.Name()
	vnetName := ifName + "-unix"
	n := &intf.node
	n.i = intf
	n.rxRefs = make(chan rxRef, vnet.MaxVectorLen)
	n.txRefIns = make(chan txRefIn, 64)
	m.v.RegisterHwInterface(n, vnetName)
	n.Next = []string{
		rxNextTx: ifName,
	}
	m.v.RegisterInterfaceNode(n, n.Hi(), vnetName)
	ni := m.v.AddNamedNext(&m.puntNode, vnetName)
	m.puntNode.setNext(intf.si, ni)

	// Use /dev/net/tun file descriptor for input/output.
	intf.Fd = intf.dev_net_tun_fd
	iomux.Add(intf)
}

func (n *node) DriverName() string                                          { return "tuntap" }
func (n *node) GetHwInterfaceCounterNames() (nm vnet.InterfaceCounterNames) { return }
func (n *node) GetHwInterfaceCounterValues(t *vnet.InterfaceThread)         {}
func (n *node) ValidateSpeed(speed vnet.Bandwidth) (err error)              { return }

type iovec syscall.Iovec

//go:generate gentemplate -d Package=unix -id iovec -d VecType=iovecVec -d Type=iovec github.com/platinasystems/go/elib/vec.tmpl

func rwv(fd int, iov []iovec, isWrite bool) (n int, e syscall.Errno) {
	sc := syscall.SYS_READV
	if isWrite {
		sc = syscall.SYS_WRITEV
	}
	r0, _, e := syscall.Syscall(uintptr(sc), uintptr(fd), uintptr(unsafe.Pointer(&iov[0])), uintptr(len(iov)))
	n = int(r0)
	return
}

func readv(fd int, iov []iovec) (int, syscall.Errno)  { return rwv(fd, iov, false) }
func writev(fd int, iov []iovec) (int, syscall.Errno) { return rwv(fd, iov, true) }

type rxRef struct {
	ref vnet.Ref
	len uint
}

type packet struct {
	iovs  iovecVec
	chain vnet.RefChain
	refs  vnet.RefVec
}

func (p *packet) allocRefs(m *Main, n uint) {
	m.bufferPool.AllocRefs(p.refs[:n])
	for i := uint(0); i < n; i++ {
		p.iovs[i].Base = (*byte)(p.refs[i].Data())
		p.iovs[i].Len = uint64(m.bufferPool.Size)
	}
}

func (p *packet) initForRx(m *Main, intf *Interface) {
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

func (m *Main) getRxPacket(intf *Interface) (p *packet) {
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

func (n *node) InterfaceInput(o *vnet.RefOut) {
	m := n.i.m
	toTx := &o.Outs[rxNextTx]
	toTx.BufferPool = m.bufferPool
	t := n.GetIfThread()
	nPackets, nBytes, nDrops := uint(0), uint(0), uint(0)

	done := false
	for !done {
		select {
		case r := <-n.rxRefs:
			if r.len == ^uint(0) {
				nDrops++
			} else {
				nBytes += r.len
				r.ref.Si = n.Si() // use xxx-unix interface as receive interface.
				toTx.Refs[nPackets] = r.ref
				nPackets++
				if m.verbosePackets {
					m.v.Logf("unix rx %d: %x\n", r.len, r.ref.DataSlice())
				}
				done = nPackets >= uint(len(toTx.Refs))
			}
		default:
			done = true
		}
	}

	vnet.IfRxCounter.Add(t, n.Si(), nPackets, nBytes)
	vnet.IfDrops.Add(t, n.Si(), nDrops)
	toTx.SetLen(m.v, nPackets)
	n.Activate(false)
}

func (intf *Interface) ReadReady() (err error) {
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
		n.rxRefs <- rxRef{len: ^uint(0)}
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
	n.rxRefs <- r
	n.Activate(true)

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

func (intf *Interface) WriteAvailable() (ok bool) {
	n := &intf.node
	ri := &n.txRefIn
	return n.txAvailable > 0 || ri.in != nil && ri.i < ri.in.Len()
}

func (intf *Interface) WriteReady() (err error) {
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
			}
			select {
			case *ri = <-n.txRefIns:
				atomic.AddInt32(&n.txAvailable, -1)
				ri.i = 0
			default:
				iomux.Update(intf)
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

func (intf *Interface) ErrorReady() (err error) {
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
