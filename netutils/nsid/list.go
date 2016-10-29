// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package nsid

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/platinasystems/go/netlink"
)

func (nsid *nsid) List() ([]Entry, error) {
	pid := uint32(os.Getpid())
	dir, err := ioutil.ReadDir(VarRunNetns)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, len(dir))
	stop := make(chan struct{})
	nlch := make(chan netlink.Message, 64)
	nlsock, err := netlink.New(nlch, netlink.NOOP_RTNLGRP)
	if err != nil {
		return nil, err
	}
	go func(sock *netlink.Socket, wait <-chan struct{}) {
		defer sock.Close()
		<-wait
	}(nlsock, stop)
	go nlsock.Listen(netlink.NoopListenReq)
	defer close(stop)
	for i, info := range dir {
		name := info.Name()
		entries[i].Name = name
		f, err := os.Open(filepath.Join(VarRunNetns, name))
		if err != nil {
			return nil, err
		}
		req := netlink.NewNetnsMessage()
		req.Type = netlink.RTM_GETNSID
		req.Flags = netlink.NLM_F_REQUEST
		req.Sequence = atomic.AddUint32(&nsid.seq, 1)
		req.Pid = pid
		req.AddressFamily = netlink.AF_UNSPEC
		req.Attrs[netlink.NETNSA_FD] = netlink.Uint32Attr(f.Fd())
		req.TxAdd(nlsock)
		nlsock.TxFlush()
	rx:
		for msg := range nlch {
			switch t := msg.(type) {
			case *netlink.NetnsMessage:
				rpid := t.Pid
				nsid := t.NSID()
				t.Close()
				if rpid == pid {
					entries[i].Pid = rpid
					entries[i].Id = nsid
					break rx
				}
			default:
				err = fmt.Errorf("unexpected: %s", msg)
				msg.Close()
				return nil, err
			}
		}
		f.Close()
	}
	return entries, nil
}
