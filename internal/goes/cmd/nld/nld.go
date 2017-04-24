// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nld

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/redis/rpc/args"
	"github.com/platinasystems/go/internal/redis/rpc/reply"
	"github.com/platinasystems/go/internal/sockfile"
)

const (
	Name = "nld"

	withCounters    = true
	withoutCounters = false
)

type cmd struct {
	Info
}

type Info struct {
	mutex  sync.Mutex
	nl     *netlink.Socket
	rpc    *sockfile.RpcServer
	pub    *publisher.Publisher
	bynsid map[int]*devidx
	name   struct {
		addr [netlink.IFA_MAX]string
		attr [netlink.IFLA_MAX]string
		stat [netlink.N_link_stat]string
	}
}

type devidx struct {
	dev map[uint32]*dev
	idx map[string]uint32
}

func newdevidx() *devidx {
	return &devidx{
		dev: make(map[uint32]*dev),
		idx: make(map[string]uint32),
	}
}

type dev struct {
	name   string
	family uint8
	flags  netlink.IfInfoFlags
	addrs  map[string]string
	attrs  map[string]string
	stats  netlink.LinkStats64
}

func newdev() *dev {
	return &dev{
		addrs: make(map[string]string),
		attrs: make(map[string]string),
	}
}

type getAddrReqParm struct {
	idx uint32
	af  netlink.AddressFamily
}

func New() *cmd { return &cmd{} }

var kbuf = &bytes.Buffer{}

