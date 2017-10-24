// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nl

import "syscall"

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

func DoNothing([]byte) {}
