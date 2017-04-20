// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Periodically poll link counters and print deltas.
package nlcounters

import (
	"fmt"
	"sort"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/internal/parms"
)

const Usage = "nlcounters [-i SECONDS] [-n COUNT] [NSID]..."

var istats map[uint32]*netlink.LinkStats64
var nsids []int

// Dump all or select netlink messages
func Main(args ...string) error {
	usage := func(format string, args ...interface{}) error {
		return fmt.Errorf(format+"\nusage: "+Usage, args...)
	}
	interval := 0
	n := 0
	parm, args := parms.New(args, "-i", "-n")
	if len(args) > 0 {
		for _, arg := range args {
			var nsid int
			_, err := fmt.Sscan(arg, &nsid)
			if err != nil {
				return usage("%s: %v", arg, err)
			}
			nsids = append(nsids, nsid)
		}
	} else {
		nsids = []int{
			netlink.DefaultNsid,
		}
	}
	sort.Ints(nsids)
	for _, x := range []struct {
		name string
		p    *int
	}{
		{"-i", &interval},
		{"-n", &n},
	} {
		if arg := parm[x.name]; len(arg) > 0 {
			_, err := fmt.Sscan(arg, x.p)
			if err != nil {
				return usage("%s: %v", x.name[1:], err)
			}
		}
	}
	istats = make(map[uint32]*netlink.LinkStats64)
	nl, err := netlink.New(netlink.LinkMulticastGroups...)
	if err != nil {
		return err
	}
	getlink := func() {
		for _, nsid := range nsids {
			nl.GetlinkReq(nsid)
		}
	}
	getlink()
	if interval <= 0 {
		return nl.RxUntilDone(handler)
	}
	t := time.NewTicker(time.Duration(interval) * time.Second)
	defer t.Stop()
	for i := 0; ; {
		select {
		case <-t.C:
			if n > 0 {
				if i++; i == n {
					return nil
				}
			}
			getlink()
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

func handler(msg netlink.Message) (err error) {
	defer msg.Close()
	switch msg.MsgType() {
	case netlink.NLMSG_ERROR:
		e := msg.(*netlink.ErrorMessage)
		if e.Errno != 0 {
			err = syscall.Errno(-e.Errno)
		}
	case netlink.RTM_NEWLINK:
		nsid := *msg.Nsid()
		i := sort.SearchInts(nsids, nsid)
		if i >= len(nsids) || nsids[i] != nsid {
			return
		}
		ifinfo := msg.(*netlink.IfInfoMessage)
		attr := ifinfo.Attrs[netlink.IFLA_IFNAME]
		name := attr.(netlink.StringAttr).String()
		attr = ifinfo.Attrs[netlink.IFLA_STATS64]
		stats := attr.(*netlink.LinkStats64)
		old, found := istats[ifinfo.Index]
		if !found {
			old = new(netlink.LinkStats64)
			istats[ifinfo.Index] = old
		}
		for i, v := range stats {
			k := netlink.Key(netlink.LinkStatType(i).String())
			if v != 0 && v != old[i] {
				fmt.Print(name, ".", k, ": ", v-old[i], "\n")
				old[i] = v
			}
		}
	}
	return
}
