// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nlcounters

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "nlcounters"
	Apropos = "periodic print of netlink interface counters"
	Usage   = "nlcounters [-i SECONDS | -n COUNT]"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Help(args ...string) string {
	help := "no help"
	switch {
	case len(args) == 0:
		help = "[-i SECONDS]"
	case args[0] == "-i":
		return "SECONDS"
	}
	return help
}

func (Command) Main(args ...string) error {
	usage := func(format string, args ...interface{}) error {
		return fmt.Errorf(format+"\nusage: "+Usage, args...)
	}
	var interval, n int
	parm, args := parms.New(args, "-i", "-n")
	if len(args) > 0 {
		return usage("%v: unexpected", args)
	}
	for _, x := range []struct {
		name string
		p    *int
	}{
		{"-i", &interval},
		{"-n", &n},
	} {
		if arg := parm.ByName[x.name]; len(arg) > 0 {
			_, err := fmt.Sscan(arg, x.p)
			if err != nil {
				return usage("%s: %v", x.name[1:], err)
			}
		}
	}
	h, err := newHandler()
	if err != nil {
		return err
	}
	nl, err := netlink.New(netlink.LinkMulticastGroups...)
	if err != nil {
		return err
	}
	err = nl.Listen(h.Handle, netlink.LinkListenReqs...)
	if err != nil {
		return err
	}
	if interval <= 0 {
		return nil
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
			nl.GetlinkReq()
		case msg, opened := <-nl.Rx:
			if !opened {
				return nil
			}
			if err = h.Handle(msg); err != nil {
				return err
			}
		}
	}
	return err
}

type Handler struct {
	istats map[uint32]*netlink.LinkStats64
	netns  string
}

func newHandler() (*Handler, error) {
	var netns string

	f, err := os.Open("/proc/mounts")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		x := strings.Fields(scanner.Text())
		if x[1] == "/sys" && x[0] != "sysfs" {
			netns = x[0]
		}
	}
	err = scanner.Err()
	return &Handler{make(map[uint32]*netlink.LinkStats64), netns}, err
}

func (h *Handler) Handle(msg netlink.Message) (err error) {
	defer msg.Close()
	switch msg.MsgType() {
	case netlink.NLMSG_ERROR:
		e := msg.(*netlink.ErrorMessage)
		if e.Errno != 0 {
			err = syscall.Errno(-e.Errno)
		}
	case netlink.RTM_NEWLINK:
		ifinfo := msg.(*netlink.IfInfoMessage)
		attr := ifinfo.Attrs[netlink.IFLA_IFNAME]
		name := attr.(netlink.StringAttr).String()
		if len(h.netns) > 0 && !strings.HasPrefix(name, "eth-") {
			name = fmt.Sprint(name, "[", h.netns, "]")
		}
		attr = ifinfo.Attrs[netlink.IFLA_STATS64]
		stats := attr.(*netlink.LinkStats64)
		old, found := h.istats[ifinfo.Index]
		if !found {
			old = new(netlink.LinkStats64)
			h.istats[ifinfo.Index] = old
		}
		for i, v := range stats {
			k := netlink.Key(netlink.LinkStatType(i).String())
			if !found || v != old[i] {
				fmt.Print(name, ".", k, ": ", v-old[i], "\n")
				old[i] = v
			}
		}
	}
	return
}
