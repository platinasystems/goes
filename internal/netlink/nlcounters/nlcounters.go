// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Periodically poll link counters and print deltas.
package nlcounters

import (
	"fmt"
	"syscall"
	"time"

	. "github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/internal/parms"
)

const Usage = "nlcounters [-i SECONDS]"

var istats map[uint32]*LinkStats64

// Dump all or select netlink messages
func Main(args ...string) error {
	seconds := 5 * time.Second
	switch len(args) {
	case 0:
	case 2:
		if args[0] != "-i" {
			return fmt.Errorf("%s: unknown", args[0])
		}
		_, err := fmt.Sscan(args[1], &seconds)
		if err != nil {
			return err
		}
		seconds *= time.Second
	default:
		return fmt.Errorf("%v: unexpected", args)
	}
	istats = make(map[uint32]*LinkStats64)
	nl, err := New()
	if err != nil {
		return err
	}
	t := time.NewTicker(seconds)
	defer t.Stop()
	nl.GetlinkReq(DefaultNsid)
	for i := 0; ; {
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

func handler(msg Message) (err error) {
	defer msg.Close()
	switch msg.MsgType() {
	case NLMSG_ERROR:
		e := msg.(*ErrorMessage)
		if e.Errno != 0 {
			err = syscall.Errno(-e.Errno)
		}
	case RTM_NEWLINK:
		ifinfo := msg.(*IfInfoMessage)
		name := ifinfo.Attrs[IFLA_IFNAME].(StringAttr).String()
		stats := ifinfo.Attrs[IFLA_STATS64].(*LinkStats64)
		old, found := istats[ifinfo.Index]
		if !found {
			old = new(LinkStats64)
			istats[ifinfo.Index] = old
		}
		for i, v := range stats {
			if v != 0 && v != old[i] {
				fmt.Print(name, ".",
					Key(LinkStatType(i).String()), ": ",
					v-old[i], "\n")
				old[i] = v
			}
		}
	}
	return
}
