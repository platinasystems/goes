// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nsid

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"syscall"

	"github.com/platinasystems/go/internal/netlink"
)

func (nsid *nsid) Set(name string, id int32) error {
	return nsid.setunset(name, id, netlink.RTM_NEWNSID)
}

func (nsid *nsid) Unset(name string, id int32) error {
	return nsid.setunset(name, id, netlink.RTM_DELNSID)
}

func (nsid *nsid) setunset(name string, id int32, mt netlink.MsgType) error {
	f, err := os.Open(filepath.Join(VarRunNetns, name))
	if err != nil {
		return err
	}
	defer f.Close()
	pid := uint32(os.Getpid())
	stop := make(chan struct{})
	nlch := make(chan netlink.Message, 64)
	nlsock, err := netlink.New(nlch, netlink.NOOP_RTNLGRP)
	if err != nil {
		return err
	}
	go func(sock *netlink.Socket, wait <-chan struct{}) {
		defer sock.Close()
		<-wait
	}(nlsock, stop)
	go nlsock.Listen(netlink.NoopListenReq)
	defer close(stop)
	req := netlink.NewNetnsMessage()
	req.Type = mt
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	req.Sequence = atomic.AddUint32(&nsid.seq, 1)
	req.Pid = pid
	req.AddressFamily = netlink.AF_UNSPEC
	req.Attrs[netlink.NETNSA_NSID] = netlink.Int32Attr(id)
	req.Attrs[netlink.NETNSA_FD] = netlink.Uint32Attr(f.Fd())
	req.TxAdd(nlsock)
	nlsock.TxFlush()
rx:
	for msg := range nlch {
		switch t := msg.(type) {
		case *netlink.ErrorMessage:
			if t.Errno != 0 {
				err = syscall.Errno(-t.Errno)
			}
			t.Close()
			break rx
		default:
			err = fmt.Errorf("unexpected: %s", msg)
			msg.Close()
			break rx
		}
	}
	return err
}
