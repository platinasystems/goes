// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netlink

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"

	"github.com/platinasystems/goes/v2/pkg/netlink/iflink"
	"github.com/platinasystems/goes/v2/pkg/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xos"
	"golang.org/x/sys/unix"
)

const SOL_NETLINK = 270

var BufAlign = xnet.Align(NLBUF_ALIGNTO).Roundup

type NL struct {
	sock xnet.Netlink
	addr unix.Sockaddr
	seq  atomic.Uint32
	pid  uint32
	buf  []byte
	rem  []byte
}

// Open netlink socket with paired (option, value)s.
func Open(optval ...int) (nl *NL, err error) {
	sock, err := xnet.OpenNetlink()
	if err = xerrors.Mark(err); err != nil {
		return
	}
	defer func() {
		if err != nil {
			xnet.Close(sock)
		}
	}()
	for ; len(optval) >= 2; optval = optval[2:] {
		opt, val := optval[0], optval[1]
		err = unix.SetsockoptInt(int(sock), SOL_NETLINK, opt, val)
		if err = xerrors.Mark(err); err != nil {
			return
		}

	}
	addr, err := sock.Bind()
	if err = xerrors.Mark(err); err != nil {
		return
	}
	lsa, err := xnet.Addr(sock)
	if err = xerrors.Mark(err); err != nil {
		return
	}
	if lsanl, ok := lsa.(*unix.SockaddrNetlink); !ok {
		err = xerrors.Invalid("sock-type")
	} else {
		pgsz := os.Getpagesize()
		nl = &NL{
			sock: sock,
			addr: addr,
			pid:  lsanl.Pid,
			buf:  make([]byte, pgsz, pgsz),
		}
	}
	return
}

func (nl *NL) Close() error {
	nl.buf = nl.buf[:0]
	nl.rem = nl.rem[:0]
	return xnet.Close(nl.sock)
}

func (nl *NL) IfIndex(ctx context.Context, ifname string) (int32, error) {
	hdr, req := ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_GETLINK
	hdr.Flags = NLM_F_REQUEST | NLM_F_DUMP
	gen, req := ExpandRtGenMsg(req)
	gen.Family = xnet.AF_UNSPEC
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
			if t == iflink.IFLA_IFNAME {
				if CloneString(v) == ifname {
					ifindex = ifinfo.Index
				}
			}
			data = datá
		}
		if ifindex > 0 {
			return ifindex, nil
		}
	}
	return -1, fmt.Errorf("%q %w", ifname, ErrNotFound)
}

func IsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// This returns references to the next netlink message header and data that the
// caller must release before subsequent Next calls.
func (nl *NL) Next(ctx context.Context) (*MsgHdr, []byte, error) {
	const peek = unix.MSG_PEEK | unix.MSG_TRUNC | unix.MSG_DONTWAIT
	const donotwait = unix.MSG_DONTWAIT
	fd := int(nl.sock)
	if len(nl.rem) < NLMSG_HDRLEN {
		if IsDone(ctx) {
			return nil, nil, ctx.Err()
		}
		for {
			var sel xos.Selection
			sel.Read.Set(fd)
			_, err := sel.Select()
			if IsDone(ctx) {
				return nil, nil, ctx.Err()
			}
			if err != nil {
				if errors.Is(err, unix.EAGAIN) {
					continue
				}
				return nil, nil, xerrors.Mark(err)
			}
			if !sel.Read.IsSet(fd) {
				continue
			}
			// peek to see if we need to expand buffer
			n, _, err := xnet.Recvfrom(nl.sock, nl.buf, peek)
			if err != nil {
				return nil, nil, xerrors.Mark(err)
			}
			if n < len(nl.buf) {
				break
			} else if n < cap(nl.buf) {
				nl.buf = nl.buf[:cap(nl.buf)]
			} else {
				nl.buf = make([]byte, BufAlign(n))
			}
		}
		n, _, err := xnet.Recvfrom(nl.sock, nl.buf, donotwait)
		if err != nil {
			return nil, nil, xerrors.Mark(err)
		} else if n < NLMSG_HDRLEN {
			return nil, nil, xerrors.Invalid("rcv-count")
		} else {
			nl.rem = nl.buf[:n]
		}
	}
	hdr := Pointer[MsgHdr](nl.rem)
	n := int(hdr.Len)
	al := NLMSG_ALIGN(n)
	if hdr.Len < NLMSG_HDRLEN {
		return nil, nil, xerrors.Invalid("header-len")
	}
	if al > len(nl.rem) {
		return nil, nil, xerrors.Invalid("align-len")
	}
	if hdr.PID != nl.pid {
		return nil, nil, xerrors.Invalid("pid")
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
	return xnet.Sendto(nl.sock, msg, 0, nl.addr)
}

func (nl *NL) SetOpt(opt, val int) error {
	return unix.SetsockoptInt(int(nl.sock), SOL_NETLINK, opt, val)
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
