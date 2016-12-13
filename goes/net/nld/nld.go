// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nld

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/sockfile"
	"github.com/platinasystems/go/netlink"
	"github.com/platinasystems/go/redis"
	"github.com/platinasystems/go/redis/rpc/args"
	"github.com/platinasystems/go/redis/rpc/reply"
)

const (
	Name = "nld"

	withCounters    = true
	withoutCounters = false
)

// Hook may be reassigned to something that sets the following Prefixes.
var Hook = func() error { return nil }

// Prefixes of device names that nld should service.
var Prefixes = []string{"lo."}

type cmd struct {
	stop chan<- struct{}
	sock *sockfile.RpcServer
	info Info
}

type Info struct {
	mutex     sync.Mutex
	pub       chan<- string
	req       chan<- netlink.Message
	idx       map[string]uint32
	dev       map[uint32]*dev
	efilter   map[uint32]struct{}
	indices   []uint32
	addrNames [netlink.IFA_MAX]string
	attrNames [netlink.IFLA_MAX]string
	statNames [netlink.N_link_stat]string

	getCountersStop chan<- struct{}
	getLinkAddrStop chan<- struct{}
}

type dev struct {
	name   string
	family uint8
	flags  netlink.IfInfoFlags
	addrs  map[string]string
	attrs  map[string]string
	stats  netlink.LinkStats64
}

type getAddrReqParm struct {
	idx uint32
	af  netlink.AddressFamily
}

func New() *cmd { return &cmd{} }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	err := Hook()
	if err != nil {
		return err
	}
	cmd.sock, err = sockfile.NewRpcServer(Name)
	if err != nil {
		return err
	}
	cmd.info.pub, err = redis.Publish(redis.Machine)
	if err != nil {
		return err
	}
	wait := make(chan struct{})
	cmd.stop = wait
	cmd.info.idx = make(map[string]uint32)
	cmd.info.dev = make(map[uint32]*dev)
	cmd.info.efilter = make(map[uint32]struct{})
	cmd.info.indices = make([]uint32, len(Prefixes))

	for i, prefix := range Prefixes {
		name := strings.TrimSuffix(prefix, ".")
		itf, err := net.InterfaceByName(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		idx := uint32(itf.Index)
		cmd.info.idx[name] = uint32(itf.Index)
		cmd.info.dev[idx] = &dev{
			name:  name,
			addrs: make(map[string]string),
			attrs: make(map[string]string),
		}
		cmd.info.indices[i] = idx
	}

	for i := range cmd.info.addrNames {
		s := netlink.IfAddrAttrKind(i).String()
		lc := strings.ToLower(s)
		cmd.info.addrNames[i] = strings.Replace(lc, " ", "_", -1)
	}
	for i := range cmd.info.attrNames {
		s := netlink.IfInfoAttrKind(i).String()
		lc := strings.ToLower(s)
		cmd.info.attrNames[i] = strings.Replace(lc, " ", "_", -1)
	}
	for i := range cmd.info.statNames {
		s := netlink.LinkStatType(i).String()
		lc := strings.ToLower(s)
		cmd.info.statNames[i] = strings.Replace(lc, " ", "_", -1)
	}

	rpc.Register(&cmd.info)
	for _, prefix := range Prefixes {
		key := fmt.Sprintf("%s:%s", redis.Machine, prefix)
		err = redis.Assign(key, Name, "Info")
		if err != nil {
			return err
		}
	}

	cmd.info.getCounters()
	cmd.info.getLinkAddr()

	<-wait
	return nil
}

func (cmd *cmd) Close() error {
	defer close(cmd.stop)
	close(cmd.info.getCountersStop)
	close(cmd.info.getLinkAddrStop)
	close(cmd.info.pub)
	return cmd.sock.Close()
}

func (info *Info) Hdel(args args.Hset, reply *reply.Hset) error {
	err := info.delLinkAddr(redis.Split(args.Field))
	if err == nil {
		*reply = 1
	}
	return err
}

func (info *Info) Hset(args args.Hset, reply *reply.Hset) (err error) {
	a := redis.Split(args.Field)
	s := string(args.Value)
	switch {
	case len(a) == 1:
		err = info.newLinkAddr(a[0], s, "", "")
	case strings.ContainsAny(a[1], ".:"):
		var attr string
		if len(a) > 2 {
			attr = a[2]
		}
		err = info.newLinkAddr(a[0], a[1], attr, s)
	default:
		err = info.setLinkAttr(a[0], a[1], s)
	}
	if err == nil {
		*reply = 1
	}
	return
}

