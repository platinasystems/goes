// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nldump

import (
	"fmt"
	"os"
	"time"

	"github.com/platinasystems/go/internal/indent"
	. "github.com/platinasystems/go/internal/netlink"
)

const Usage = "nldump [ link|addr|route|neighor|nsid ]..."

type flags uint

// Dump all or select netlink messages
func Main(args ...string) error {
	const (
		link flags = 1 << iota
		addr
		route
		neighbor
		nsid
		end

		all = end - 1
	)
	var (
		dump   flags
		groups []MulticastGroup
		reqs   []ListenReq
	)
	for _, arg := range args {
		switch arg {
		case "-h", "-help", "--help":
			fmt.Println("usage:", Usage)
			return nil
		case "link":
			dump |= link
		case "addr":
			dump |= addr
		case "route":
			dump |= route
		case "neighbor":
			dump |= neighbor
		case "nsid":
			dump |= nsid
		default:
			return fmt.Errorf("%s: unknown", arg)
		}
	}
	if dump == 0 {
		dump = all
	}
	if dump.has(link) {
		groups = append(groups, RTNLGRP_LINK)
		reqs = append(reqs,
			ListenReq{RTM_GETLINK, AF_PACKET})
	}
	if dump.has(addr) {
		groups = append(groups,
			RTNLGRP_IPV4_IFADDR,
			RTNLGRP_IPV6_IFADDR)
		reqs = append(reqs,
			ListenReq{RTM_GETADDR, AF_INET},
			ListenReq{RTM_GETADDR, AF_INET6})
	}
	if dump.has(route) {
		groups = append(groups,
			RTNLGRP_IPV4_ROUTE,
			RTNLGRP_IPV6_ROUTE,
			RTNLGRP_IPV4_MROUTE,
			RTNLGRP_IPV6_MROUTE)
		reqs = append(reqs,
			ListenReq{RTM_GETROUTE, AF_INET},
			ListenReq{RTM_GETROUTE, AF_INET6})
	}
	if dump.has(neighbor) {
		groups = append(groups, RTNLGRP_NEIGH)
		reqs = append(reqs,
			ListenReq{RTM_GETNEIGH, AF_INET},
			ListenReq{RTM_GETNEIGH, AF_INET6})
	}
	if dump.has(nsid) {
		groups = append(groups, RTNLGRP_NSID)
		reqs = append(reqs,
			ListenReq{RTM_GETNSID, AF_UNSPEC})
	}
	nl, err := New(groups...)
	if err != nil {
		return err
	}
	handler := func(msg Message) (err error) {
		defer msg.Close()
		if msg.MsgType() != NLMSG_DONE {
			_, err = msg.WriteTo(indent.New(os.Stdout, "    "))
		}
		return err
	}
	if err = nl.Listen(handler, reqs...); err != nil {
		return err
	}
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	nl.GetlinkReq(DefaultNsid)
	for {
		select {
		case <-t.C:
			nl.GetlinkReq(DefaultNsid)
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

func (f flags) has(t flags) bool { return (f & t) == t }
