// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nldump

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/indent"
	"github.com/platinasystems/go/internal/netlink"
)

const Usage = "nldump [ link|addr|route|neighor|nsid ]... [NSID]..."

// Dump all or select netlink messages
func Main(args ...string) error {
	var (
		groups []netlink.MulticastGroup
		reqs   []netlink.ListenReq
		nsids  []int
	)
	flag, args := flags.New(args, "-help", "-h", "--help",
		"link", "addr", "route", "neighbor", "nsid")
	flag.Aka("-help", "-h", "--help")
	if len(args) > 0 {
		for _, s := range args {
			var nsid int
			_, err := fmt.Sscan(s, &nsid)
			if err != nil {
				return fmt.Errorf("%s: %v\nusage: %s",
					s, err, Usage)
			}
			nsids = append(nsids, nsid)
		}
	} else {
		nsids = []int{netlink.DefaultNsid}
	}
	if flag["-help"] {
		fmt.Println("usage:", Usage)
		return nil
	}
	if len(flag) == 0 {
		flag["link"] = true
		flag["addr"] = true
		flag["route"] = true
		flag["neighbor"] = true
		flag["nsid"] = true
	}
	sort.Ints(nsids)
	if flag["link"] {
		groups = append(groups, netlink.LinkMulticastGroups...)
		reqs = append(reqs, netlink.LinkListenReqs...)
	}
	if flag["addr"] {
		groups = append(groups, netlink.AddrMulticastGroups...)
		reqs = append(reqs, netlink.AddrListenReqs...)
	}
	if flag["route"] {
		groups = append(groups, netlink.RouteMulticastGroups...)
		reqs = append(reqs, netlink.RouteListenReqs...)
	}
	if flag["neighbor"] {
		groups = append(groups, netlink.NeighborMulticastGroups...)
		reqs = append(reqs, netlink.NeighborListenReqs...)
	}
	if flag["nsid"] {
		groups = append(groups, netlink.NsidMulticastGroups...)
		reqs = append(reqs, netlink.NsidListenReqs...)
	}
	nl, err := netlink.New(groups...)
	if err != nil {
		return err
	}
	handler := func(msg netlink.Message) (err error) {
		defer msg.Close()
		nsid := *msg.Nsid()
		i := sort.SearchInts(nsids, nsid)
		found := i < len(nsids) && nsids[i] == nsid
		if msg.MsgType() != netlink.NLMSG_DONE && found {
			_, err = msg.WriteTo(indent.New(os.Stdout, "    "))
		}
		return err
	}
	if err = nl.Listen(handler, reqs...); err != nil {
		return err
	}
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	nl.GetlinkReq()
	for {
		select {
		case <-t.C:
			nl.GetlinkReq()
		case msg, opened := <-nl.Rx:
			if !opened {
				return nil
			}
			if err = handler(msg); err != nil {
				return err
			}
		}
	}
	return err
}
