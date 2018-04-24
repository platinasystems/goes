// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"
)

var (
	DefaultMulticastGroups = []MulticastGroup{
		RTNLGRP_LINK,
		RTNLGRP_NEIGH,
		RTNLGRP_IPV4_IFADDR,
		RTNLGRP_IPV4_ROUTE,
		RTNLGRP_IPV4_MROUTE,
		RTNLGRP_IPV6_IFADDR,
		RTNLGRP_IPV6_ROUTE,
		RTNLGRP_IPV6_MROUTE,
		RTNLGRP_NSID,
	}
	LinkMulticastGroups = []MulticastGroup{
		RTNLGRP_LINK,
	}
	AddrMulticastGroups = []MulticastGroup{
		RTNLGRP_IPV4_IFADDR,
		RTNLGRP_IPV6_IFADDR,
	}
	RouteMulticastGroups = []MulticastGroup{
		RTNLGRP_IPV4_ROUTE,
		RTNLGRP_IPV6_ROUTE,
		RTNLGRP_IPV4_MROUTE,
		RTNLGRP_IPV6_MROUTE,
	}
	NeighborMulticastGroups = []MulticastGroup{
		RTNLGRP_NEIGH,
	}
	NsidMulticastGroups = []MulticastGroup{
		RTNLGRP_NSID,
	}
	DefaultListenReqs = []ListenReq{
		{RTM_GETNSID, AF_UNSPEC},
		{RTM_GETLINK, AF_PACKET},
		{RTM_GETADDR, AF_INET},
		{RTM_GETADDR, AF_INET6},
		{RTM_GETNEIGH, AF_INET},
		{RTM_GETNEIGH, AF_INET6},
		{RTM_GETROUTE, AF_INET},
		{RTM_GETROUTE, AF_INET6},
	}
	LinkListenReqs = []ListenReq{
		{RTM_GETLINK, AF_PACKET},
	}
	AddrListenReqs = []ListenReq{
		{RTM_GETADDR, AF_INET},
		{RTM_GETADDR, AF_INET6},
	}
	RouteListenReqs = []ListenReq{
		{RTM_GETROUTE, AF_INET},
		{RTM_GETROUTE, AF_INET6},
	}
	NeighborListenReqs = []ListenReq{
		{RTM_GETNEIGH, AF_INET},
		{RTM_GETNEIGH, AF_INET6},
	}
	NsidListenReqs = []ListenReq{
		{RTM_GETNSID, AF_UNSPEC},
	}
	NoopListenReq = ListenReq{NLMSG_NOOP, AF_UNSPEC}
	PageSize      int
)

func init() { PageSize = syscall.Getpagesize() }

// Zero means use default value.
type SocketConfig struct {
	RxBytes int
	TxBytes int

	RxMessages int
	TxMessages int

	DontListenAllNsid bool

	Groups []MulticastGroup
}

type Handler func(Message) error

type ListenReq struct {
	MsgType
	AddressFamily
}

type Socket struct {
	wg   sync.WaitGroup
	fd   int
	addr *syscall.SockaddrNetlink
	Rx   <-chan Message
	rx   chan<- Message
	Tx   chan<- Message
	tx   <-chan Message
	SocketConfig

	// Pipe which exists just for gorx to have a file descriptor to detect socket close.
	rx_close_kludge_pipe [2]int
}

func New(groups ...MulticastGroup) (*Socket, error) {
	return NewWithConfig(SocketConfig{
		Groups: groups,

		DontListenAllNsid: true,
	})
}

func NewWithConfig(cf SocketConfig) (s *Socket, err error) {
	return NewWithConfigAndFile(cf, -1)
}

