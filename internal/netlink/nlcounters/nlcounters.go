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

const Usage = "nlcounters [-i SECONDS] [-n COUNT] [-nsid ID]"

var istats map[uint32]*LinkStats64

// Dump all or select netlink messages
func Main(args ...string) error {
	usage := func(format string, args ...interface{}) error {
		return fmt.Errorf(format+"\nusage: "+Usage, args...)
	}
	interval := 0
	n := 0
	nsid := DefaultNsid
	parm, args := parms.New(args, "-i", "-n", "-nsid")
	if len(args) > 0 {
		return usage("%s: unknown", args[0])
	}
	for _, x := range []struct {
		name string
		p    *int
	}{
		{"-i", &interval},
		{"-n", &n},
		{"-nsid", &nsid},
	} {
		if arg := parm[x.name]; len(arg) > 0 {
			_, err := fmt.Sscan(arg, x.p)
			if err != nil {
				return usage("%s: %v", x.name[1:], err)
			}
		}
	}
	istats = make(map[uint32]*LinkStats64)
	nl, err := New()
	if err != nil {
		return err
	}
	nl.GetlinkReq(nsid)
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
			nl.GetlinkReq(nsid)
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
