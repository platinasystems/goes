// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package monitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "monitor"
	Apropos = "print netlink messages"
	Usage   = `
	ip monitor file FILE [label]
	ip monitor [ all | OBJECT... ] save file
	ip monitor [ all | OBJECT... ] [label] [all-nsid] [-t | -ts]

OBJECT := link | address | route | mroute | prefix | neigh | netconf | rule |
	nsid`
	Man = `
OPTIONS
	file FILE
		read netlink messages from FILE instead of socket

	save FILE
		instead of print, output raw, time stamped messages to FILE

	label	identify type of message (e.g. LINK, ADDR, NEIGH, ROUTE)

	all-nsid
		listen in all identified network namespaces

	-t, -timestamp
		Prints timestamp before the event message on the separated line
		in format:

		Timestamp: <Day> <Mon> <DD> <hh:mm:ss><.ns> <z> <YYYY>
		<EVENT>

	-ts, -tshort
		Prints short timestamp before the event message on the same
		line in format:

		[<YYYY>-<MM>-<DD>T<hh:mm:ss>.<ns><+|-tz>] <EVENT>

SEE ALSO
	ip man monitor || ip monitor -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		"all",
		"link",
		"address",
		"route",
		"mroute",
		"prefix",
		"neigh",
		"netconf",
		"rule",
		"nsid",
		"all-nsid",
		"label",
	}
	Parms = []interface{}{
		"file",
		"save",
		"dev",
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var err error
	var handle func([]byte)
	var save save
	var show show

	show.opt, args = options.New(args)
	args = show.opt.Flags.More(args, Flags...)
	args = show.opt.Parms.More(args, Parms...)

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	if fn := show.opt.Parms.ByName["save"]; len(fn) > 0 {
		save.File, err = os.Create(fn)
		if err != nil {
			return err
		}
		defer save.Close()
		save.tsbuf = make([]byte, rtnl.SizeofHdr+sizeofTstamp)
		handle = save.Handle
	} else {
		show.nsid = -1
		handle = show.Handle
	}

	if fn := show.opt.Parms.ByName["file"]; len(fn) > 0 {
		b, err := ioutil.ReadFile(fn)
		if err != nil {
			return err
		}
		for err == nil && len(b) > rtnl.SizeofHdr {
			var msg []byte
			msg, b, err = rtnl.Pop(b)
			handle(msg)
		}
		return err
	}

	if show.ifnames, err = ifnamesByIndex(); err != nil {
		return err
	}

	sock, err := rtnl.NewSock(16, groups(show.opt),
		show.opt.Flags.ByName["all-nsid"])
	if err != nil {
		return err
	}
	defer sock.Close()

	sigch := make(chan os.Signal)
	signal.Notify(sigch, os.Interrupt, os.Signal(syscall.SIGTERM))

selectLoop:
	for err == nil {
		select {
		case <-sigch:
			break selectLoop
		case b, opened := <-sock.RxCh:
			if !opened {
				break selectLoop
			}
			for err == nil && len(b) > rtnl.SizeofHdr {
				var msg []byte
				msg, b, err = rtnl.Pop(b)
				handle(msg)
			}
		}
	}
	return err
}

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
	cpv["file"] = options.CompleteFile
	cpv["save"] = options.NoComplete
	if method, found := cpv[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"file",
			"save",
			"label",
			"all-nsid",
			"all",
			"link",
			"address",
			"route",
			"mroute",
			"prefix",
			"neigh",
			"netconf",
			"rule",
			"nsid",
		) {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	return
}

func groups(opt *options.Options) uint32 {
	var groups uint32
	if opt.Flags.ByName["link"] {
		groups |= rtnl.RTNLGRP_LINK.Bit()
	}
	if opt.Flags.ByName["neigh"] {
		groups |= rtnl.RTNLGRP_NEIGH.Bit()
		groups |= rtnl.RTNLGRP_ND_USEROPT.Bit()
	}
	if opt.Flags.ByName["nsid"] {
		groups |= rtnl.RTNLGRP_NSID.Bit()
	}
	switch opt.Parms.ByName["-f"] {
	case "inet":
		if opt.Flags.ByName["address"] {
			groups |= rtnl.RTNLGRP_IPV4_IFADDR.Bit()
		}
		if opt.Flags.ByName["route"] {
			groups |= rtnl.RTNLGRP_IPV4_ROUTE.Bit()
		}
		if opt.Flags.ByName["mroute"] {
			groups |= rtnl.RTNLGRP_IPV4_MROUTE.Bit()
		}
		if opt.Flags.ByName["netconf"] {
			groups |= rtnl.RTNLGRP_IPV4_NETCONF.Bit()
		}
		if opt.Flags.ByName["rule"] {
			groups |= rtnl.RTNLGRP_IPV4_RULE.Bit()
		}
	case "inet6":
		if opt.Flags.ByName["address"] {
			groups |= rtnl.RTNLGRP_IPV6_IFADDR.Bit()
		}
		if opt.Flags.ByName["route"] {
			groups |= rtnl.RTNLGRP_IPV6_ROUTE.Bit()
		}
		if opt.Flags.ByName["mroute"] {
			groups |= rtnl.RTNLGRP_IPV6_MROUTE.Bit()
		}
		if opt.Flags.ByName["netconf"] {
			groups |= rtnl.RTNLGRP_IPV6_NETCONF.Bit()
		}
		if opt.Flags.ByName["prefix"] {
			groups |= rtnl.RTNLGRP_IPV6_PREFIX.Bit()
		}
		if opt.Flags.ByName["rule"] {
			groups |= rtnl.RTNLGRP_IPV6_RULE.Bit()
		}
	default:
		if opt.Flags.ByName["address"] {
			groups |= rtnl.RTNLGRP_IPV4_IFADDR.Bit()
			groups |= rtnl.RTNLGRP_IPV6_IFADDR.Bit()
		}
		if opt.Flags.ByName["route"] {
			groups |= rtnl.RTNLGRP_IPV4_ROUTE.Bit()
			groups |= rtnl.RTNLGRP_IPV6_ROUTE.Bit()
		}
		if opt.Flags.ByName["mroute"] {
			groups |= rtnl.RTNLGRP_IPV4_MROUTE.Bit()
			groups |= rtnl.RTNLGRP_IPV6_MROUTE.Bit()
		}
		if opt.Flags.ByName["netconf"] {
			groups |= rtnl.RTNLGRP_IPV4_NETCONF.Bit()
			groups |= rtnl.RTNLGRP_IPV6_NETCONF.Bit()
		}
		if opt.Flags.ByName["prefix"] {
			groups |= rtnl.RTNLGRP_IPV6_PREFIX.Bit()
		}
		if opt.Flags.ByName["rule"] {
			groups |= rtnl.RTNLGRP_IPV4_RULE.Bit()
			groups |= rtnl.RTNLGRP_IPV6_RULE.Bit()
		}
	}
	if groups != 0 && !opt.Flags.ByName["all"] {
		return groups
	}
	groups |= rtnl.RTNLGRP_LINK.Bit()
	groups |= rtnl.RTNLGRP_NEIGH.Bit()
	groups |= rtnl.RTNLGRP_ND_USEROPT.Bit()
	groups |= rtnl.RTNLGRP_NSID.Bit()
	switch opt.Parms.ByName["-f"] {
	case "inet":
		groups |= rtnl.RTNLGRP_IPV4_IFADDR.Bit()
		groups |= rtnl.RTNLGRP_IPV4_ROUTE.Bit()
		groups |= rtnl.RTNLGRP_IPV4_MROUTE.Bit()
		groups |= rtnl.RTNLGRP_IPV4_NETCONF.Bit()
		groups |= rtnl.RTNLGRP_IPV4_RULE.Bit()
	case "inet6":
		groups |= rtnl.RTNLGRP_IPV6_IFADDR.Bit()
		groups |= rtnl.RTNLGRP_IPV6_ROUTE.Bit()
		groups |= rtnl.RTNLGRP_IPV6_MROUTE.Bit()
		groups |= rtnl.RTNLGRP_IPV6_NETCONF.Bit()
		groups |= rtnl.RTNLGRP_IPV6_PREFIX.Bit()
		groups |= rtnl.RTNLGRP_IPV6_RULE.Bit()
	default:
		groups |= rtnl.RTNLGRP_IPV4_IFADDR.Bit()
		groups |= rtnl.RTNLGRP_IPV4_ROUTE.Bit()
		groups |= rtnl.RTNLGRP_IPV4_MROUTE.Bit()
		groups |= rtnl.RTNLGRP_IPV4_NETCONF.Bit()
		groups |= rtnl.RTNLGRP_IPV4_RULE.Bit()
		groups |= rtnl.RTNLGRP_IPV6_IFADDR.Bit()
		groups |= rtnl.RTNLGRP_IPV6_ROUTE.Bit()
		groups |= rtnl.RTNLGRP_IPV6_MROUTE.Bit()
		groups |= rtnl.RTNLGRP_IPV6_NETCONF.Bit()
		groups |= rtnl.RTNLGRP_IPV6_PREFIX.Bit()
		groups |= rtnl.RTNLGRP_IPV6_RULE.Bit()
	}
	return groups
}

type save struct {
	*os.File
	tsbuf []byte
}

func (save *save) Handle(b []byte) {
	if len(b) < rtnl.SizeofHdr {
		return
	}
	now := time.Now()
	*(rtnl.HdrPtr(save.tsbuf)) = rtnl.Hdr{
		Len:  uint32(len(save.tsbuf)),
		Type: rtnl.NLMSG_TSTAMP,
	}
	*(*tstamp)(unsafe.Pointer(&save.tsbuf[rtnl.SizeofHdr])) = tstamp{
		secs:  uint32(now.Unix()),
		usecs: uint32(now.UnixNano() / 1000),
	}
	save.Write(save.tsbuf)
	save.Write(b)
}

type show struct {
	opt     *options.Options
	ifnames map[int32]string
	nsid    int
}

func (show *show) Handle(b []byte) {
	const tfmt = "Mon Jan 01 15:04:05.999999999-07:00 2006"
	var deleted bool
	if len(b) < rtnl.SizeofHdr {
		return
	}
	h := rtnl.HdrPtr(b)
	heading := func(label string) {
		if show.opt.Flags.ByName["-t"] {
			show.opt.Print(time.Now().Format(tfmt), "\n")
		} else if show.opt.Flags.ByName["-ts"] {
			show.opt.Print("[",
				time.Now().Format(time.RFC3339Nano),
				"] ")
		}
		if show.opt.Flags.ByName["all-nsid"] {
			if show.nsid == -1 {
				show.opt.Print("[nsid current] ")
			} else {
				show.opt.Print("[nsid ", show.nsid, "] ")
			}
		}
		if show.opt.Flags.ByName["label"] && len(label) > 0 {
			show.opt.Print("[", label, "] ")
		}
		if deleted {
			show.opt.Print("Deleted ")
		}
	}
	switch h.Type {
	case rtnl.NLMSG_NSID:
		show.nsid = *(*int)(unsafe.Pointer(&b[rtnl.SizeofHdr]))
	case rtnl.RTM_DELROUTE:
		deleted = true
		fallthrough
	case rtnl.RTM_NEWROUTE:
		heading("ROUTE")
		show.opt.ShowRoute(b, show.ifnames)
	case rtnl.RTM_DELLINK:
		deleted = true
		heading("LINK")
		show.opt.ShowIfInfo(b)
		msg := rtnl.IfInfoMsgPtr(b)
		delete(show.ifnames, msg.Index)
	case rtnl.RTM_NEWLINK:
		var ifla rtnl.Ifla
		heading("LINK")
		ifla.Write(b)
		msg := rtnl.IfInfoMsgPtr(b)
		show.ifnames[msg.Index] =
			rtnl.Kstring(ifla[rtnl.IFLA_IFNAME])
		show.opt.ShowIfInfo(b)
	case rtnl.RTM_DELADDR:
		deleted = true
		fallthrough
	case rtnl.RTM_NEWADDR:
		const withCacheInfo = false
		heading("ADDR")
		show.opt.ShowIfAddr(b, withCacheInfo)
	case rtnl.RTM_DELADDRLABEL:
		deleted = true
		fallthrough
	case rtnl.RTM_NEWADDRLABEL:
		heading("ADDRLABEL")
		show.opt.ShowIfAddrLbl(b, show.ifnames)
	case rtnl.RTM_DELNEIGH:
		deleted = true
		fallthrough
	case rtnl.RTM_NEWNEIGH, rtnl.RTM_GETNEIGH:
		heading("NEIGH")
		show.opt.ShowNeigh(b, show.ifnames)
	case rtnl.RTM_NEWPREFIX:
		heading("PREFIX")
		show.opt.ShowPrefix(b, show.ifnames)
	case rtnl.RTM_DELRULE:
		deleted = true
		fallthrough
	case rtnl.RTM_NEWRULE:
		heading("RULE")
		show.opt.ShowRule(b, show.ifnames)
	case rtnl.RTM_NEWNETCONF:
		heading("NETCONF")
		show.opt.ShowNetconf(b, show.ifnames)
	case rtnl.NLMSG_TSTAMP:
		ts := *(*tstamp)(unsafe.Pointer(&b[rtnl.SizeofHdr]))
		show.opt.Print("Timestamp: ", time.Unix(int64(ts.secs),
			int64(ts.usecs*1000)))
	case rtnl.RTM_DELNSID:
		deleted = true
		fallthrough
	case rtnl.RTM_NEWNSID:
		heading("NSID")
		var netnsa rtnl.Netnsa
		var sep string
		netnsa.Write(b)
		if val := netnsa[rtnl.NETNSA_NSID]; len(val) > 0 {
			show.opt.Print("nsid=", rtnl.Int32(val))
			sep = ", "
		}
		if val := netnsa[rtnl.NETNSA_PID]; len(val) > 0 {
			show.opt.Print(sep, "pid=", rtnl.Uint32(val))
			sep = ", "
		}
		if val := netnsa[rtnl.NETNSA_FD]; len(val) > 0 {
			show.opt.Print(sep, "fd=", rtnl.Uint32(val))
		}
	case rtnl.NLMSG_NOOP:
		heading("NOOP")
		show.opt.Print("pid=", h.Pid, ", seq=", h.Seq)
	case rtnl.NLMSG_DONE:
		heading("DONE")
		show.opt.Print("pid=", h.Pid, ", seq=", h.Seq)
	case rtnl.NLMSG_ERROR:
		p := rtnl.NlmsgerrPtr(b)
		if p == nil || p.Errno == 0 {
			heading("ACK")
			show.opt.Print("pid=", h.Pid, " seq=", h.Seq)
		} else {
			heading("ERROR")
			show.opt.Println(syscall.Errno(p.Errno))
			show.opt.Print("type=", p.Req.Type)
			show.opt.Print("; pid=", p.Req.Pid)
			show.opt.Print("; seq=", p.Req.Seq)
		}
	}
	fmt.Println()
}

func ifnamesByIndex() (map[int32]string, error) {
	sock, err := rtnl.NewSock()
	if err != nil {
		return nil, err
	}
	defer sock.Close()
	return rtnl.NewSockReceiver(sock).IfNamesByIndex()
}

const sizeofTstamp = 4 + 4

type tstamp struct {
	secs, usecs uint32
}
