// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package counters

import (
	"fmt"
	"strings"
	"time"

	"github.com/platinasystems/go/goes/cmd/ip/internal/netns"
	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const (
	Name    = "counters"
	Apropos = "periodic print or publish link counters and state"
	Usage   = "counters [OPTION]..."
	Man     = `
OPTIONS
	-interval SECONDS
		netlink query interval (default: 5)

	-total NUMBER
		quit after this number of queries
		runs forever if total is 0 (default)

	-n[etns] NAME
		print or publish counters of links in the given namespace
		prefaced by this name

	-publish
		Instead of print, publish counters on the local redis server.
		This should be run as a daemon, e.g.
			goes-daemons start ip link counters -publish
		or
			goes-daemons start ip link counters -n NAME -publish
`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{"-publish"}
	Parms = []interface{}{"-interval", "-total", []string{"-n", "-netns"}}
)

func New() Command { return Command{} }

type Command struct{}

type counters struct {
	last    map[int32][]byte
	ifname  map[int32]string
	updated map[int32]bool
	sr      *rtnl.SockReceiver
	printf  func(string, ...interface{}) (int, error)
	prefix  string
}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Complete(args ...string) (list []string) {
	var larg, llarg string
	n := len(args)
	if n > 0 {
		larg = args[n-1]
	}
	if n > 1 {
		llarg = args[n-2]
	}
	cpv := options.CompleteParmValue
	cpv["-interval"] = options.NoComplete
	cpv["-total"] = options.NoComplete
	if method, found := cpv[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"-publish",
			"-interval",
			"-total") {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	return
}

func (Command) Help(args ...string) string {
	help := "no help"
	switch {
	case len(args) == 0:
		help = Usage
	case args[0] == "-interval":
		return "SECONDS"
	case args[0] == "-total":
		return "NUMBER"
	case args[0] == "-netns":
		return "NETNS"
	}
	return help
}

func (Command) Main(args ...string) error {
	var total int
	var c counters
	interval := 5

	flag, args := flags.New(args, Flags...)
	parm, args := parms.New(args, Parms...)
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	if name := parm.ByName["-n"]; len(name) > 0 {
		if err := netns.Switch(name); err != nil {
			return err
		}
		c.prefix = name + "."
	}

	sock, err := rtnl.NewSock()
	if err != nil {
		return fmt.Errorf("socket: %v", err)
	}
	defer sock.Close()

	c.sr = rtnl.NewSockReceiver(sock)

	for _, x := range []struct {
		parm string
		name string
		p    *int
	}{
		{"-interval", "SECONDS", &interval},
		{"-total", "NUMBER", &total},
	} {
		if s := parm.ByName[x.parm]; len(s) > 0 {
			if _, err := fmt.Sscan(s, x.p); err != nil {
				return fmt.Errorf("%s: %v", x.name, err)
			}
		}
	}
	if interval <= 0 {
		return fmt.Errorf("invalid interval")
	}

	c.printf = fmt.Printf
	if flag.ByName["-publish"] {
		pub, err := publisher.New()
		if err != nil {
			return fmt.Errorf("publisher: %v", err)
		}
		defer pub.Close()
		c.printf = pub.Printf
	}

	c.last = make(map[int32][]byte)
	c.ifname = make(map[int32]string)
	c.updated = make(map[int32]bool)

	t := time.NewTicker(time.Duration(interval) * time.Second)
	defer t.Stop()
	for i := 0; ; {
		err = c.counters()
		if err != nil {
			break
		}
		if i++; total > 0 && i >= total {
			break
		}
		<-t.C
	}
	return err
}

func (c *counters) counters() error {
	req, err := rtnl.NewMessage(
		rtnl.Hdr{
			Type:  rtnl.RTM_GETLINK,
			Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_MATCH,
		},
		rtnl.IfInfoMsg{
			Family: rtnl.AF_UNSPEC,
		},
	)
	if err != nil {
		return err
	}
	for k := range c.updated {
		c.updated[k] = false
	}
	err = c.sr.UntilDone(req, func(b []byte) {
		var ifla, lifla rtnl.Ifla
		var lmsg *rtnl.IfInfoMsg
		var loper uint8
		var lstats *rtnl.IfStats64
		if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWLINK {
			return
		}
		msg := rtnl.IfInfoMsgPtr(b)
		ifla.Write(b)
		ifname := rtnl.Kstring(ifla[rtnl.IFLA_IFNAME])
		c.ifname[msg.Index] = ifname
		lb, found := c.last[msg.Index]
		if found {
			lmsg = rtnl.IfInfoMsgPtr(lb)
			lifla.Write(lb)
			if val := lifla[rtnl.IFLA_OPERSTATE]; len(val) > 0 {
				loper = uint8(val[0])
			}
			if val := lifla[rtnl.IFLA_STATS64]; len(val) > 0 {
				lstats = rtnl.IfStats64Attr(val)
			}
		}
		if !found || msg.Flags != lmsg.Flags {
			for _, x := range []struct {
				bit  uint32
				name string
				t, f string
			}{
				{rtnl.IFF_UP, "admin", "up", "down"},
				{rtnl.IFF_LOWER_UP, "lower", "up", "down"},
				{rtnl.IFF_RUNNING, "running", "yes", "no"},
			} {
				iff := msg.Flags & x.bit
				if !found || iff != lmsg.Flags&x.bit {
					if iff != 0 {
						c.printf("%s%s.%s: %s\n",
							c.prefix,
							ifname,
							x.name,
							x.t)
					} else if len(x.f) > 0 {
						c.printf("%s%s.%s: %s\n",
							c.prefix,
							ifname,
							x.name,
							x.f)
					}
				}
			}
		}
		if val := ifla[rtnl.IFLA_OPERSTATE]; len(val) > 0 {
			if oper := uint8(val[0]); !found || oper != loper {
				c.printf("%s%s.state: %s\n",
					c.prefix,
					ifname,
					rtnl.IfOperName[oper])
			}
		}
		if val := ifla[rtnl.IFLA_STATS64]; len(val) > 0 {
			stats := rtnl.IfStats64Attr(val)
			for i := 0; i < rtnl.N_link_stat; i++ {
				if !found || lstats == nil ||
					stats[i] != lstats[i] {
					c.printf("%s%s.%s: %d\n",
						c.prefix,
						ifname,
						rtnl.IfStatNames[i],
						stats[i])
				}
			}
		}
		c.last[msg.Index] = b
		c.updated[msg.Index] = true
	})
	if err != nil {
		return err
	}
	for k := range c.last {
		if !c.updated[k] {
			c.printf("delete: %s%s\n", c.prefix, c.ifname[k])
			delete(c.last, k)
		}
	}
	return nil
}
