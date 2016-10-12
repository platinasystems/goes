// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package netlink

import (
	"io"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/elib"
)

var DefaultGroups = []MulticastGroup{
	RTNLGRP_LINK,
	RTNLGRP_NEIGH,
	RTNLGRP_IPV4_IFADDR,
	RTNLGRP_IPV4_ROUTE,
	RTNLGRP_IPV4_MROUTE,
	RTNLGRP_IPV6_IFADDR,
	RTNLGRP_IPV6_ROUTE,
	RTNLGRP_IPV6_MROUTE,
}

type Socket struct {
	socket             int
	pid                uint32
	tx_sequence_number uint
	tx_buffer          elib.ByteVec
	rx_buffer          elib.ByteVec
	rx_chan            chan Message
	quit_chan          chan struct{}
	sync.Mutex
	rsvp map[uint32]chan *ErrorMessage
}

func New(rx chan Message, groups ...MulticastGroup) (s *Socket, err error) {
	s = &Socket{
		rx_chan:   rx,
		quit_chan: make(chan struct{}),
	}
	s.socket, err = syscall.Socket(syscall.AF_NETLINK, syscall.SOCK_RAW, syscall.NETLINK_ROUTE)
	if err != nil {
		err = os.NewSyscallError("socket", err)
		return
	}
	defer func() {
		if err != nil && s.socket > 0 {
			syscall.Close(s.socket)
		}
	}()

	var groupbits uint32
	if len(groups) == 0 {
		groups = DefaultGroups
	}
	for _, group := range groups {
		if group != NOOP_RTNLGRP {
			groupbits |= 1 << group
		}
	}

	sa := &syscall.SockaddrNetlink{
		Family: uint16(AF_NETLINK),
		Pid:    s.pid,
		Groups: groupbits,
	}

	if err = syscall.Bind(s.socket, sa); err != nil {
		err = os.NewSyscallError("bind", err)
		return
	}

	// Increase socket buffering.
	bytes := 1024 << 10
	if err = os.NewSyscallError("setsockopt SO_RCVBUF", syscall.SetsockoptInt(s.socket, syscall.SOL_SOCKET, syscall.SO_RCVBUF, bytes)); err != nil {
		return
	}
	if err = os.NewSyscallError("setsockopt SO_SNDBUF", syscall.SetsockoptInt(s.socket, syscall.SOL_SOCKET, syscall.SO_SNDBUF, bytes)); err != nil {
		return
	}
	return
}

func (s *Socket) Close() error {
	close(s.quit_chan)
	s.Lock()
	defer s.Unlock()
	for k, ch := range s.rsvp {
		close(ch)
		delete(s.rsvp, k)
	}
	s.rsvp = nil
	return nil
}

func (s *Socket) Listen(reqs ...ListenReq) {
	if len(reqs) == 0 {
		reqs = DefaultListenReqs
	}
	for _, r := range reqs {
		if r.MsgType == NLMSG_NOOP {
			continue
		}
		m := pool.GenMessage.Get().(*GenMessage)
		m.Type = r.MsgType
		m.Flags = NLM_F_DUMP
		m.AddressFamily = r.AddressFamily
		s.Tx(m)
		s.rxUntilDone()
	}

	for {
		select {
		case _ = <-s.quit_chan:
			syscall.Close(s.socket)
			s.socket = -1
			close(s.rx_chan)
			s.rx_chan = nil
			s.quit_chan = nil
			return
		default:
		}
		s.rx()
	}
}

func (s *Socket) Rsvp(m Message) chan *ErrorMessage {
	var hp *Header
	s.Lock()
	defer s.Unlock()
	ch := make(chan *ErrorMessage, 1)
	switch t := m.(type) {
	case *IfInfoMessage:
		hp = &t.Header
	case *IfAddrMessage:
		hp = &t.Header
	case *RouteMessage:
		hp = &t.Header
	case *NeighborMessage:
		hp = &t.Header
	case *NetnsMessage:
		hp = &t.Header
	default:
		panic("unsupported netlink message type")
	}
	s.TxAdd(m)
	if s.rsvp == nil {
		s.rsvp = make(map[uint32]chan *ErrorMessage)
	}
	s.rsvp[hp.Sequence] = ch
	s.TxFlush()
	return ch
}

func (s *Socket) Rx() (Message, error) {
	if s.rx_chan != nil {
		if msg, opened := <-s.rx_chan; opened {
			return msg, nil
		}
	}
	return nil, io.EOF
}

func (s *Socket) Tx(m Message) {
	s.TxAdd(m)
	s.TxFlush()
}

func (s *Socket) TxAdd(m Message) { m.TxAdd(s) }