func NewWithConfigAndFile(cf SocketConfig, fd int) (s *Socket, err error) {
	if fd < 0 {
		fd, err = syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_ROUTE)
		if err != nil {
			err = os.NewSyscallError("socket", err)
			return
		}
	}

	if cf.RxMessages == 0 {
		cf.RxMessages = DefaultMessages
	}
	rx := make(chan Message, cf.RxMessages)

	if cf.TxMessages == 0 {
		cf.TxMessages = DefaultMessages
	}
	tx := make(chan Message, cf.TxMessages)

	if len(cf.Groups) == 0 {
		cf.Groups = DefaultMulticastGroups
	}

	s = &Socket{
		fd: fd,

		Rx: rx,
		rx: rx,
		Tx: tx,
		tx: tx,

		SocketConfig: cf,
	}

	err = syscall.Pipe(s.rx_close_kludge_pipe[:])
	if err != nil {
		err = os.NewSyscallError("pipe rx_close_pipe", err)
		return
	}

	defer func() {
		if err != nil {
			if fd > 0 {
				syscall.Close(fd)
				if s != nil {
					if s.rx != nil {
						close(s.rx)
					}
					if s.Tx != nil {
						close(s.Tx)
					}
					if s.rx_close_kludge_pipe[0] > 0 {
						syscall.Close(s.rx_close_kludge_pipe[0])
						syscall.Close(s.rx_close_kludge_pipe[1])
					}
					s.addr = nil
					s = nil
				}
			}
		}
	}()

	addr := syscall.SockaddrNetlink{
		Family: uint16(AF_NETLINK),
	}

	for _, group := range cf.Groups {
		if group != NOOP_RTNLGRP {
			addr.Groups |= 1 << group
		}
	}

	err = os.NewSyscallError("bind", syscall.Bind(s.fd, &addr))
	if err != nil {
		return
	}

	xaddr, err := syscall.Getsockname(s.fd)
	err = os.NewSyscallError("Getsockname", err)
	if err != nil {
		return
	}
	s.addr = xaddr.(*syscall.SockaddrNetlink)

	// Increase socket buffering.
	if s.RxBytes != 0 {
		err = os.NewSyscallError("setsockopt SO_RCVBUF",
			syscall.SetsockoptInt(s.fd, syscall.SOL_SOCKET,
				syscall.SO_RCVBUF, s.RxBytes))
		if err != nil {
			return
		}
		// Verify buffer size is at least as large as requested.
		var v int
		v, err = syscall.GetsockoptInt(s.fd, syscall.SOL_SOCKET,
			syscall.SO_RCVBUF)
		if err != nil {
			return
		} else if v < s.RxBytes {
			err = fmt.Errorf(`
SO_RCVBUF truncated to %d bytes; run: sysctl -w net.core.rmem_max=%d`[1:],
				v, s.RxBytes)
			return
		}
	}

	if s.TxBytes != 0 {
		err = os.NewSyscallError("setsockopt SO_SNDBUF",
			syscall.SetsockoptInt(s.fd, syscall.SOL_SOCKET,
				syscall.SO_SNDBUF, s.TxBytes))
		if err != nil {
			return
		}
		// Verify buffer size is at least as large as requested.
		var v int
		v, err = syscall.GetsockoptInt(s.fd, syscall.SOL_SOCKET,
			syscall.SO_SNDBUF)
		if err != nil {
			return
		} else if v < s.TxBytes {
			err = fmt.Errorf(`
SO_SNDBUF truncated to %d bytes; run: sysctl -w net.core.wmem_max=%d`[1:],
				v, s.TxBytes)
			return
		}
	}

	if !s.DontListenAllNsid {
		err = os.NewSyscallError("setsockopt NETLINK_LISTEN_ALL_NSID",
			syscall.SetsockoptInt(s.fd, SOL_NETLINK,
				NETLINK_LISTEN_ALL_NSID, 1))
		if err != nil {
			return
		}
	}

	go s.gorx()
	go s.gotx()

	return
}

func (s *Socket) Close() (err error) {
	if s.Tx != nil {
		close(s.Tx)
	}
	// Close write side of pipe and wake up sleeping gorx epoll wait.
	// gorx will close read side of pipe.
	syscall.Close(s.rx_close_kludge_pipe[1])
	s.rx_close_kludge_pipe[1] = -1
	return
}

