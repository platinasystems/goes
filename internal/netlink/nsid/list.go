// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nsid

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/internal/netlink"
)

type Entry struct {
	Name string
	Nsid int32
}

func List(args ...string) ([]Entry, error) {
	if len(args) == 0 {
		dir, err := ioutil.ReadDir(VarRunNetns)
		if err != nil {
			return nil, err
		}
		for _, info := range dir {
			args = append(args, info.Name())
		}
	}
	nl, err := netlink.New(netlink.NOOP_RTNLGRP)
	if err != nil {
		return nil, err
	}
	defer nl.Close()
	entries := make([]Entry, len(args))
	for i, arg := range args {
		var f *os.File
		entries[i].Name = arg
		f, err = os.Open(filepath.Join(VarRunNetns, arg))
		if err != nil {
			break
		}
		req := netlink.NewNetnsMessage()
		req.Type = netlink.RTM_GETNSID
		req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
		req.AddressFamily = netlink.AF_UNSPEC
		req.Attrs[netlink.NETNSA_FD] = netlink.Uint32Attr(f.Fd())
		nl.Tx <- req
		err = nl.RxUntilDone(func(msg netlink.Message) error {
			defer msg.Close()
			if msg.MsgType() == netlink.RTM_NEWNSID {
				netns := msg.(*netlink.NetnsMessage)
				attr := netns.Attrs[netlink.NETNSA_NSID]
				if attr != nil {
					i32attr := attr.(netlink.Int32Attr)
					entries[i].Nsid = i32attr.Int()
				}
			}
			return nil
		})
		f.Close()
		if err != nil {
			break
		}
	}
	return entries, err
}