// txAdd adds a both a nlmsghdr and a request header (e.g. ifinfomsg, ifaddrmsg, rtmsg, ...)
//   to the end of the tx buffer.
func (s *Socket) TxAddReq(header *Header, nBytes int) []byte {
	i := len(s.tx_buffer)
	s.tx_buffer.Resize(uint(messageAlignLen(nBytes) + SizeofHeader))
	h := (*Header)(unsafe.Pointer(&s.tx_buffer[i]))
	h.Len = uint32(nBytes + SizeofHeader)
	h.Type = header.Type
	h.Flags = header.Flags | NLM_F_REQUEST
	h.Pid = s.pid
	header.Pid = s.pid

	// Sequence 0 is reserved for unsolicited messages from kernel.
	if header.Sequence == 0 {
		if s.tx_sequence_number == 0 {
			s.tx_sequence_number = 1
		}
		h.Sequence = uint32(s.tx_sequence_number)
		header.Sequence = uint32(s.tx_sequence_number)
		s.tx_sequence_number++
	}

	return s.tx_buffer[i:]
}

func (s *Socket) TxFlush() {
	for i := 0; i < len(s.tx_buffer); {
		n, err := syscall.Write(s.socket, s.tx_buffer[i:])
		if err != nil {
			panic(err)
		}
		i += n
	}
	s.reset_tx_buffer()
}

func (s *Socket) fillRxBuffer() {
	i := len(s.rx_buffer)
	s.rx_buffer.Resize(4096)
	m, err := syscall.Read(s.socket, s.rx_buffer[i:])
	if err != nil {
		panic(err)
	}
	s.rx_buffer = s.rx_buffer[:i+m]
}

func (n *Socket) reset_tx_buffer() {
	if len(n.tx_buffer) != 0 {
		n.tx_buffer = n.tx_buffer[:0]
	}
}

func (s *Socket) rx() (done bool) {
	s.fillRxBuffer()
	i := 0
	for {
		q := len(s.rx_buffer)
		// Have at least a valid message header in buffer?
		if i+SizeofHeader > q {
			s.rx_buffer = s.rx_buffer[:q-i]
			break
		}
		// Have a full message in recieve buffer?
		h := (*Header)(unsafe.Pointer(&s.rx_buffer[i]))
		l := messageAlignLen(int(h.Len))
		if i+l > q {
			if i == len(s.rx_buffer) {
				s.rx_buffer = s.rx_buffer[:0]
			} else {
				copy(s.rx_buffer, s.rx_buffer[i:])
				s.rx_buffer = s.rx_buffer[:q-i]
			}
			break
		}

		done = h.Type == NLMSG_DONE
		s.rxDispatch(h, s.rx_buffer[i:i+int(h.Len)])
		i += l
	}
	return
}

func (s *Socket) rxDispatch(h *Header, msg []byte) {
	var m Message
	var errMsg *ErrorMessage
	switch h.Type {
	case NLMSG_NOOP:
		m = NewNoopMessageBytes(msg)
	case NLMSG_ERROR:
		errMsg = NewErrorMessageBytes(msg)
		m = errMsg
	case NLMSG_DONE:
		m = NewDoneMessageBytes(msg)
	case RTM_NEWLINK, RTM_DELLINK, RTM_GETLINK, RTM_SETLINK:
		m = NewIfInfoMessageBytes(msg)
	case RTM_NEWADDR, RTM_DELADDR, RTM_GETADDR:
		m = NewIfAddrMessageBytes(msg)
	case RTM_NEWROUTE, RTM_DELROUTE, RTM_GETROUTE:
		m = NewRouteMessageBytes(msg)
	case RTM_NEWNEIGH, RTM_DELNEIGH, RTM_GETNEIGH:
		m = NewNeighborMessageBytes(msg)
	case RTM_NEWNSID, RTM_DELNSID, RTM_GETNSID:
		m = NewNetnsMessageBytes(msg)
	default:
		panic("unhandled message " + h.Type.String())
	}
	if errMsg != nil && errMsg.Req.Pid == s.pid {
		s.Lock()
		defer s.Unlock()
		if s.rsvp != nil {
			ch, found := s.rsvp[errMsg.Req.Sequence]
			if found {
				ch <- errMsg
				close(ch)
				delete(s.rsvp, errMsg.Req.Sequence)
				return
			}
		}
	}
	if s.rx_chan != nil {
		s.rx_chan <- m
	}
}

func (s *Socket) rxUntilDone() {
	for !s.rx() {
	}
}

type ListenReq struct {
	MsgType
	AddressFamily
}

var DefaultListenReqs = []ListenReq{
	{RTM_GETLINK, AF_PACKET},
	{RTM_GETADDR, AF_INET},
	{RTM_GETROUTE, AF_INET},
	{RTM_GETNEIGH, AF_INET},
	{RTM_GETADDR, AF_INET6},
	{RTM_GETNEIGH, AF_INET6},
	{RTM_GETROUTE, AF_INET6},
}

var NoopListenReq = ListenReq{NLMSG_NOOP, AF_UNSPEC}