func (s *Socket) GetlinkReq() {
	req := NewGenMessage()
	req.Type = RTM_GETLINK
	req.Flags = NLM_F_REQUEST | NLM_F_MATCH
	req.AddressFamily = AF_UNSPEC
	s.Tx <- req
}

// The Listen handler is for messages that we receive while waiting for the
// DONE or ERROR acknowledgement to each dump request.
func (s *Socket) Listen(handler Handler, reqs ...ListenReq) (err error) {
	if len(reqs) == 0 {
		reqs = DefaultListenReqs
	}
	for _, r := range reqs {
		if r.MsgType == NLMSG_NOOP {
			continue
		}
		for tries := 1; true; tries++ {
			msg := NewGenMessage()
			msg.Type = r.MsgType
			msg.Flags = NLM_F_REQUEST | NLM_F_DUMP
			msg.AddressFamily = r.AddressFamily
			s.Tx <- msg
			err := s.RxUntilDone(handler)
			if err == nil {
				break
			}
			if tries >= 5 {
				return err
			}
		}
	}
	return nil
}

func (s *Socket) RxUntilDone(handler Handler) (err error) {
	for msg := range s.Rx {
		switch msg.MsgType() {
		case NLMSG_ERROR:
			e := msg.(*ErrorMessage)
			if e.Errno != 0 {
				err = syscall.Errno(-e.Errno)
			}
			msg.Close()
			return
		case NLMSG_DONE:
			msg.Close()
			return
		default:
			err = handler(msg)
			if err != nil {
				return
			}
		}
	}
	return
}

