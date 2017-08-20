// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package socket

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/iomux"

	"fmt"
	"sync"
	"syscall"
)

type socket struct {
	iomux.File

	flags Flags

	maxReadBytes uint

	txBufLock          sync.Mutex
	txBuffer, RxBuffer elib.ByteVec

	SelfAddr, PeerAddr syscall.Sockaddr
}

type Server struct {
	socket
}
type Client struct {
	socket
}

type Flags uint32

const (
	// Listen (server side) rather than connect (client side)
	Listen Flags = 1 << iota
	// Client accepted by server.
	AcceptedClient
	// Non-blocking connect in progress.
	ConnectInProgress
	// Use UDP instead of TCP
	UDP
	TCPDelay // set to enable Nagle algorithm (default is disabled)
	Closed
)

func tst(err error, tag string) error {
	if err != nil {
		err = fmt.Errorf("%s %s", tag, err)
	}
	return err
}

func (s *socket) Close() (err error) {
	s.txBufLock.Lock()
	defer s.txBufLock.Unlock()

	// For unix listen sockets remove socket file on close.
	if s.flags&Listen != 0 {
		switch v := s.SelfAddr.(type) {
		case *syscall.SockaddrUnix:
			syscall.Unlink(v.Name)
		}
	}

	err = syscall.Close(s.Fd)
	if err != nil {
		err = fmt.Errorf("close: %s", err)
	}
	s.flags |= Closed
	return
}

func (s *socket) IsClosed() bool { return s.flags&Closed != 0 }

func (s *socket) WriteAvailable() bool { return len(s.txBuffer) > 0 || s.flags&ConnectInProgress != 0 }

func (s *socket) ReadReady() (err error) {
	i := len(s.RxBuffer)

	if s.maxReadBytes <= 0 {
		s.maxReadBytes = 4 << 10
	}
	s.RxBuffer.Resize(s.maxReadBytes)

	var n int
	n, err = syscall.Read(s.Fd, s.RxBuffer[i:])
	if err != nil {
		switch err {
		case syscall.EAGAIN:
			err = nil
			return
		}
		err = tst(err, "read")
		return
	}
	s.RxBuffer = s.RxBuffer[:i+n]

	if elog.Enabled() {
		s.elogData(Read, s.RxBuffer)
	}

	if n == 0 {
		iomux.Del(s)
		s.Close()
	}
	return
}

func (s *socket) Read(advance int) []byte {
	if advance >= len(s.RxBuffer) {
		s.RxBuffer = s.RxBuffer[:0]
	} else {
		s.RxBuffer = s.RxBuffer[advance:]
	}
	return s.RxBuffer
}

func (s *Server) AcceptClient(c *Client) (err error) {
	fd, sa, err := syscall.Accept(s.Fd)
	if err != nil {
		return
	}
	c.flags = AcceptedClient
	c.Fd = fd
	c.SelfAddr = s.SelfAddr
	c.PeerAddr = sa
	return
}

func (s *Server) ReadReady() (err error) {
	c := &Client{}
	err = s.AcceptClient(c)
	return
}

func (s *socket) ClientWriteReady() (newConnection bool, err error) {
	s.txBufLock.Lock()

	if s.IsClosed() {
		return
	}

	needUpdate := false
	defer func() {
		s.txBufLock.Unlock()
		if needUpdate {
			iomux.Update(s)
		}
	}()

	newConnection = s.flags&ConnectInProgress != 0
	if newConnection {
		s.flags &^= ConnectInProgress
		var errno int
		errno, err = syscall.GetsockoptInt(s.Fd, syscall.SOL_SOCKET, syscall.SO_ERROR)
		if err = tst(err, "getsockopt"); err != nil {
			return
		}
		if errno != 0 {
			err = fmt.Errorf("connect: %s", syscall.Errno(errno))
			return
		}
		// Update since connection in progress implies write available.
		needUpdate = true
	}

	if len(s.txBuffer) > 0 {
		var n int

		n, err = syscall.Write(s.Fd, s.txBuffer)
		if err != nil {
			err = tst(err, "write")
			return
		}
		l := len(s.txBuffer)
		elog.F("socket write #%d %d %d %x", s.File.Index(), n, l, s.txBuffer[0:n])
		switch {
		case n == l:
			s.txBuffer = s.txBuffer[:0]
		case n > 0:
			copy(s.txBuffer, s.txBuffer[n:])
			s.txBuffer = s.txBuffer[:l-n]
		}
		// Whole buffer written => toggle write available.
		needUpdate = true
	}
	return
}

func (s *socket) WriteReady() (err error) {
	_, err = s.ClientWriteReady()
	return
}

func (s *socket) Write(p []byte) (n int, err error) {
	s.txBufLock.Lock()
	defer s.txBufLock.Unlock()
	if s.IsClosed() {
		return
	}
	i := len(s.txBuffer)
	n = len(p)
	if n > 0 {
		s.txBuffer.Resize(uint(n))
		copy(s.txBuffer[i:i+n], p)
		iomux.Update(s)
	}
	return
}

func (s *socket) TxBuf() elib.ByteVec { return s.txBuffer }
func (s *socket) TxLen() int          { return len(s.txBuffer) }

func (s *socket) ErrorReady() (err error) {
	var v int
	v, err = syscall.GetsockoptInt(s.Fd, syscall.SOL_SOCKET,
		syscall.SO_ERROR)
	if err == nil && v != 0 {
		err = syscall.Errno(v)
	}
	return
}

