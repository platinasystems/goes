// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package counters

import (
	"fmt"
	"sort"
	"time"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
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

	-prefix STRING
		STRING prepended to each counter name

	-publish
		Instead of print, publish counters to local redis server.
		This should be run as a daemon, e.g.
			goes-daemons start ip link counters -publish
		or
			goes-daemons start ip -netns NAME \
				link counters -publish -prefix NAME.
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
	Parms = []interface{}{"-interval", "-total", "-prefix"}
)

func New() *Command { return &Command{} }

type Command struct {
	prefix string

	ifinfos     [][]byte
	lastByIndex map[int32][]byte
}

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Man() lang.Alt     { return man }
func (*Command) String() string    { return Name }
func (*Command) Usage() string     { return Usage }

func (*Command) Help(args ...string) string {
	help := "no help"
	switch {
	case len(args) == 0:
		help = Usage
	case args[0] == "-interval":
		return "SECONDS"
	case args[0] == "-total":
		return "NUMBER"
	case args[0] == "-prefix":
		return "PREFIX"
	}
	return help
}

func (c *Command) Main(args ...string) error {
	var err error
	var total int
	interval := 5

	if args, err = options.Netns(args); err != nil {
		return err
	}

	opt, args := options.New(args)
	args = opt.Flags.More(args, Flags...)
	args = opt.Parms.More(args, Parms...)

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	publish := opt.Flags.ByName["-publish"]

	for _, x := range []struct {
		parm string
		name string
		p    *int
	}{
		{"-interval", "SECONDS", &interval},
		{"-total", "NUMBER", &total},
	} {
		if s := opt.Parms.ByName[x.parm]; len(s) > 0 {
			if _, err := fmt.Sscan(s, x.p); err != nil {
				return fmt.Errorf("%s: %v", x.name, err)
			}
		}
	}
	if interval <= 0 {
		return fmt.Errorf("invalid interval")
	}

	c.prefix = opt.Parms.ByName["-prefix"]
	c.lastByIndex = make(map[int32][]byte)

	t := time.NewTicker(time.Duration(interval) * time.Second)
	defer t.Stop()
	for i := 0; ; {
		err = c.getIfInfos(publish)
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

func (c *Command) getIfInfos(publish bool) error {
	var printf func(string, ...interface{}) (int, error)

	sock, err := rtnl.NewSock()
	if err != nil {
		return fmt.Errorf("socket: %v", err)
	}
	defer sock.Close()

	sr := rtnl.NewSockReceiver(sock)

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
		return fmt.Errorf("request: %v", err)
	}
	c.ifinfos = c.ifinfos[:0]
	if err = sr.UntilDone(req, func(b []byte) {
		if rtnl.HdrPtr(b).Type == rtnl.RTM_NEWLINK {
			c.ifinfos = append(c.ifinfos, b)
		}
	}); err != nil {
		return fmt.Errorf("response: %v", err)
	}
	sort.Slice(c.ifinfos, func(i, j int) bool {
		iIndex := rtnl.IfInfoMsgPtr(c.ifinfos[i]).Index
		jIndex := rtnl.IfInfoMsgPtr(c.ifinfos[j]).Index
		return iIndex < jIndex
	})
	if publish {
		pub, err := publisher.New()
		if err != nil {
			return fmt.Errorf("publisher: %v", err)
		}
		defer pub.Close()
		printf = pub.Printf
	} else {
		printf = fmt.Printf
	}
	for _, b := range c.ifinfos {
		var ifla, lifla rtnl.Ifla
		var lmsg *rtnl.IfInfoMsg
		var loper uint8
		var lstats *rtnl.IfStats64
		msg := rtnl.IfInfoMsgPtr(b)
		ifla.Write(b)
		ifname := rtnl.Kstring(ifla[rtnl.IFLA_IFNAME])
		lb, found := c.lastByIndex[msg.Index]
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
						printf("%s%s.%s: %s\n",
							c.prefix,
							ifname,
							x.name,
							x.t)
					} else if len(x.f) > 0 {
						printf("%s%s.%s: %s\n",
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
				printf("%s%s.state: %s\n",
					c.prefix, ifname,
					rtnl.IfOperName[oper])
			}
		}
		if val := ifla[rtnl.IFLA_STATS64]; len(val) > 0 {
			stats := rtnl.IfStats64Attr(val)
			for i := 0; i < rtnl.N_link_stat; i++ {
				if !found || lstats == nil ||
					stats[i] != lstats[i] {
					printf("%s%s.%s: %d\n",
						c.prefix,
						ifname,
						rtnl.IfStatNames[i],
						stats[i])
				}
			}
		}
		c.lastByIndex[msg.Index] = b
	}
	return nil
}
