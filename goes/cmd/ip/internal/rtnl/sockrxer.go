// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"os"
	"path/filepath"
	"syscall"
)

// Usage: NewSockReceiver(sock[, dontwait bool])
func NewSockReceiver(sock *Sock, opts ...interface{}) *SockReceiver {
	var dontwait bool

	if len(opts) > 0 {
		dontwait = opts[0].(bool)
	}

	return &SockReceiver{sock, Empty, dontwait}
}

// A SockReceiver wraps a Sock to pop packed received messages.
type SockReceiver struct {
	*Sock
	b []byte

	dontwait bool
}

// Pop next message from packed buffer received from channel.
func (sr *SockReceiver) Recv() (msg []byte, err error) {
	msg = Empty
	for len(sr.b) < SizeofHdr {
		var opened bool
		if sr.dontwait && len(sr.Sock.RxCh) == 0 {
			return
		}
		if sr.b, opened = <-sr.Sock.RxCh; !opened {
			err = sr.Sock.Err
			return
		}
	}
	msg, sr.b, err = Pop(sr.b)
	return
}

// After setting the sequence number, send the request to netlink and call the
// given handler for each received message until DONE or ERROR.
func (sr *SockReceiver) UntilDone(req []byte, do func([]byte)) error {
	seq := Seq()
	HdrPtr(req).Seq = seq
	if err := sr.Sock.Send(req); err != nil {
		return err
	}
	for {
		b, err := sr.Recv()
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
		if sr.Sock.Pid != h.Pid {
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
			return syscall.Errno(-e.Errno)
		default:
			do(b)
		}
	}
	return nil
}

func (sr *SockReceiver) IfNamesByIndex() (map[int32]string, error) {
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
	} else if err = sr.UntilDone(req, func(b []byte) {
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

// Send a RTM_GETNSID request to netlink and return the response attribute.
func (sr *SockReceiver) Nsid(name string) (int32, error) {
	nsid := int32(-1)

	f, err := os.Open(filepath.Join(VarRunNetns, name))
	if err != nil {
		return -1, err
	}
	defer f.Close()

	if req, err := NewMessage(Hdr{
		Type:  RTM_GETNSID,
		Flags: NLM_F_REQUEST | NLM_F_ACK,
	}, NetnsMsg{
		Family: AF_UNSPEC,
	},
		Attr{NETNSA_FD, Uint32Attr(f.Fd())},
	); err != nil {
		return -1, err
	} else if err = sr.UntilDone(req, func(b []byte) {
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
	}); err != nil {
		return -1, err
	}
	return nsid, nil
}