func newkey(nsid int, vs ...interface{}) string {
	kbuf.Reset()
	kbuf.WriteString("nl")
	if nsid != -1 {
		fmt.Fprint(kbuf, ".", nsid)
	}
	for _, v := range vs {
		fmt.Fprint(kbuf, ".", v)
	}
	return kbuf.String()
}

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	var err error
	cmd.nl, err = netlink.New(
		netlink.RTNLGRP_LINK,
		netlink.RTNLGRP_IPV4_IFADDR,
		netlink.RTNLGRP_IPV6_IFADDR)
	if err != nil {
		return err
	}
	cmd.rpc, err = sockfile.NewRpcServer(Name)
	if err != nil {
		return err
	}
	cmd.pub, err = publisher.New()
	if err != nil {
		return err
	}

	cmd.bynsid = make(map[int]*devidx)

	for i := range cmd.name.addr {
		cmd.name.addr[i] =
			netlink.Key(netlink.IfAddrAttrKind(i).String())
	}
	for i := range cmd.name.attr {
		cmd.name.attr[i] =
			netlink.Key(netlink.IfInfoAttrKind(i).String())
	}
	for i := range cmd.name.stat {
		cmd.name.stat[i] =
			netlink.Key(netlink.LinkStatType(i).String())
	}

	rpc.Register(&cmd.Info)
	err = redis.Assign(redis.DefaultHash+":nl.", Name, "Info")
	if err != nil {
		return err
	}

	err = cmd.nl.Listen(cmd.handler,
		netlink.ListenReq{netlink.RTM_GETLINK, netlink.AF_PACKET},
		netlink.ListenReq{netlink.RTM_GETADDR, netlink.AF_INET},
		netlink.ListenReq{netlink.RTM_GETADDR, netlink.AF_INET6},
	)
	if err != nil {
		return err
	}

	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	cmd.nl.GetlinkReq()
	for {
		select {
		case <-t.C:
			cmd.nl.GetlinkReq()
		case msg, opened := <-cmd.nl.Rx:
			if !opened {
				return nil
			}
			if err = cmd.handler(msg); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cmd *cmd) Close() error {
	var err error
	for _, c := range []io.Closer{
		cmd.nl,
		cmd.rpc,
		cmd.pub,
	} {
		t := c.Close()
		if err == nil {
			err = t
		}
	}
	return err
}

func (cmd *cmd) handler(msg netlink.Message) error {
	defer msg.Close()
	switch msg.MsgType() {
	case netlink.RTM_NEWLINK, netlink.RTM_GETLINK:
		// FIXME what about RTM_DELLINK and RTM_SETLINK ?
		cmd.ifInfo(msg.(*netlink.IfInfoMessage))
	case netlink.RTM_DELADDR:
		cmd.ifDelAddr(msg.(*netlink.IfAddrMessage))
	case netlink.RTM_NEWADDR:
		cmd.ifNewAddr(msg.(*netlink.IfAddrMessage))
	}
	return nil
}

func (info *Info) Hdel(args args.Hset, reply *reply.Hset) error {
	a := redis.Split(args.Field)
	if len(a) > 0 && a[0] == "nl" {
		a = a[1:]
	}
	err := info.delLinkAddr(a)
	if err == nil {
		*reply = 1
	}
	return err
}

func (info *Info) Hset(args args.Hset, reply *reply.Hset) (err error) {
	nsid := -1
	a := redis.Split(args.Field)
	if len(a) > 0 && a[0] == "nl" {
		a = a[1:]
	}
	if len(a) > 0 {
		_, xerr := fmt.Sscan(a[0], &nsid)
		if xerr == nil {
			a = a[1:]
		} else {
			nsid = -1
		}
	}
	s := string(args.Value)
	switch {
	case len(a) == 1:
		err = info.newLinkAddr(nsid, a[0], s, "", "")
	case strings.ContainsAny(a[1], ".:"):
		var attr string
		if len(a) > 2 {
			attr = a[2]
		}
		err = info.newLinkAddr(nsid, a[0], a[1], attr, s)
	default:
		err = info.setLinkAttr(nsid, a[0], a[1], s)
	}
	if err == nil {
		*reply = 1
	}
	return
}

func (info *Info) ifDelAddr(msg *netlink.IfAddrMessage) {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	nsid := *msg.Nsid()
	idx := msg.Index
	di, found := info.bynsid[nsid]
	if !found {
		return
	}
	dev, found := di.dev[idx]
	if !found {
		return
	}

	cidr := fmt.Sprint(msg.Attrs[netlink.IFA_ADDRESS], "/", msg.Prefixlen)
	if strings.Contains(cidr, ".") {
		cidr = fmt.Sprint("[", cidr, "]")
	}

	for i, attr := range msg.Attrs {
		if attr == nil {
			continue
		}

		addr := fmt.Sprint(cidr, ".", info.name.addr[i])

		_, found := dev.addrs[addr]
		if found {
			delete(dev.addrs, addr)
			k := newkey(nsid, dev.name, addr)
			info.pub.Print("delete: ", k)
		}
	}
}

func (info *Info) ifNewAddr(msg *netlink.IfAddrMessage) {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	nsid := *msg.Nsid()
	idx := msg.Index
	di, found := info.bynsid[nsid]
	if !found {
		di = newdevidx()
		info.bynsid[nsid] = di
	}
	dev, found := di.dev[idx]
	if !found {
		dev = newdev()
		di.dev[idx] = dev
	}

	cidr := fmt.Sprint(msg.Attrs[netlink.IFA_ADDRESS], "/", msg.Prefixlen)
	if strings.Contains(cidr, ".") {
		cidr = fmt.Sprint("[", cidr, "]")
	}

	for i, attr := range msg.Attrs {
		if attr == nil {
			continue
		}

		addr := fmt.Sprint(cidr, ".", info.name.addr[i])
		s := attr.String()

		as, found := dev.addrs[addr]
		if !found || s != as {
			k := newkey(nsid, dev.name, addr)
			info.pub.Print(k, ": ", s)
			dev.addrs[addr] = s
		}
	}
}

func (info *Info) ifInfo(msg *netlink.IfInfoMessage) {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	nsid := *msg.Nsid()
	idx := uint32(msg.Index)

	var name string
	for i, attr := range msg.Attrs {
		if attr == nil {
			continue
		}
		if netlink.IfInfoAttrKind(i) == netlink.IFLA_IFNAME {
			name = attr.String()
			break
		}
	}

	di, found := info.bynsid[nsid]
	if !found {
		di = newdevidx()
		info.bynsid[nsid] = di
	}
	dev, found := di.dev[idx]
	if !found {
		dev = newdev()
		dev.name = name
		di.dev[idx] = dev
		di.idx[name] = idx
	}

	if len(name) > 0 && name != dev.name {
		// renamed
		delete(di.idx, dev.name)
		dev.name = name
		di.idx[name] = idx
	}

	if dev.family == 0 {
		dev.family = msg.Family
	}

	switch {
	case dev.flags == 0:
		dev.flags = msg.IfInfomsg.Flags
	case msg.Change != 0:
		newDevFlags := dev.flags
		if flags := msg.IfInfomsg.Flags & msg.Change; flags != 0 {
			newDevFlags |= flags
		}
		if flags := ^msg.IfInfomsg.Flags & msg.Change; flags != 0 {
			newDevFlags &= ^flags
		}
		if msg.IfInfomsg.Flags&netlink.IFF_LOWER_UP == 0 {
			newDevFlags &= ^netlink.IFF_LOWER_UP
		} else {
			newDevFlags |= netlink.IFF_LOWER_UP
		}
		if dev.flags != newDevFlags {
			dev.flags = newDevFlags
		}
	}
	for _, flag := range []struct {
		bit  netlink.IfInfoFlags
		name string
		show string
	}{
		{netlink.IFF_UP, "admin", "up"},
		{netlink.IFF_BROADCAST, "may-broadcast", "true"},
		{netlink.IFF_DEBUG, "debug", "enabled"},
		{netlink.IFF_LOOPBACK, "loopback", "enabled"},
		{netlink.IFF_POINTOPOINT, "point-to-point", "true"},
		{netlink.IFF_NOTRAILERS, "no-trailers", "enabled"},
		{netlink.IFF_RUNNING, "running", "yes"},
		{netlink.IFF_NOARP, "no-arp", "enabled"},
		{netlink.IFF_PROMISC, "promiscuous", "enabled"},
		{netlink.IFF_ALLMULTI, "all-multicast", "enabled"},
		{netlink.IFF_MASTER, "is-master", "true"},
		{netlink.IFF_SLAVE, "is-slave", "true"},
		{netlink.IFF_MULTICAST, "multicast-encap", "enabled"},
		{netlink.IFF_PORTSEL, "port-select", "enabled"},
		{netlink.IFF_AUTOMEDIA, "automedia", "enabled"},
		{netlink.IFF_DYNAMIC, "dynamic", "enabled"},
		{netlink.IFF_LOWER_UP, "link", "up"},
		{netlink.IFF_DORMANT, "dormant", "true"},
		{netlink.IFF_ECHO, "echo", "true"},
	} {
		_, found = dev.attrs[flag.name]
		if dev.flags&flag.bit == flag.bit {
			if !found {
				dev.attrs[flag.name] = flag.show
				k := newkey(nsid, dev.name, flag.name)
				info.pub.Print(k, ": ", flag.show)
			}
		} else if found {
			delete(dev.attrs, flag.name)
			k := newkey(nsid, dev.name, flag.name)
			info.pub.Print("delete: ", k)
		}
	}
	for i, attr := range msg.Attrs {
		if attr == nil {
			continue
		}
		k := netlink.IfInfoAttrKind(i)
		aname := info.name.attr[i]
		switch k {
		case netlink.IFLA_STATS:
		case netlink.IFLA_AF_SPEC:
		case netlink.IFLA_STATS64:
			for i, n := range attr.(*netlink.LinkStats64) {
				if n > dev.stats[i] {
					dev.stats[i] = n
					k := newkey(nsid, dev.name,
						info.name.stat[i])
					info.pub.Print(k, ": ", n)
				}
			}
		default:
			s := attr.String()
			as, found := dev.attrs[aname]
			if !found || s != as {
				dev.attrs[aname] = s
				k := newkey(nsid, dev.name, aname)
				info.pub.Print(k, ": ", s)
				break
			}
		}
	}
}

// args: [NSID,] LINK, CIDR
func (info *Info) delLinkAddr(args []string) error {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	var link, cidr string
	nsid := -1

	switch len(args) {
	case 2:
		_, err := fmt.Sscan(args[0], &nsid)
		if err != nil {
			return err
		}
		link, cidr = args[0], args[1]
	case 3:
		link, cidr = args[1], args[2]
	default:
		return fmt.Errorf("%v invalid", args)
	}

	di, found := info.bynsid[nsid]
	if !found {
		return fmt.Errorf("nsid: %d: not found", nsid)
	}
	idx, found := di.idx[link]
	if !found {
		return fmt.Errorf("link: %s: not found", link)
	}
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	req := netlink.NewIfAddrMessage()
	*req.Nsid() = nsid
	req.Header.Type = netlink.RTM_DELADDR
	req.Header.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_REPLACE
	req.IfAddrmsg.Index = idx

	prefixlen, _ := ipnet.Mask.Size()
	req.IfAddrmsg.Prefixlen = uint8(prefixlen)

	if ip4 := ip.To4(); ip4 != nil {
		req.IfAddrmsg.Family = netlink.AF_INET
		req.Attrs[netlink.IFA_ADDRESS] =
			netlink.NewIp4AddressBytes(ip4[:4])
	} else {
		req.IfAddrmsg.Family = netlink.AF_INET6
		req.Attrs[netlink.IFA_ADDRESS] =
			netlink.NewIp6AddressBytes(ip[:16])
	}

	info.nl.Tx <- req
	return nil
}

func (info *Info) newLinkAddr(nsid int, link, cidr, attr, value string) error {
	var addrAttr netlink.Attr

	info.mutex.Lock()
	defer info.mutex.Unlock()

	di, found := info.bynsid[nsid]
	if !found {
		return fmt.Errorf("nsid: %d: not found", nsid)
	}
	idx, found := di.idx[link]
	if !found {
		return fmt.Errorf("link: %s: not found", link)
	}

	req := netlink.NewIfAddrMessage()
	req.Header.Type = netlink.RTM_NEWADDR
	req.Header.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_CREATE
	req.IfAddrmsg.Index = idx

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	prefixlen, _ := ipnet.Mask.Size()
	req.IfAddrmsg.Prefixlen = uint8(prefixlen)
	req.Attrs[netlink.IFA_FLAGS] =
		netlink.IfAddrFlagAttr(netlink.IFA_F_PERMANENT)

	if ip4 := ip.To4(); ip4 != nil {
		req.IfAddrmsg.Family = netlink.AF_INET
		addrAttr = netlink.NewIp4AddressBytes(ip4[:4])
		if prefixlen == 32 {
			req.IfAddrmsg.Scope = netlink.RT_SCOPE_HOST.Uint()
		} else {
			req.IfAddrmsg.Scope = netlink.RT_SCOPE_UNIVERSE.Uint()
		}
	} else {
		req.IfAddrmsg.Family = netlink.AF_INET6
		addrAttr = netlink.NewIp6AddressBytes(ip[:16])
		if ip4.IsLinkLocalUnicast() || ip4.IsLinkLocalMulticast() {
			req.IfAddrmsg.Scope = netlink.RT_SCOPE_LINK.Uint()
		}
	}

	if ip.IsMulticast() {
		req.Attrs[netlink.IFA_MULTICAST] = addrAttr
	} else {
		req.Attrs[netlink.IFA_ADDRESS] = addrAttr
	}

	req.Attrs[netlink.IFA_LOCAL] = addrAttr

	switch attr {
	case "broadcast":
		var bip4address netlink.Ip4Address
		var bip6address netlink.Ip6Address
		var baddrAttr netlink.Attr
		bip := net.ParseIP(value)
		if bip == nil {
			return fmt.Errorf("%s: invalid address", value)
		}
		if bip4 := bip.To4(); bip4 != nil {
			copy(bip4address[:], bip4[:4])
			baddrAttr = &bip4address
		} else {
			copy(bip6address[:], bip[:16])
			baddrAttr = &bip6address
		}
		req.Attrs[netlink.IFA_BROADCAST] = baddrAttr
	case "local":
	case "label":
	case "anycast":
	case "":
	default:
		return fmt.Errorf("%s: unknown", attr)
	}

	info.nl.Tx <- req
	return nil
}

func (info *Info) setLinkAttr(nsid int, link, attr, value string) error {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	di, found := info.bynsid[nsid]
	if !found {
		return fmt.Errorf("nsid: %d: not found", nsid)
	}
	idx, found := di.idx[link]
	if !found {
		return fmt.Errorf("link: %s: not found", link)
	}

	dev := di.dev[idx]

	req := netlink.NewIfInfoMessage()
	req.Header.Type = netlink.RTM_SETLINK
	req.Header.Flags = netlink.NLM_F_REQUEST
	req.IfInfomsg.Index = idx
	req.IfInfomsg.Family = dev.family
	req.IfInfomsg.Flags = dev.flags
	req.IfInfomsg.Change = 0xffffffff

	bit, found := map[string]netlink.IfInfoFlags{
		"admin":           netlink.IFF_UP,
		"debug":           netlink.IFF_DEBUG,
		"no-trailers":     netlink.IFF_NOTRAILERS,
		"no-arp":          netlink.IFF_NOARP,
		"promiscuous":     netlink.IFF_PROMISC,
		"all-multicast":   netlink.IFF_ALLMULTI,
		"multicast-encap": netlink.IFF_MULTICAST,
		"port-select":     netlink.IFF_PORTSEL,
		"automedia":       netlink.IFF_AUTOMEDIA,
		"dynamic":         netlink.IFF_DYNAMIC,
	}[attr]

	switch {
	case found:
		switch value {
		case "up", "Up", "UP",
			"t", "true", "True", "TRUE",
			"y", "yes", "Yes", "YES",
			"enable", "Enable", "ENABLE":
			req.IfInfomsg.Flags |= bit
		case "down", "Down", "DOWN",
			"f", "false", "False", "FALSE",
			"n", "no", "No", "NO",
			"disable", "Disable", "DISABLE":
			req.IfInfomsg.Flags &^= bit
			req.Change = bit
		default:
			return fmt.Errorf("%s: invalid", value)
		}
	case attr == "mtu":
		var mtu netlink.Uint32Attr
		_, err := fmt.Sscan(value, &mtu)
		if err != nil {
			return err
		}
		req.Attrs[netlink.IFLA_MTU] = mtu
	default:
		return fmt.Errorf("can't set %s", attr)
	}

	info.nl.Tx <- req
	return nil
}
