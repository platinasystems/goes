// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nldump

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/indent"
	"github.com/platinasystems/go/internal/netlink"
)

const (
	Name    = "nldump"
	Apropos = "print netlink multicast messages"
	Usage   = "nldump [ link|addr|route|neighor|nsid ]... [NSID]..."
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var (
		groups []netlink.MulticastGroup
		reqs   []netlink.ListenReq
		nsids  []int
	)
	flag, args := flags.New(args, "link", "addr", "route", "neighbor",
		"nsid")
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
	if flag.ByName["-help"] {
		fmt.Println("usage:", Usage)
		return nil
	}
	nothing := flag.ByName["link"] == false &&
		flag.ByName["addr"] == false &&
		flag.ByName["route"] == false &&
		flag.ByName["neighbor"] == false &&
		flag.ByName["nsid"] == false
	if nothing {
		flag.ByName["link"] = true
		flag.ByName["addr"] = true
		flag.ByName["route"] = true
		flag.ByName["neighbor"] = true
		flag.ByName["nsid"] = true
	}
	sort.Ints(nsids)
	if flag.ByName["link"] {
		groups = append(groups, netlink.LinkMulticastGroups...)
		reqs = append(reqs, netlink.LinkListenReqs...)
	}
	if flag.ByName["addr"] {
		groups = append(groups, netlink.AddrMulticastGroups...)
		reqs = append(reqs, netlink.AddrListenReqs...)
	}
	if flag.ByName["route"] {
		groups = append(groups, netlink.RouteMulticastGroups...)
		reqs = append(reqs, netlink.RouteListenReqs...)
	}
	if flag.ByName["neighbor"] {
		groups = append(groups, netlink.NeighborMulticastGroups...)
		reqs = append(reqs, netlink.NeighborListenReqs...)
	}
	if flag.ByName["nsid"] {
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
