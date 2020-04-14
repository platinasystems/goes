// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package provides a sideband control interface to an XETH driver.
// Usage,
//	var wg sync.WaitGroup
//	defer wg.Wait()
//	stopch := make(chan struct{})
//	defer close(stopch)
//	task, err := xeth.Start(&wg, stopch)
//	if err {
//		panic(err)
//	}
//	task.DumpIfInfo()
//	for buf := range task.RxCh {
//		if xeth.Class(buf) == xeth.ClassBreak {
//			break
//		}
//		msg := xeth.Parse(buf)
//		...
//		xeth.Pool(msg)
//	}
//	...
//	task.DumpFib()
//	for buf := range task.RxCh {
//		if xeth.Class(buf) == xeth.ClassBreak {
//			break
//		}
//		msg := xeth.Parse(buf)
//		...
//		xeth.Pool(msg)
//	}
//	...
//	go func() {
//		wg.Add(1)
//		defer wg.Done()
//		for buf := range task.RxCh {
//			msg := xeth.Parse(buf)
//			...
//			xeth.Pool(msg)
//		}
//	}()
//	...
package xeth

import (
	"io"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/platinasystems/goes/external/xeth/internal"
)

//go:generate sh -c "go tool cgo -godefs godefs.go > godefed.go"
//go:generate sh -c "go tool cgo -godefs internal/godefs.go > internal/godefed.go"

const unixpacket = "unixpacket"

const (
	ClassUnknown = iota
	ClassBreak
	ClassInterface
	ClassAddress
	ClassFib
	ClassNeighbor
)

const (
	VLAN_PRIO_MASK   = 0xe000
	VLAN_PRIO_SHIFT  = 13
	VLAN_CFI_MASK    = 0x1000
	VLAN_TAG_PRESENT = VLAN_CFI_MASK
)

type pooler interface {
	Pool()
}

type Break struct{}

type Buffer interface{ buffer }

var (
	Cloned  Counter // cloned received messages
	Parsed  Counter // messages parsed by user
	Dropped Counter // messages that overflowed transmit channel
	Sent    Counter // messages and exception frames sent to driver
	Unknown Counter // LinkOf xid w/o IfInfo
)

type Task struct {
	RxCh <-chan Buffer // cloned msgs received from driver

	wg *sync.WaitGroup

	stopch <-chan struct{}

	sock *net.UnixConn

	loch chan<- buffer // low priority, leaky-bucket tx channel
	hich chan<- buffer // high priority, unbuffered, no-drop tx channel

	rxErr error // error that stopped the rx service
	txErr error // error that stopped the tx service

	fd  int
	fda syscall.SockaddrLinklayer
}

// Connect @xeth socket and run channel service routines.
func Start(wg *sync.WaitGroup, stopch <-chan struct{}) (task *Task, err error) {
	xethif, err := net.InterfaceByName("xeth")
	if err != nil {
		return
	}

	atxeth, err := net.ResolveUnixAddr(unixpacket, "@xeth")
	if err != nil {
		return
	}

	muxfd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW,
		syscall.ETH_P_ALL)
	if err != nil {
		return
	}

	atsock, err := net.DialUnix(unixpacket, nil, atxeth)
	for err != nil {
		if !isEAGAIN(err) {
			return
		}
		atsock, err = net.DialUnix(unixpacket, nil, atxeth)
	}

	loch := make(chan buffer, 4)
	hich := make(chan buffer, 4)
	rxch := make(chan Buffer, 1024)

	task = &Task{
		RxCh:   rxch,
		wg:     wg,
		stopch: stopch,
		sock:   atsock,
		loch:   loch,
		hich:   hich,
		fd:     muxfd,
		fda: syscall.SockaddrLinklayer{
			Protocol: syscall.ETH_P_ARP,
			Ifindex:  xethif.Index,
			Hatype:   syscall.ARPHRD_ETHER,
		},
	}

	task.wg.Add(1)
	go task.goClose()
	task.wg.Add(1)
	go task.goRx(rxch)
	task.wg.Add(1)
	go task.goTx(loch, hich)

	return
}

func kind(buf buffer) uint8 {
	return (*internal.MsgHeader)(buf.pointer()).Kind
}

func Class(buf Buffer) int {
	switch kind(buf) {
	case internal.MsgKindBreak:
		return ClassBreak
	case internal.MsgKindChangeUpperXid,
		internal.MsgKindEthtoolFlags,
		internal.MsgKindEthtoolLinkModesSupported,
		internal.MsgKindEthtoolLinkModesAdvertising,
		internal.MsgKindEthtoolLinkModesLPAdvertising,
		internal.MsgKindEthtoolSettings,
		internal.MsgKindIfInfo:
		return ClassInterface
	case internal.MsgKindIfa, internal.MsgKindIfa6:
		return ClassAddress
	case internal.MsgKindFibEntry, internal.MsgKindFib6Entry:
		return ClassFib
	case internal.MsgKindNeighUpdate:
		return ClassNeighbor
	default:
		return ClassUnknown
	}
}