/* Return and bind to an unused port. */
func (s *socket) bindFreePort(a []byte) (port int, err error) {
	// 5000 => IPPORT_USERRESERVED
	for port = 5000; port < 1<<16; port++ {
		switch len(a) {
		case 4:
			sa := &syscall.SockaddrInet4{Port: port}
			copy(sa.Addr[:], a)
			err = syscall.Bind(s.Fd, sa)
		case 16:
			sa := &syscall.SockaddrInet6{Port: port}
			copy(sa.Addr[:], a)
			err = syscall.Bind(s.Fd, sa)
		default:
			panic(a)
		}
		if err == nil {
			return
		}
	}

	err = fmt.Errorf("bind: reached maximum port")
	return
}

func (s *socket) Config(cfg string, flags Flags) (err error) {
	var sa syscall.Sockaddr

	/* Anything that begins with a / is a local Unix file socket. */
	af := syscall.AF_INET
	if len(cfg) > 0 && cfg[0] == '/' {
		sa = &syscall.SockaddrUnix{Name: cfg}
		af = syscall.AF_UNIX
	} else {
		var a Ip4Socket
		if _, err = fmt.Sscanf(cfg, "%s", &a); err == nil {
			sa = &syscall.SockaddrInet4{Addr: a.Address, Port: int(a.Port)}
		} else {
			err = fmt.Errorf("failed to parse config from `%s': %s", cfg, err)
			return
		}
	}

	// Sanitize flags.
	flags &= Listen | UDP | TCPDelay

	kind := syscall.SOCK_STREAM
	if flags&UDP != 0 {
		kind = syscall.SOCK_DGRAM
	}
	s.Fd, err = syscall.Socket(af, kind, 0)
	if err = tst(err, "socket"); err != nil {
		return
	}
	defer func() {
		if err != nil {
			s.Close()
		}
	}()

	if af != syscall.AF_UNIX {
		nodelay := 1
		if flags&TCPDelay != 0 {
			nodelay = 0
		}
		err = syscall.SetsockoptInt(s.Fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, nodelay)
		if err = tst(err, "setsockopt TCP_NODELAY"); err != nil {
			return
		}
	}

	if flags&Listen != 0 {
		needBind := true
		switch v := sa.(type) {
		case *syscall.SockaddrInet4:
			if IpPort(v.Port) == NilIpPort {
				v.Port, err = s.bindFreePort(v.Addr[:])
				if err != nil {
					return
				}
				needBind = false
			}
		case *syscall.SockaddrUnix:
			syscall.Unlink(v.Name)
		default:
			panic(v)
		}

		err = syscall.SetsockoptInt(s.Fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
		if err = tst(err, "setsockopt SO_REUSEADDR"); err != nil {
			return
		}

		if needBind {
			err = syscall.Bind(s.Fd, sa)
			if err = tst(err, "bind"); err != nil {
				return
			}
		}

		err = syscall.Listen(s.Fd, syscall.SOMAXCONN)
		if err = tst(err, "listen"); err != nil {
			return
		}

		s.SelfAddr = sa
	} else {
		s.PeerAddr = sa

		err = syscall.Connect(s.Fd, sa)
		if err = tst(err, "connect"); err != nil {
			return
		}

		s.SelfAddr, err = syscall.Getsockname(s.Fd)
		if err = tst(err, "getsockname"); err != nil {
			return
		}
		flags |= ConnectInProgress
	}

	s.flags = flags
	return
}

func NewServer(cfg string) (s *Server, err error) {
	s = &Server{}
	err = s.Config(cfg, Listen)
	return
}

func NewClient(cfg string) (s *Client, err error) {
	s = &Client{}
	err = s.Config(cfg, 0)
	return
}

func SockaddrString(a syscall.Sockaddr) string {
	switch v := a.(type) {
	case *syscall.SockaddrInet4:
		s := Ip4Socket{Address: v.Addr, Port: IpPort(v.Port)}
		return s.String()
	case *syscall.SockaddrUnix:
		return fmt.Sprintf("unix:%s", v.Name)
	default:
		panic(v)
	}
}

func (s *socket) String() string {
	return fmt.Sprintf("%s -> %s", SockaddrString(s.SelfAddr), SockaddrString(s.PeerAddr))
}

// Event logging.
type event struct {
	flags eventFlag
	s     [elog.EventDataBytes - 1]byte
}

//go:generate gentemplate -d Package=socket -id event -d Type=event github.com/platinasystems/go/elib/elog/event.tmpl

type eventFlag uint8

const (
	// low 4 bits are op code
	Read   eventFlag = 0
	Write  eventFlag = 1
	IsData eventFlag = 1 << (iota + 4)
)

var opNames = []string{
	Read:  "read",
	Write: "write",
}

func (s *socket) elogf(f eventFlag, format string, args ...interface{}) (e event) {
	e = event{flags: f}
	b := elog.PutUvarint(e.s[:], int(s.File.Index()))
	elog.Printf(b, format, args...)
	e.Log()
	return
}

func (s *socket) elogData(f eventFlag, p []byte) (e event) {
	e = event{flags: f | IsData}
	b := elog.PutUvarint(e.s[:], int(s.File.Index()))
	elog.PutData(b, p)
	e.Log()
	return
}

func (e *event) Strings(x *elog.Context) []string {
	op := opNames[e.flags&0xf]
	var d string
	b := e.s[:]
	b, fi := elog.Uvarint(b)
	if e.flags&IsData != 0 {
		d = elog.HexData(b)
	} else {
		d = elog.String(b)
	}
	return []string{fmt.Sprintf("socket #%d %s %s", fi, op, d)}
}

func (e *event) Encode(x *elog.Context, b []byte) int {
	b[0] = byte(e.flags)
	copy(b[1:], e.s[:])
	return 1 + len(e.s)
}
func (e *event) Decode(x *elog.Context, b []byte) int {
	e.flags = eventFlag(b[0])
	copy(e.s[:], b[1:])
	return 1 + len(e.s)
}
