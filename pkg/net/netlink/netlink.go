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

// This returns references to the next netlink message header and data that the
// caller must release before subsequent Next calls.
func (nl *Netlink) Next() (*NlMsghdr, []byte, error) {
	if len(nl.rem) < NLMSG_HDRLEN {
		for {
			n, _, err := af.Recvfrom(nl.sock, nl.buf, MSG_PEEK)
			if err != nil {
				return nil, nil, egress.Marked(err)
			}
			if n < len(nl.buf) {
				break
			} else if n < cap(nl.buf) {
				nl.buf = nl.buf[:cap(nl.buf)]
			} else {
				nl.buf = make([]byte, align.Page.Roundup(n))
			}
		}
		if n, _, err := af.Recvfrom(nl.sock, nl.buf, 0); err != nil {
			return nil, nil, egress.Marked(err)
		} else if n < NLMSG_HDRLEN {
			return nil, nil, egress.Marked(EINVAL)
		} else {
			nl.rem = nl.buf[:n]
		}
	}
	hdr := Pointer[NlMsghdr](nl.rem)
	n := int(hdr.Len)
	al := align.NLMSG.Roundup(n)
	if hdr.Len < NLMSG_HDRLEN {
		return nil, nil, egress.Marked(EINVAL)
	}
	if al > len(nl.rem) {
		return nil, nil, egress.Marked(EINVAL)
	}
	if hdr.Pid != nl.pid {
		return nil, nil, egress.Marked(EINVAL)
	}
	data := nl.rem[NLMSG_HDRLEN:n]
	nl.rem = nl.rem[al:]
	return hdr, data, nil
}

// Set message header's length and next sequence number before socket send.
func (nl *Netlink) Request(msg []byte) error {
	req := Pointer[NlMsghdr](msg)
	req.Seq = nl.seq.Add(1)
	req.Len = uint32(len(msg))
	return af.Sendto(nl.sock, msg, 0, nl.addr)
}

// Wait for DONE or ERROR response to the identified request.
func (nl *Netlink) Wait(seq uint32) error {
	for {
		hdr, data, err := nl.Next()
		if err != nil {
			return err
		} else if hdr.Seq != seq {
			continue
		} else if hdr.Type == NLMSG_DONE {
			return nil
		} else if hdr.Type == NLMSG_ERROR {
			return ExtractError(data)
		}
	}
}