// getLinkAddr listens to Netlink multicast groups for link info (besides
// counters) and address changes.
func (info *Info) getLinkAddr() {
	stop := make(chan struct{})
	info.getLinkAddrStop = stop
	rxch := make(chan netlink.Message, 64)
	sock, err := netlink.New(rxch,
		netlink.RTNLGRP_LINK,
		netlink.RTNLGRP_IPV4_IFADDR,
		netlink.RTNLGRP_IPV6_IFADDR)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	go func(sock *netlink.Socket, wait <-chan struct{}) {
		defer sock.Close()
		<-wait
	}(sock, stop)
	go sock.Listen(
		netlink.ListenReq{netlink.RTM_GETLINK, netlink.AF_UNSPEC},
		netlink.ListenReq{netlink.RTM_GETADDR, netlink.AF_INET},
		netlink.ListenReq{netlink.RTM_GETADDR, netlink.AF_INET6},
	)
	go func(rxch <-chan netlink.Message) {
		desc := "GETADDR "
		for msg := range rxch {
			switch t := msg.(type) {
			case *netlink.DoneMessage:
			case *netlink.ErrorMessage:
				if t.Errno != 0 {
					info.logerr(t)
				}
			case *netlink.IfInfoMessage:
				info.ifInfo(t, withoutCounters)
			case *netlink.IfAddrMessage:
				switch t.Header.Type {
				case netlink.RTM_DELADDR:
					info.ifDelAddr(t)
				case netlink.RTM_NEWADDR:
					info.ifNewAddr(t)
				}
			default:
				fmt.Fprintln(os.Stderr, desc, "unexpected: ",
					msg)
			}
			msg.Close()
		}
	}(rxch)
}

func (info *Info) ifDelAddr(msg *netlink.IfAddrMessage) {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	idx := msg.Index
	dev, found := info.dev[idx]
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

		k := fmt.Sprint(cidr, ".", info.addrNames[i])

		_, found := dev.addrs[k]
		if found {
			delete(dev.addrs, k)
			info.pub <- fmt.Sprintf("delete: %s.%s", dev.name, k)
		}
	}
}

func (info *Info) ifNewAddr(msg *netlink.IfAddrMessage) {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	idx := msg.Index
	dev, found := info.dev[idx]
	if !found { // ignore
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

		k := fmt.Sprint(cidr, ".", info.addrNames[i])
		s := attr.String()

		as, found := dev.addrs[k]
		if !found || s != as {
			info.pub <- fmt.Sprintf("%s.%s: %s", dev.name, k, s)
			dev.addrs[k] = s
		}
	}
}

// getCounters uses a periodic ticker to request link counters.
// The same request socket is used for SETLINK, NEWADDR and DELADDR.
func (info *Info) getCounters() {
	stop := make(chan struct{})
	info.getCountersStop = stop
	rxch := make(chan netlink.Message, 64)
	sock, err := netlink.New(rxch, netlink.NOOP_RTNLGRP)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	go sock.Listen(netlink.NoopListenReq)
	go info.getCounterTicker(sock, stop)
	go info.getCounterRx(rxch)
	return
}

func (info *Info) getCounterTicker(sock *netlink.Socket, wait <-chan struct{}) {
	defer sock.Close()
	seq := uint32(1000000)
	reqch := make(chan netlink.Message, 4)
	info.req = reqch
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			req := netlink.NewGenMessage()
			req.Header.Type = netlink.RTM_GETLINK
			req.Header.Flags = netlink.NLM_F_MATCH
			req.Header.Sequence = seq
			req.AddressFamily = netlink.AF_UNSPEC
			seq++
			req.TxAdd(sock)
			sock.TxFlush()
		case req := <-reqch:
			switch t := req.(type) {
			case *netlink.IfInfoMessage:
				t.Header.Sequence = seq
			case *netlink.IfAddrMessage:
				t.Header.Sequence = seq
			}
			seq++
			req.TxAdd(sock)
			sock.TxFlush()
		case <-wait:
			return
		}
	}
}

