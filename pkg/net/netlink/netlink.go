// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netlink

import (
	"sync/atomic"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
	"github.com/platinasystems/goes/v2/pkg/syscall/align"
)

type Netlink struct {
	sock af.Netlink
	addr Sockaddr
	seq  atomic.Uint32
	pid  uint32
	buf  []byte
	rem  []byte
}

// Open netlink socket with paired (option, value)s.
func Open(optval ...int) (nl *Netlink, err error) {
	sock, err := af.Open[af.Netlink]()
	if err = egress.Marked(err); err != nil {
		return
	}
	defer func() {
		if err != nil {
			af.Close(sock)
		}
	}()
	for ; len(optval) >= 2; optval = optval[2:] {
		opt, val := optval[0], optval[1]
		err = syscall.SetsockoptInt(int(sock), SOL_NETLINK, opt, val)
		if err = egress.Marked(err); err != nil {
			return
		}

	}
	addr, err := sock.Bind()
	if err = egress.Marked(err); err != nil {
		return
	}
	lsa, err := af.Addr(sock)
	if err = egress.Marked(err); err != nil {
		return
	}
	if lsanl, ok := lsa.(*SockaddrNetlink); !ok {
		err = egress.Marked(EINVAL)
	} else {
		nl = &Netlink{
			sock: sock,
			addr: addr,
			pid:  lsanl.Pid,
			buf:  make([]byte, align.Page.Size()),
		}
	}
	return
}

func (nl *Netlink) Close() error {
	nl.buf = nl.buf[:0]
	nl.rem = nl.rem[:0]
	return af.Close(nl.sock)
}

// Set request's header length and next sequence number before writing it to
// socket and processing responses until error or done.  The processing is
// skipped if `f` is nil.
func (nl *Netlink) Request(msg []byte, f func([]byte) error) error {
	req := Pointer[NlMsghdr](msg)
	req.Seq = nl.seq.Add(1)
	req.Len = uint32(len(msg))
	if err := af.Sendto(nl.sock, msg, 0, nl.addr); err != nil {
		return err
	}
	for {
		data, err := nl.next()
		if err != nil {
			return err
		}
		h, msg := Extract[NlMsghdr](data)
		if h.Seq != req.Seq {
			continue
		}
		if h.Type == NLMSG_DONE {
			break
		}
		if h.Type == NLMSG_ERROR {
			msgerr, _ := Extract[NlMsgerr](msg)
			if msgerr.Error != 0 {
				return egress.Marked(Errno(-msgerr.Error))
			}
			return nil
		}
		if f == nil {
			continue
		}
		if err = f(data); err != nil {
			return err
		}
	}
	return nil
}

func (nl *Netlink) next() ([]byte, error) {
	if len(nl.rem) < NLMSG_HDRLEN {
		for {
			n, from, err := af.Recvfrom(nl.sock, nl.buf, MSG_PEEK)
			_ = from
			if err != nil {
				return nil, egress.Marked(err)
			}
			if n < len(nl.buf) {
				break
			} else if n < cap(nl.buf) {
				nl.buf = nl.buf[:cap(nl.buf)]
			} else {
				nl.buf = make([]byte, align.Page.Roundup(n))
			}
		}
		if n, from, err := af.Recvfrom(nl.sock, nl.buf, 0); err != nil {
			_ = from
			return nil, egress.Marked(err)
		} else if n < NLMSG_HDRLEN {
			return nil, egress.Marked(EINVAL)
		} else {
			nl.rem = nl.buf[:n]
		}
	}
	h := Pointer[NlMsghdr](nl.rem)
	n := int(h.Len)
	al := align.NLMSG.Roundup(n)
	if h.Len < NLMSG_HDRLEN {
		return nil, egress.Marked(EINVAL)
	}
	if al > len(nl.rem) {
		return nil, egress.Marked(EINVAL)
	}
	if h.Pid != nl.pid {
		return nil, egress.Marked(EINVAL)
	}
	msg := nl.rem[:n]
	nl.rem = nl.rem[al:]
	return msg, nil
}
