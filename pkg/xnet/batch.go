// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build android || linux

package xnet

import (
	"context"
	"net/netip"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xos"
	"golang.org/x/sys/unix"
)

const BatchCap = 16
const CanBatch = true

type MMsghdr struct {
	Hdr unix.Msghdr
	Len uint32
	_   [4]byte
}

const (
	SizeofCmsghdr = int(unsafe.Sizeof(unix.Cmsghdr{}))
	SizeofIovec   = int(unsafe.Sizeof(unix.Iovec{}))
	SizeofMsghdr  = int(unsafe.Sizeof(unix.Msghdr{}))
	SizeofMmsghdr = int(unsafe.Sizeof(MMsghdr{}))
	SAsCap        = BatchCap * SizeofSockaddrInet6
)

type msgBatch struct {
	iovs []unix.Iovec
	mmhs []MMsghdr
	msgs []*Msg
	sas  []byte
}

func newBatch() *msgBatch {
	mb := &msgBatch{
		iovs: make([]unix.Iovec, BatchCap, BatchCap),
		mmhs: make([]MMsghdr, BatchCap, BatchCap),
		msgs: make([]*Msg, BatchCap, BatchCap),
		sas:  make([]byte, SAsCap, SAsCap),
	}
	for i := range mb.mmhs {
		mb.mmhs[i].Hdr.Iov = &mb.iovs[i]
		mb.mmhs[i].Hdr.Iovlen = 1
		mb.mmhs[i].Hdr.Controllen = 0
		mb.mmhs[i].Hdr.Name = (*byte)(&mb.sas[i*SizeofSockaddrInet6])
		mb.mmhs[i].Hdr.Namelen = uint32(SizeofSockaddrInet6)
	}
	return mb
}

func (mb *msgBatch) putMsgs(mp *MsgPool) {
	for i, m := range mb.msgs {
		if m == nil {
			break
		}
		mp.Put(m)
		mb.msgs[i] = nil
		mb.iovs[i].Base = nil
		mb.mmhs[i].Hdr.Namelen = uint32(SizeofSockaddrInet6)
	}
}

func (mb *msgBatch) sa(i int) SAInBuf {
	j := i * SizeofSockaddrInet6
	return SAInBuf(mb.sas[j : j+SizeofSockaddrInet6])
}

// Copy to pooled messages from socket with “recvmmsg” then send to channel
// until context is done; then close channel before returning.
func (mp *MsgPool) RecvBatchService(
	ctx context.Context, ch chan<- *Msg, conn syscall.Conn,
) error {
	defer close(ch)

	mb := newBatch()

	var flags int
	if runtime.NumCPU() > 1 {
		flags = unix.MSG_WAITFORONE
	} else {
		flags = unix.MSG_DONTWAIT
	}

	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	for {
		for i, m := range mb.msgs {
			if m != nil {
				break
			}
			m := mp.Get()
			mb.msgs[i] = m
			ptr := unsafe.Pointer(&m.Data[0])
			mb.iovs[i].Base = (*byte)(ptr)
			mb.iovs[i].Len = uint64(len(m.Data))
		}
		var n int
		raw.Read(func(sock uintptr) bool {
			n, err = RecvMMsg(sock, mb.mmhs, flags)
			if err == nil {
				return true
			} else if xos.IsBlocked(err) {
				err = ctx.Err()
				return err == nil
			} else {
				return true
			}
		})
		if err != nil {
			break
		}
		for i := 0; i < n; i++ {
			m := mb.msgs[i]
			m.Data = m.Data[:mb.mmhs[i].Len]
			mb.msgs[i].AddrPort, _ = mb.sa(i).Decode()
			ch <- m
			mb.msgs[i] = nil
		}
	}
	return xerrors.Suppress(err, context.Canceled)
}

func RecvMMsg(sock uintptr, hs []MMsghdr, flags int) (int, error) {
	const trap = unix.SYS_RECVMMSG
	var err error
	hsp := uintptr(unsafe.Pointer(&hs[0]))
	hsn := uintptr(len(hs))
	𝑓lags := uintptr(flags)
	n, _, errno := syscall.Syscall6(trap, sock, hsp, hsn, 𝑓lags, 0, 0)
	if errno != 0 {
		err = errno
	}
	return int(n), err
}

// Copy messages from channel to socket with “sendmmsg” and return to pool
// until the channel is closed.
func (mp *MsgPool) SendBatchService(conn syscall.Conn, ch <-chan *Msg) error {
	const flags = 0
	var rap netip.AddrPort
	if raer, ok := conn.(RemoteAddrer); ok {
		if netra := raer.RemoteAddr(); netra != nil {
			if aper, ok := netra.(AddrPorter); ok {
				rap = aper.AddrPort()
			}
		}
	}

	mb := newBatch()

	raw, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	defer mb.putMsgs(mp)

	for err == nil {
		var n int
	selection:
		for n = 0; n < BatchCap; n++ {
			select {
			case m, ok := <-ch:
				if !ok {
					return nil
				}
				if !m.AddrPort.IsValid() {
					if !rap.IsValid() {
						return xerrors.
							Invalid("address")
					}
					m.AddrPort = rap
				}
				mb.msgs[n] = m
				ptr := unsafe.Pointer(&m.Data[0])
				mb.iovs[n].Base = (*byte)(ptr)
				mb.iovs[n].Len = uint64(len(m.Data))
				mb.mmhs[n].Hdr.Namelen =
					uint32(mb.sa(n).Encode(m.AddrPort, 0))
			default:
				break selection
			}
		}
		for i, sent := 0, 0; err == nil && i < n; i += sent {
			raw.Write(func(sock uintptr) bool {
				sent, err = SendMMsg(sock, mb.mmhs[i:n], flags)
				return !xos.IsBlocked(err)
			})

		}
		mb.putMsgs(mp)
	}
	return err
}

func SendMMsg(sock uintptr, hs []MMsghdr, flags int) (int, error) {
	const trap = unix.SYS_SENDMMSG
	var err error
	hsp := uintptr(unsafe.Pointer(&hs[0]))
	hsn := uintptr(len(hs))
	𝑓lags := uintptr(flags)
	n, _, errno := syscall.Syscall6(trap, sock, hsp, hsn, 𝑓lags, 0, 0)
	if errno != 0 {
		err = errno
	}
	return int(n), err
}
