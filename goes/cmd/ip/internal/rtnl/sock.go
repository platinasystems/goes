// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"unsafe"
)

const SizeofInt = (32 << (^uint(0) >> 63)) >> 3
const (
	debugSockDone = false
	debugSockGrps = false
)

var Eclosed = errors.New("already closed")

// This creates cascaded channels and attendant go-routines to pack received
// netlink messages into a page, if possible; otherwise, one per oversized
// buffer.
//
// Usage: NewSock([depth int[, groups uint32[, allnsid bool[,
//	sorcvbuf[, sosndbuf] int]]]])
// e.g.
//	NewSock()
//	NewSock(16)
//	NewSock(16, RTNLGRP_NEIGH.Bit())
//	NewSock(16, groups, true)
//	NewSock(16, groups, false, 4 << 20)
//	NewSock(16, groups, false, 4 << 20, 1 << 20)
//
//	depth	of Rx buffer channel (default, 4)
//	groups	to listen (default, none)
//	allnsid	listen in all identified net namespaces (default, false)
//	sorcvbuf, sosndbuf
//		respective receive and send socket buffer size
//		(default, kernel config)
func NewSock(opts ...interface{}) (*Sock, error) {
	var allnsid bool
	var groups uint32
	depth := 4
	sorcvbuf := -1
	sosndbuf := -1

	if len(opts) > 0 {
		depth = opts[0].(int)
	}
	if len(opts) > 1 {
		groups = opts[1].(uint32)
	}
	if len(opts) > 2 {
		allnsid = opts[2].(bool)
	}
	if len(opts) > 3 {
		sorcvbuf = opts[3].(int)
	}
	if len(opts) > 3 {
		sosndbuf = opts[3].(int)
	}

	fd, err := syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW,
		syscall.NETLINK_ROUTE)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}
	defer func() {
		if err != nil {
			syscall.Close(fd)
		}
	}()
	sa := &syscall.SockaddrNetlink{
		Family: syscall.AF_NETLINK,
		Groups: groups,
	}
	if err = syscall.Bind(fd, sa); err != nil {
		return nil, os.NewSyscallError("bind", err)
	}
	gsa, err := syscall.Getsockname(fd)
	if err != nil {
		return nil, os.NewSyscallError("getsockname", err)
	}
	if debugSockGrps {
		printRtnlGrps(gsa.(*syscall.SockaddrNetlink).Groups)
	}
	if allnsid {
		err = os.NewSyscallError("NETLINK_LISTEN_ALL_NSID",
			syscall.SetsockoptInt(fd, SOL_NETLINK,
				NETLINK_LISTEN_ALL_NSID, 1))
		if err != nil {
			return nil, err
		}
	}
	for _, x := range []struct {
		opt  int
		val  int
		name string
	}{
		{syscall.SO_RCVBUF, sorcvbuf, "SO_RCVBUF"},
		{syscall.SO_SNDBUF, sosndbuf, "SO_SNDBUF"},
	} {
		if x.val > 0 {
			err = os.NewSyscallError(x.name,
				syscall.SetsockoptInt(fd, syscall.SOL_SOCKET,
					x.opt, x.val))
			if err != nil {
				return nil, err
			}
			// Verify buffer size is at least as large.
			var ver int
			ver, err = syscall.GetsockoptInt(fd,
				syscall.SOL_SOCKET, x.opt)
			if err != nil {
				return nil, fmt.Errorf("%s: can't verify: %v",
					x.name, err)
			}
			if ver < x.val {
				err = fmt.Errorf("%s: truncated", x.name)
				return nil, err
			}
		}
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	rxch := make(chan []byte, depth)
	sock := &Sock{
		Pid:  gsa.(*syscall.SockaddrNetlink).Pid,
		RxCh: rxch,
		rxch: rxch,
		fd:   fd,
		sa:   sa,
		pr:   pr,
		pw:   pw,

		rxdone:  make(chan struct{}),
		allnsid: allnsid,
	}
	rxrdy := make(chan struct{})
	go sock.gorx(rxrdy)
	<-rxrdy
	return sock, nil
}

type Sock struct {
	Pid  uint32
	RxCh <-chan []byte
	rxch chan<- []byte
	Err  error
	fd   int
	sa   *syscall.SockaddrNetlink
	// use a pipe to signal close to gogorx
	pr      *os.File
	pw      *os.File
	rxdone  chan struct{}
	allnsid bool
}