// parse driver message and cache ifinfo in xid maps.
func Parse(buf Buffer) interface{} {
	defer Parsed.Inc()
	defer buf.pool()
	switch kind(buf) {
	case internal.MsgKindBreak:
		return Break{}
	case internal.MsgKindChangeUpperXid:
		msg := (*internal.MsgChangeUpperXid)(buf.pointer())
		lower := Xid(msg.Lower)
		upper := Xid(msg.Upper)
		if msg.Linking != 0 {
			return lower.join(upper)
		} else {
			return lower.quit(upper)
		}
	case internal.MsgKindEthtoolFlags:
		msg := (*internal.MsgEthtoolFlags)(buf.pointer())
		return Xid(msg.Xid).RxEthtoolFlags(msg.Flags)
	case internal.MsgKindEthtoolLinkModesSupported:
		msg := (*internal.MsgEthtoolLinkModes)(buf.pointer())
		return Xid(msg.Xid).RxSupported(msg.Modes)
	case internal.MsgKindEthtoolLinkModesAdvertising:
		msg := (*internal.MsgEthtoolLinkModes)(buf.pointer())
		return Xid(msg.Xid).RxAdvertising(msg.Modes)
	case internal.MsgKindEthtoolLinkModesLPAdvertising:
		msg := (*internal.MsgEthtoolLinkModes)(buf.pointer())
		return Xid(msg.Xid).RxLPAdvertising(msg.Modes)
	case internal.MsgKindEthtoolSettings:
		msg := (*internal.MsgEthtoolSettings)(buf.pointer())
		return Xid(msg.Xid).RxEthtoolSettings(msg)
	case internal.MsgKindFibEntry:
		msg := (*internal.MsgFibEntry)(buf.pointer())
		return fib4(msg)
	case internal.MsgKindFib6Entry:
		msg := (*internal.MsgFib6Entry)(buf.pointer())
		return fib6(msg)
	case internal.MsgKindIfa:
		msg := (*internal.MsgIfa)(buf.pointer())
		xid := Xid(msg.Xid)
		if msg.Event == internal.IFA_ADD {
			return xid.RxIP4Add(msg.Address, msg.Mask)
		} else {
			return xid.RxIP4Del(msg.Address, msg.Mask)
		}
	case internal.MsgKindIfa6:
		msg := (*internal.MsgIfa6)(buf.pointer())
		xid := Xid(msg.Xid)
		if msg.Event == internal.IFA_ADD {
			addr := []byte(msg.Address[:])
			length := int(msg.Length)
			return xid.RxIP6Add(addr, length)
		} else {
			return xid.RxIP6Del(msg.Address[:])
		}
	case internal.MsgKindIfInfo:
		msg := (*internal.MsgIfInfo)(buf.pointer())
		xid := Xid(msg.Xid)
		switch msg.Reason {
		case internal.IfInfoReasonNew:
			return xid.RxIfInfo(msg)
		case internal.IfInfoReasonDump:
			return xid.RxIfInfo(msg)
		case internal.IfInfoReasonDel:
			return RxDelete(xid)
		case internal.IfInfoReasonUp:
			return xid.RxUp()
		case internal.IfInfoReasonDown:
			return xid.RxDown()
		case internal.IfInfoReasonReg:
			netns := NetNs(msg.Net)
			return xid.RxReg(netns)
		case internal.IfInfoReasonUnreg:
			return xid.RxUnreg()
		}
	case internal.MsgKindNeighUpdate:
		msg := (*internal.MsgNeighUpdate)(buf.pointer())
		return neighbor(msg)
	}
	return nil
}

func Pool(msg interface{}) {
	if method, found := msg.(pooler); found {
		method.Pool()
	}
}

// request fib dump
func (task *Task) DumpFib() {
	buf := newBuffer(internal.SizeofMsgDumpFibInfo)
	msg := (*internal.MsgHeader)(buf.pointer())
	msg.Set(internal.MsgKindDumpFibInfo)
	task.hich <- buf
}

// request ifinfo dump
func (task *Task) DumpIfInfo() {
	buf := newBuffer(internal.SizeofMsgDumpIfInfo)
	msg := (*internal.MsgHeader)(buf.pointer())
	msg.Set(internal.MsgKindDumpIfInfo)
	task.hich <- buf
}

// Send an exception frame to driver through raw socket.
func (task *Task) ExceptionFrame(b []byte) {
	const (
		hitci = (2 * syscall.ARPHRD_IEEE802) + 2
		lotci = hitci + 1
		prio  = (7 << (VLAN_PRIO_SHIFT - 8))
	)
	// set priority so that the xeth will forward to the
	// respective upper device rather than it's port
	b[hitci] |= prio
	syscall.Sendto(task.fd, b, 0, &task.fda)
}

// Send carrier change to driver through hi-priority channel.
func (task *Task) SetCarrier(xid Xid, on bool) {
	buf := newBuffer(internal.SizeofMsgCarrier)
	msg := (*internal.MsgCarrier)(buf.pointer())
	msg.Header.Set(internal.MsgKindCarrier)
	msg.Xid = uint32(xid)
	if on {
		msg.Flag = internal.CarrierOn
	} else {
		msg.Flag = internal.CarrierOff
	}
	task.hich <- buf
}

