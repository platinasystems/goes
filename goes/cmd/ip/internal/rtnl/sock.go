// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const SizeofInt = (32 << (^uint(0) >> 63)) >> 3

var Eclosed = errors.New("already closed")

// This creates cascaded channels and attendant go-routines to pack received
// netlink messages into a page, if possible; otherwise, one per oversized
// sized buffer. NonBlockingRecv() and [blocking]Recv() return a slice
// reference to the next message without copy.
//
// Usage: NewSock([depth int[, groups uint32[, allnsid bool]]])
// e.g.
//	NewSock()
//	NewSock(16)
//	NewSock(16, RTNLGRP_NEIGH.Bit())
//	NewSock(16, RTNLGRP_NEIGH.Bit() | RTNLGRP_IPV4_IFADDR.Bit(), true)
//
//	depth	of Rx buffer channel (default, 4)
//	groups	to listen (default, none)
//	allnsid	listen in all identified net namespaces (default, false)
func NewSock(opts ...interface{}) (*Sock, error) {
	var allnsid bool
	var groups uint32
	depth := 4

	if len(opts) > 0 {
		depth = opts[0].(int)
	}
	if len(opts) > 1 {
		groups = opts[1].(uint32)
	}
	if len(opts) > 2 {
		allnsid = opts[2].(bool)
	}

	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW,
		syscall.NETLINK_ROUTE)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}
	sa := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: groups,
	}
	if err := syscall.Bind(fd, sa); err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("bind", err)
	}
	gsa, err := syscall.Getsockname(fd)
	if err != nil {
		syscall.Close(fd)
		return nil, os.NewSyscallError("getsockname", err)
	}
	if false { // debug
		printRtnlGrps(gsa.(*syscall.SockaddrNetlink).Groups)
	}
	if allnsid {
		err = syscall.SetsockoptInt(fd, SOL_NETLINK,
			NETLINK_LISTEN_ALL_NSID, 1)
		if err != nil {
			syscall.Close(fd)
			return nil, os.NewSyscallError("setsockopt", err)
		}
	}
	sock := &Sock{
		fd:  fd,
		Pid: gsa.(*syscall.SockaddrNetlink).Pid,
		rx:  make(chan []byte, depth),
		b:   Empty,
		sa:  sa,

		allnsid: allnsid,
	}
	rxrdy := make(chan struct{})
	go sock.gorx(rxrdy)
	<-rxrdy
	return sock, nil
}

type Sock struct {
	fd      int
	Pid     uint32
	rx      chan []byte
	b       []byte
	closed  bool
	allnsid bool
	sa      *syscall.SockaddrNetlink
	err     error
}

func (sock *Sock) Close() error {
	if !sock.closed {
		sock.closed = true
		return syscall.Close(sock.fd)
	}
	return Eclosed
}

// Recv next message buffer from channel.
//
// Usage: Recv([dontblock bool])
func (sock *Sock) Recv(opts ...interface{}) ([]byte, error) {
	var dontblock bool

	if len(opts) > 0 {
		dontblock = opts[0].(bool)
	}

	if len(sock.b) == 0 {
		if dontblock && len(sock.rx) == 0 {
			return Empty, nil
		}
		if rxb, ok := <-sock.rx; !ok {
			sock.b = Empty
			return Empty, sock.err
		} else {
			sock.b = rxb
		}
	}
	h := HdrPtr(sock.b)
	n := Align(int(h.Len))
	b := sock.b[:int(h.Len)]
	if n == len(sock.b) {
		sock.b = Empty
	} else {
		sock.b = sock.b[n:]
	}
	return b, nil
}

func (sock *Sock) Send(b []byte) error {
	return syscall.Sendto(sock.fd, b, 0, sock.sa)
}

// After setting the sequence number, send the request to netlink and call the
// given handler for each received message until DONE or ERROR.
func (sock *Sock) UntilDone(req []byte, do func([]byte)) error {
	seq := Seq()
	HdrPtr(req).Seq = seq
	if err := sock.Send(req); err != nil {
		return err
	}
	for {
		b, err := sock.Recv()
		if err != nil {
			return err
		}
		if len(b) < SizeofHdr {
			return syscall.ENOMSG
		}
		h := HdrPtr(b)
		if h == nil {
			return Ehdr(len(b))
		}
		if int(h.Len) != len(b) {
			return h.Elen()
		}
		if sock.Pid != h.Pid {
			return h.Epid()
		}
		if seq != h.Seq {
			return h.Eseq()
		}
		switch h.Type {
		case NLMSG_DONE:
			return nil
		case NLMSG_ERROR:
			e := NlmsgerrPtr(b)
			if e == nil || e.Errno == 0 {
				return nil
			}
			return os.NewSyscallError("nack",
				syscall.Errno(e.Errno))
		default:
			do(b)
		}
	}
	return nil
}