func (s *Socket) gorx() {
	buf := make([]byte, 4*PageSize)
	oob := make([]byte, PageSize)

	hasNsid := func(scm syscall.SocketControlMessage) bool {
		return scm.Header.Level == SOL_NETLINK &&
			scm.Header.Type == NETLINK_LISTEN_ALL_NSID
	}
	getNsid := func(scm syscall.SocketControlMessage) int {
		return *(*int)(unsafe.Pointer(&scm.Data[0]))
	}

	epollFd, err := syscall.EpollCreate1(0)
	if err != nil {
		panic(os.NewSyscallError("epoll_create1", err))
	}
	defer syscall.Close(epollFd)

	err = syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD,
		s.fd, &syscall.EpollEvent{
			Events: syscall.EPOLLIN,
			Fd:     int32(s.fd),
		})
	if err != nil {
		panic(os.NewSyscallError("epoll_ctl add", err))
	}
	err = syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD,
		s.rx_close_kludge_pipe[0], &syscall.EpollEvent{
			Events: syscall.EPOLLIN,
			Fd:     int32(s.rx_close_kludge_pipe[0]),
		})
	if err != nil {
		panic(os.NewSyscallError("epoll_ctl add rx_close_pipe", err))
	}

	s.wg.Add(1)

	var einval_retry int
	for {
		// Wait for input to be ready.
		// Need to poll since otherwise Recvmsg would block forever.
		// Close of s.fd does not wake/unblock Recvmsg.
		var ee [1]syscall.EpollEvent
		if _, err := syscall.EpollWait(epollFd, ee[:], -1); err != nil {
			if e, ok := err.(syscall.Errno); ok && e.Temporary() {
				continue
			}
			err = os.NewSyscallError("epoll_wait", err)
			fmt.Fprintln(os.Stderr, "Recv:", err)
			break
		}
		// Socket was closed?
		if ee[0].Fd == int32(s.rx_close_kludge_pipe[0]) {
			break
		}
		n, noob, _, _, err := syscall.Recvmsg(s.fd, buf, oob, syscall.MSG_DONTWAIT)
		if err != nil {
			// Re-try after timeout or interrupt unless socket was closed.
			if e, ok := err.(syscall.Errno); ok && e.Temporary() {
				continue
			} else {
				// EINVAL can happen when socket is being closed (not sure why); no need for noise in that case.
				// Retry 3 times before giving up
				einval_retry++
				if e != syscall.EINVAL {
					fmt.Fprintln(os.Stderr, "Recv:", err)
				}
				if einval_retry > 3 {
					if false {
						fmt.Println("gorx() EINVAL - socket closing!")
					}
					break
				} else {
					if false {
						fmt.Println("gorx() EINVAL - retry!")
					}
					continue
				}
			}
		}

		nsid := DefaultNsid
		if noob > 0 {
			scms, err :=
				syscall.ParseSocketControlMessage(oob[:noob])
			if err != nil {
				panic(err)
			}
			for _, scm := range scms {
				if hasNsid(scm) {
					nsid = getNsid(scm)
				}
			}
		}

		for i, l := 0, 0; i < n; i += l {
			if i+SizeofHeader > n {
				panic("incomplete header")
			}
			h := (*Header)(unsafe.Pointer(&buf[i]))
			l = h.MsgLen()
			var msg Message
			switch h.Type {
			case NLMSG_NOOP:
				msg = NewNoopMessage()
			case NLMSG_ERROR:
				msg = NewErrorMessage()
			case NLMSG_DONE:
				msg = NewDoneMessage()
			case RTM_GETLINK:
				msg = NewGenMessage()
			case RTM_NEWLINK, RTM_DELLINK, RTM_SETLINK:
				msg = NewIfInfoMessage()
			case RTM_NEWADDR, RTM_DELADDR, RTM_GETADDR:
				msg = NewIfAddrMessage()
			case RTM_NEWROUTE, RTM_DELROUTE, RTM_GETROUTE:
				msg = NewRouteMessage()
			case RTM_NEWNEIGH, RTM_DELNEIGH, RTM_GETNEIGH:
				msg = NewNeighborMessage()
			case RTM_NEWNSID, RTM_DELNSID, RTM_GETNSID:
				msg = NewNetnsMessage()
			}
			if msg == nil {
				continue
			}
			_, err = msg.Write(buf[i : i+int(h.Len)])
			if err != nil {
				errno, ok := err.(syscall.Errno)
				if !ok {
					errno = syscall.EINVAL
				}
				msg.Close()
				e := NewErrorMessage()
				e.Errormsg.Errno = -int32(errno)
				e.Errormsg.Req = *h
				msg = e
			}
			*msg.Nsid() = nsid
			if false {
				fmt.Print("Rx: ", msg)
			}
			s.rx <- msg
		}
	}
	close(s.rx)
	s.rx = nil

	// Close read side of pipe (write side is closed in Close method).
	syscall.Close(s.rx_close_kludge_pipe[0])
	s.rx_close_kludge_pipe[0] = -1

	s.wg.Done()

	// Wait for tx routine to finish before closing file descriptor.
	s.wg.Wait()
	syscall.Close(s.fd)
	s.fd = -1
}

func (s *Socket) gotx() {
	seq := uint32(1)
	buf := make([]byte, 4*PageSize)
	bh := (*Header)(unsafe.Pointer(&buf[0]))
	s.wg.Add(1)
	for msg := range s.tx {
		mh := msg.MsgHeader()
		if mh.Flags == 0 {
			mh.Flags = NLM_F_REQUEST
		}
		if mh.Pid == 0 {
			mh.Pid = s.addr.Pid
		}
		if mh.Sequence == 0 {
			mh.Sequence = seq
			seq++
		}
		n, err := msg.Read(buf)
		if err != nil {
			goto emsg
		}
		mh.Len = uint32(n)
		bh.Len = mh.Len
		if false {
			fmt.Print("Tx: ", msg)
		}
		_, err = syscall.Write(s.fd, buf[:n])
		if err != nil {
			goto emsg
		}
		msg.Close()
		continue
	emsg:
		errno, ok := err.(syscall.Errno)
		if !ok {
			errno = syscall.EINVAL
		}
		e := NewErrorMessage()
		e.Errormsg.Errno = -int32(errno)
		e.Errormsg.Req = *mh
		s.rx <- e
		msg.Close()
	}
	s.wg.Done()
}