// Send ethtool stat change to driver through leaky-bucket channel.
func (task *Task) SetEthtoolStat(xid Xid, stat uint32, n uint64) {
	task.setStat(internal.MsgKindEthtoolStat, xid, stat, n)
}

// Send link stat change to driver through leaky-bucket channel.
func (task *Task) SetLinkStat(xid Xid, stat uint32, n uint64) {
	task.setStat(internal.MsgKindLinkStat, xid, stat, n)
}

func (task *Task) setStat(kind uint8, xid Xid, stat uint32, n uint64) {
	buf := newBuffer(internal.SizeofMsgStat)
	msg := (*internal.MsgStat)(buf.pointer())
	msg.Header.Set(kind)
	msg.Xid = uint32(xid)
	msg.Index = stat
	msg.Count = n
	task.hich <- buf
}

// Send speed change to driver through hi-priority channel.
func (task *Task) SetSpeed(xid Xid, mbps uint32) {
	buf := newBuffer(internal.SizeofMsgSpeed)
	msg := (*internal.MsgSpeed)(buf.pointer())
	msg.Header.Set(internal.MsgKindSpeed)
	msg.Xid = uint32(xid)
	msg.Mbps = mbps
	task.hich <- buf
}

// Wait for stop signal then shutdown and close socket
func (task *Task) goClose() {
	defer task.wg.Done()
	<-task.stopch
	if task.sock == nil {
		return
	}
	sock := task.sock
	task.sock = nil
	f, err := sock.File()
	if err == nil {
		syscall.Shutdown(int(f.Fd()), SHUT_RDWR)
	}
	sock.Close()
	syscall.Close(task.fd)
}

func (task *Task) goRx(rxch chan<- Buffer) {
	const minrxto = 10 * time.Millisecond
	const maxrxto = 320 * time.Millisecond

	rxto := minrxto
	rxbuf := make([]byte, PageSize, PageSize)
	rxoob := make([]byte, PageSize, PageSize)
	ptr := unsafe.Pointer(&rxbuf[0])
	h := (*internal.MsgHeader)(ptr)

	defer func() {
		rxbuf = rxbuf[:0]
		rxoob = rxoob[:0]
		close(rxch)
		task.wg.Done()
	}()

	for {
		select {
		case <-task.stopch:
			return
		default:
		}
		task.rxErr = task.sock.SetReadDeadline(time.Now().Add(rxto))
		if task.rxErr != nil {
			break
		}
		n, noob, flags, addr, err := task.sock.ReadMsgUnix(rxbuf, rxoob)
		_ = noob
		_ = flags
		_ = addr
		select {
		case <-task.stopch:
			return
		default:
		}
		if n == 0 || isTimeout(err) {
			if rxto < maxrxto {
				rxto *= 2
			}
		} else if err != nil {
			e, ok := err.(*os.SyscallError)
			if !ok || e.Err.Error() != "EOF" {
				task.rxErr = err
			}
			break
		} else if task.rxErr = h.Validate(rxbuf[:n]); task.rxErr != nil {
			break
		} else {
			rxto = minrxto
			rxch <- cloneBuffer(rxbuf[:n])
			Cloned.Inc()
		}
	}
}

func (task *Task) goTx(loch, hich <-chan buffer) {
	defer task.wg.Done()

	for task.txErr == nil {
		select {
		case <-task.stopch:
			return
		case buf, ok := <-hich:
			if ok {
				task.txErr = task.tx(buf, 0)
			} else {
				return
			}
		case buf, ok := <-loch:
			if ok {
				task.txErr = task.tx(buf, 10*time.Millisecond)
			} else {
				return
			}
		}
	}
}

// Send through low-priority, leaky-bucket.
func (task *Task) queue(buf buffer) {
	select {
	case task.loch <- buf:
	default:
		buf.pool()
		Dropped.Inc()
	}
}

func (task *Task) tx(buf buffer, timeout time.Duration) error {
	var oob []byte
	var dl time.Time
	defer buf.pool()
	if task.sock == nil {
		return io.EOF
	}
	if timeout != time.Duration(0) {
		dl = time.Now().Add(timeout)
	}
	err := task.sock.SetWriteDeadline(dl)
	if err != nil {
		return err
	}
	_, _, err = task.sock.WriteMsgUnix(buf.bytes(), oob, nil)
	if err == nil {
		Sent.Inc()
		if kind(buf) == internal.MsgKindCarrier {
			msg := (*internal.MsgCarrier)(buf.pointer())
			xid := Xid(msg.Xid)
			LinkOf(xid).LinkUp(msg.Flag == internal.CarrierOn)
		}
	}
	return err
}

func isEAGAIN(err error) bool {
	if err != nil {
		if operr, ok := err.(*net.OpError); ok {
			if oserr, ok := operr.Err.(*os.SyscallError); ok {
				if oserr.Err == syscall.EAGAIN {
					return true
				}
			}
		}
	}
	return false
}

func isTimeout(err error) bool {
	if err != nil {
		if op, ok := err.(*net.OpError); ok {
			return op.Timeout()
		}
	}
	return false
}
