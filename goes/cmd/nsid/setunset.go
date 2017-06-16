// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nsid

import (
	"os"
	"path/filepath"

	"github.com/platinasystems/go/internal/netlink"
)

func Set(name string, id int32) error {
	return setunset(name, id, netlink.RTM_NEWNSID)
}

func Unset(name string, id int32) error {
	return setunset(name, id, netlink.RTM_DELNSID)
}

func setunset(name string, id int32, mt netlink.MsgType) error {
	f, err := os.Open(filepath.Join(VarRunNetns, name))
	if err != nil {
		return err
	}
	defer f.Close()
	nl, err := netlink.New(netlink.NOOP_RTNLGRP)
	if err != nil {
		return err
	}
	defer nl.Close()
	req := netlink.NewNetnsMessage()
	req.Type = mt
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	req.AddressFamily = netlink.AF_UNSPEC
	req.Attrs[netlink.NETNSA_NSID] = netlink.Int32Attr(id)
	req.Attrs[netlink.NETNSA_FD] = netlink.Uint32Attr(f.Fd())
	nl.Tx <- req
	return nl.RxUntilDone(func(msg netlink.Message) error {
		defer msg.Close()
		return nil
	})
}
