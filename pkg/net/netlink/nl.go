// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netlink

import (
	"context"
	"fmt"
	"sync/atomic"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/iflink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

const SOL_NETLINK = 270

type NL struct {
	sock af.Netlink
	addr syscall.Sockaddr
	seq  atomic.Uint32
	pid  uint32
	buf  []byte
	rem  []byte
}

// Open netlink socket with paired (option, value)s.
func Open(optval ...int) (nl *NL, err error) {
	sock, err := af.Open[af.Netlink]()
	if err = egress.Mark(err); err != nil {
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
		if err = egress.Mark(err); err != nil {
			return
		}

	}
	addr, err := sock.Bind()
	if err = egress.Mark(err); err != nil {
		return
	}
	lsa, err := af.Addr(sock)
	if err = egress.Mark(err); err != nil {
		return
	}
	if lsanl, ok := lsa.(*syscall.SockaddrNetlink); !ok {
		err = egress.Mark(ErrInvalid)
	} else {
		nl = &NL{
			sock: sock,
			addr: addr,
			pid:  lsanl.Pid,
			buf:  page.New(),
		}
	}
	return
}

func (nl *NL) Close() error {
	nl.buf = nl.buf[:0]
	nl.rem = nl.rem[:0]
	return af.Close(nl.sock)
}

func (nl *NL) IfIndex(ctx context.Context, ifname string) (int32, error) {
	hdr, req := ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_GETLINK
	hdr.Flags = NLM_F_REQUEST | NLM_F_DUMP
	gen, req := ExpandRtGenMsg(req)
	gen.Family = af.UNSPEC
	if err := nl.Request(req); err != nil {
		return -1, err
	}
	seq := hdr.SEQ
	for {
		rsp, data, err := nl.Next(ctx)
		if err != nil {
			return -1, err
		} else if rsp.SEQ != seq {
			continue
		} else if rsp.Type == NLMSG_DONE {
			break
		} else if rsp.Type == NLMSG_ERROR {
			e, _ := ExtractMsgErr(data)
			return -1, e.Err()
		} else if rsp.Type != rtnetlink.RTM_NEWLINK {
			continue
		}
		ifinfo, data := ExtractIfInfoMsg(data)
		ifindex := int32(-1)
		for HasAttr(data) {
			t, v, datá := ExtractAttr(data)
			data = datá
			if iflink.Ifla(t) == iflink.IFLA_IFNAME {
				if CloneString(v) == ifname {
					ifindex = ifinfo.Index
				}
			}
		}
		if ifindex > 0 {
			return ifindex, nil
		}
	}
	return -1, fmt.Errorf("%q %w", ifname, ErrNotFound)
}

// This returns references to the next netlink message header and data that the
// caller must release before subsequent Next calls.
func (nl *NL) Next(ctx context.Context) (*MsgHdr, []byte, error) {
	if len(nl.rem) < NLMSG_HDRLEN {
		for {
			n, _, err := af.
				Recvfrom(nl.sock, nl.buf, syscall.MSG_PEEK)
			if err != nil {
				return nil, nil, egress.Mark(err)
			}
			if n < len(nl.buf) {
				break
			} else if n < cap(nl.buf) {
				nl.buf = nl.buf[:cap(nl.buf)]
			} else {
				nl.buf = make([]byte, page.Align(n))
			}
		}
		if n, _, err := af.Recvfrom(nl.sock, nl.buf, 0); err != nil {
			return nil, nil, egress.Mark(err)
		} else if n < NLMSG_HDRLEN {
			return nil, nil, egress.Mark(ErrInvalid)
		} else {
			nl.rem = nl.buf[:n]
		}
	}
	hdr := Pointer[MsgHdr](nl.rem)
	n := int(hdr.Len)
	al := NLMSG_ALIGN(n)
	if hdr.Len < NLMSG_HDRLEN {
		return nil, nil, egress.Mark(ErrInvalid)
	}
	if al > len(nl.rem) {
		return nil, nil, egress.Mark(ErrInvalid)
	}
	if hdr.PID != nl.pid {
		return nil, nil, egress.Mark(ErrInvalid)
	}
	data := nl.rem[NLMSG_HDRLEN:n]
	nl.rem = nl.rem[al:]
	return hdr, data, nil
}

// Set message header's length and next sequence number before socket send.
func (nl *NL) Request(msg []byte) error {
	req := Pointer[MsgHdr](msg)
	req.SEQ = nl.seq.Add(1)
	req.Len = uint32(len(msg))
	return af.Sendto(nl.sock, msg, 0, nl.addr)
}

// Wait for DONE or ERROR response to the identified request.
func (nl *NL) Wait(ctx context.Context, seq uint32) error {
	for {
		hdr, data, err := nl.Next(ctx)
		if err != nil {
			return err
		} else if hdr.SEQ != seq {
			continue
		} else if hdr.Type == NLMSG_DONE {
			return nil
		} else if hdr.Type == NLMSG_ERROR {
			m, _ := ExtractMsgErr(data)
			return m.Err()
		}
	}
}