func (info *Info) getCounterRx(rxch <-chan netlink.Message) {
	for msg := range rxch {
		switch t := msg.(type) {
		case *netlink.DoneMessage:
		case *netlink.ErrorMessage:
			if t.Errno != 0 {
				info.logerr(t)
			}
		case *netlink.IfInfoMessage:
			info.ifInfo(t, withCounters)
		default:
			fmt.Fprintln(os.Stderr, "GETLINK unexpected: ", msg)
		}
		msg.Close()
	}
}

func (info *Info) ifInfo(msg *netlink.IfInfoMessage, counters bool) {
	info.mutex.Lock()
	defer info.mutex.Unlock()

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

	dev, found := info.dev[idx]
	if !found {
		return
	}

	if len(name) > 0 && name != dev.name {
		// renamed
		delete(info.idx, dev.name)
		dev.name = name
		info.idx[dev.name] = idx
	}

	if dev.family == 0 {
		dev.family = msg.Family
	}
	switch {
	case counters:
		// don't import flags from timed counters responses
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
		if counters {
			// don't import flags from timed counters responses
		} else if dev.flags&flag.bit == flag.bit {
			if !found {
				dev.attrs[flag.name] = flag.show
				info.pub <- fmt.Sprintf("%s.%s: %s",
					dev.name, flag.name, flag.show)
			}
		} else if found {
			delete(dev.attrs, flag.name)
			info.pub <- fmt.Sprintf("delete: %s.%s",
				dev.name, flag.name)
		}
	}
	for i, attr := range msg.Attrs {
		if attr == nil {
			continue
		}
		k := netlink.IfInfoAttrKind(i)
		aname := info.attrNames[i]
		switch k {
		case netlink.IFLA_STATS:
		case netlink.IFLA_AF_SPEC:
		case netlink.IFLA_STATS64:
			if !counters {
				continue
			}
			for i, n := range attr.(*netlink.LinkStats64) {
				if n > dev.stats[i] {
					dev.stats[i] = n
					info.pub <- fmt.Sprintf("%s.%s: %v",
						dev.name, info.statNames[i], n)
				}
			}
		default:
			if counters {
				continue
			}
			s := attr.String()
			as, found := dev.attrs[aname]
			if !found || s != as {
				dev.attrs[aname] = s
				info.pub <- fmt.Sprintf("%s.%s: %s",
					dev.name, aname, s)
				break
			}
		}
	}
}

func (info *Info) logerr(msg *netlink.ErrorMessage) {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	seq := msg.Header.Sequence
	_, found := info.efilter[seq]
	if !found {
		fmt.Fprintln(os.Stderr, msg)
		info.efilter[seq] = struct{}{}
	}
}

// args: LINK, CIDR
func (info *Info) delLinkAddr(args []string) error {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	if len(args) != 2 {
		return fmt.Errorf("%v invalid", args)
	}

	idx, found := info.idx[args[0]]
	if !found {
		return fmt.Errorf("%s not found", args[0])
	}
	ip, ipnet, err := net.ParseCIDR(args[1])
	if err != nil {
		return err
	}

	req := netlink.NewIfAddrMessage()
	req.Header.Type = netlink.RTM_DELADDR
	req.Header.Flags = netlink.NLM_F_REPLACE
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

	info.req <- req
	return nil
}

func (info *Info) newLinkAddr(link, cidr, attr, value string) error {
	var addrAttr netlink.Attr

	info.mutex.Lock()
	defer info.mutex.Unlock()

	idx, found := info.idx[link]
	if !found {
		return fmt.Errorf("%s not found", link)
	}

	req := netlink.NewIfAddrMessage()
	req.Header.Type = netlink.RTM_NEWADDR
	req.Header.Flags = netlink.NLM_F_CREATE
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

	info.req <- req
	return nil
}

func (info *Info) setLinkAttr(link, attr, value string) error {
	info.mutex.Lock()
	defer info.mutex.Unlock()

	idx, found := info.idx[link]
	if !found {
		return fmt.Errorf("%s not found", link)
	}

	dev := info.dev[idx]

	req := netlink.NewIfInfoMessage()
	req.Header.Type = netlink.RTM_SETLINK
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

	info.req <- req
	return nil
}