func (sock *Sock) IfNamesByIndex() (map[int32]string, error) {
	ifnames := make(map[int32]string)
	if req, err := NewMessage(
		Hdr{
			Type:  RTM_GETLINK,
			Flags: NLM_F_REQUEST | NLM_F_DUMP,
		},
		IfInfoMsg{
			Family: AF_UNSPEC,
		},
	); err != nil {
		return nil, err
	} else if err = sock.UntilDone(req, func(b []byte) {
		if HdrPtr(b).Type != RTM_NEWLINK {
			return
		}
		var ifla Ifla
		ifla.Write(b)
		msg := IfInfoMsgPtr(b)
		if val := ifla[IFLA_IFNAME]; len(val) > 0 {
			ifnames[msg.Index] = Kstring(val)
		}
	}); err != nil {
		return nil, err
	}
	return ifnames, nil
}

func (sock *Sock) gorx(rdy chan<- struct{}) {
	defer close(sock.rx)

	var hbuf [SizeofHdr]byte
	h := (*Hdr)(unsafe.Pointer(&hbuf[0]))
	pgsz := syscall.Getpagesize()

	bytes := make(chan byte, pgsz)
	gogorxrdy := make(chan struct{})
	go sock.gogorx(bytes, gogorxrdy)
	<-gogorxrdy
	close(rdy)

	for b := make([]byte, 0, pgsz); ; {
		for i := range hbuf {
			c, ok := <-bytes
			if !ok {
				return
			}
			hbuf[i] = c
		}
		if false && h.Type == NLMSG_DONE {
			return
		}
		n := Align(int(h.Len))
		if len(b)+n > cap(b) {
			if len(b) > 0 {
				sock.rx <- b
			}
			sz := pgsz
			if sz < n {
				sz = n
			}
			b = make([]byte, 0, sz)
		}
		b = append(b, hbuf[:]...)
		for i := SizeofHdr; i < n; i++ {
			c, ok := <-bytes
			if !ok {
				return
			} else if i < cap(b) {
				b = append(b, c)
			}
			if len(b) == cap(b) {
				sock.rx <- b
				b = make([]byte, 0, pgsz)
			}
		}
		if len(b) > 0 && len(bytes) == 0 {
			sock.rx <- b
			b = make([]byte, 0, pgsz)
		}
	}
}

func (sock *Sock) gogorx(ch chan<- byte, rdy chan<- struct{}) {
	defer close(ch)
	close(rdy)
	pgsz := syscall.Getpagesize()
	b := make([]byte, 32*pgsz)
	oob := make([]byte, pgsz)
	nsidbuf := make([]byte, SizeofHdr+SizeofInt)
	*(*Hdr)(unsafe.Pointer(&nsidbuf[0])) = Hdr{
		Len:  SizeofHdr + SizeofInt,
		Type: NLMSG_NSID,
	}
	lastnsid := -1
	oobnsid := func(oob []byte) int {
		scms, err := syscall.ParseSocketControlMessage(oob)
		if err != nil {
			return -1
		}
		for _, scm := range scms {
			if scm.Header.Level != SOL_NETLINK {
				continue
			}
			if scm.Header.Type != NETLINK_LISTEN_ALL_NSID {
				continue
			}
			return *(*int)(unsafe.Pointer(&scm.Data[0]))
		}
		return -1
	}
	for {
		if sock.closed {
			break
		}
		n, noob, _, _, err := syscall.Recvmsg(sock.fd, b, oob, 0)
		if err != nil {
			if sock.err == nil {
				sock.err = err
			}
			break
		}
		if sock.allnsid {
			nsid := -1
			if noob > 0 {
				nsid = oobnsid(oob[:noob])
			}
			if nsid != lastnsid {
				*(*int)(unsafe.Pointer(&nsidbuf[SizeofHdr])) =
					nsid
				for _, c := range nsidbuf {
					ch <- c
				}
				lastnsid = nsid
			}
		}
		for _, c := range b[:n] {
			ch <- c
		}
	}
}

// Send a RTM_GETNSID request to netlink and return the response attribute.
func (sock *Sock) Nsid(name string) (int32, error) {
	nsid := int32(-1)

	f, err := os.Open(filepath.Join(VarRunNetns, name))
	if err != nil {
		return -1, err
	}
	defer f.Close()

	req, err := NewMessage(Hdr{
		Type:  RTM_GETNSID,
		Flags: NLM_F_REQUEST | NLM_F_ACK,
	}, NetnsMsg{
		Family: AF_UNSPEC,
	},
		Attr{NETNSA_FD, Uint32Attr(f.Fd())},
	)
	if err != nil {
		return -1, err
	}

	err = sock.UntilDone(req, func(b []byte) {
		var netnsa Netnsa
		if HdrPtr(b).Type != RTM_NEWNSID {
			return
		}
		n, err := netnsa.Write(b)
		if err != nil || n == 0 {
			return
		}
		if val := netnsa[NETNSA_NSID]; len(val) > 0 {
			nsid = Int32(val)
		}
	})

	return nsid, err
}