func (sock *Sock) Close() error {
	if sock.pw == nil {
		return Eclosed
	}
	sock.pw.Close()
	sock.pw = nil
	<-sock.rxdone
	if debugSockDone {
		fmt.Println("rxdone")
	}
	return nil
}

func (sock *Sock) Send(b []byte) error {
	return syscall.Sendto(sock.fd, b, 0, sock.sa)
}

func (sock *Sock) gorx(rdy chan<- struct{}) {
	var hbuf [SizeofHdr]byte
	h := (*Hdr)(unsafe.Pointer(&hbuf[0]))

	bytech := make(chan byte, PAGE.Size())
	gogorxrdy := make(chan struct{})

	go sock.gogorx(bytech, gogorxrdy)
	<-gogorxrdy

	close(rdy)
	defer close(sock.rxdone)
	defer close(sock.rxch)
	if debugSockDone {
		defer fmt.Println("gorx done")
	}

	for b := make([]byte, 0, PAGE.Size()); ; {
		for i := range hbuf {
			if c, opened := <-bytech; !opened {
				return
			} else {
				hbuf[i] = c
			}
		}
		n := NLMSG.Align(int(h.Len))
		if len(b)+n > cap(b) {
			if len(b) > 0 {
				sock.rxch <- b
			}
			sz := PAGE.Size()
			if sz < n {
				sz = n
			}
			b = make([]byte, 0, sz)
		}
		b = append(b, hbuf[:]...)
		for i := SizeofHdr; i < n; i++ {
			c, ok := <-bytech
			if !ok {
				return
			} else if i < cap(b) {
				b = append(b, c)
			}
			if len(b) == cap(b) {
				sock.rxch <- b
				b = make([]byte, 0, PAGE.Size())
			}
		}
		if len(b) > 0 && len(bytech) == 0 {
			sock.rxch <- b
			b = make([]byte, 0, PAGE.Size())
		}
	}
}

func (sock *Sock) gogorx(bytech chan<- byte, rdy chan<- struct{}) {
	b := make([]byte, PAGE.Size())
	oob := make([]byte, PAGE.Size())

	// When changed, the out-of-band nsid is forwarded as an internal
	// typed, in-band message.
	nsidbuf := make([]byte, SizeofHdr+SizeofInt)
	*(*Hdr)(unsafe.Pointer(&nsidbuf[0])) = Hdr{
		Len:  SizeofHdr + SizeofInt,
		Type: NLMSG_NSID,
	}
	nsidptr := (*int)(unsafe.Pointer(&nsidbuf[SizeofHdr]))
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

	close(rdy)
	defer close(bytech)
	if debugSockDone {
		defer fmt.Println("gogorx done")
	}
	defer sock.pr.Close()
	defer syscall.Close(sock.fd)

	prfd := int(sock.pr.Fd())

	for {
		var n, noob int
		var rfds syscall.FdSet
		FD_ZERO(&rfds)
		FD_SET(&rfds, sock.fd)
		FD_SET(&rfds, prfd)
		tv := syscall.Timeval{10, 0}
		n, sock.Err = syscall.Select(prfd+1, &rfds, nil, nil, &tv)
		if sock.Err != nil {
			break
		}
		if n == 0 {
			continue
		}
		if FD_ISSET(&rfds, prfd) {
			// Sock.Close
			break
		}
		if !FD_ISSET(&rfds, sock.fd) {
			continue
		}
		// peek first to see if we need to expand buffer
		n, _, sock.Err = syscall.Recvfrom(sock.fd, b,
			syscall.MSG_PEEK|syscall.MSG_TRUNC|syscall.MSG_DONTWAIT)
		if sock.Err != nil {
			break
		}
		if n == 0 {
			continue
		}
		if n > len(b) {
			b = make([]byte, PAGE.Align(n))
		}
		n, noob, _, _, sock.Err = syscall.Recvmsg(sock.fd, b, oob,
			syscall.MSG_DONTWAIT)
		if sock.Err != nil {
			break
		}
		if n == 0 {
			continue
		}
		if sock.allnsid {
			nsid := -1
			if noob > 0 {
				nsid = oobnsid(oob[:noob])
			}
			if nsid != lastnsid {
				*nsidptr = nsid
				for _, c := range nsidbuf {
					bytech <- c
				}
				lastnsid = nsid
			}
		}
		for _, c := range b[:n] {
			bytech <- c
		}
	}
	if sock.Err == nil || sock.Err == syscall.EINTR ||
		sock.Err == syscall.EBADF {
		sock.Err = io.EOF
	}
}
