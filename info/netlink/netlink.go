// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package netlink

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/go/info"
	"github.com/platinasystems/go/netlink"
	"github.com/platinasystems/go/redis"
)

const (
	Name = "netlink"

	withCounters    = true
	withoutCounters = false
)

type Info struct {
	mutex     sync.Mutex
	prefixes  []string
	reqch     chan<- netlink.Message
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

func New() *Info { return &Info{prefixes: []string{"lo"}} }

func (p *Info) String() string { return Name }
func (p *Info) Main(...string) error {
	p.idx = make(map[string]uint32)
	p.dev = make(map[uint32]*dev)
	p.efilter = make(map[uint32]struct{})
	p.indices = make([]uint32, len(p.prefixes))

	for i, prefix := range p.prefixes {
		name := strings.TrimSuffix(prefix, ".")
		itf, err := net.InterfaceByName(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return err
		}
		idx := uint32(itf.Index)
		p.idx[name] = uint32(itf.Index)
		p.dev[idx] = &dev{
			name:  name,
			addrs: make(map[string]string),
			attrs: make(map[string]string),
		}
		p.indices[i] = idx
	}

	for i := range p.addrNames {
		s := netlink.IfAddrAttrKind(i).String()
		lc := strings.ToLower(s)
		p.addrNames[i] = strings.Replace(lc, " ", "_", -1)
	}
	for i := range p.attrNames {
		s := netlink.IfInfoAttrKind(i).String()
		lc := strings.ToLower(s)
		p.attrNames[i] = strings.Replace(lc, " ", "_", -1)
	}
	for i := range p.statNames {
		s := netlink.LinkStatType(i).String()
		lc := strings.ToLower(s)
		p.statNames[i] = strings.Replace(lc, " ", "_", -1)
	}

	p.getCounters()
	p.getLinkAddr()

	return nil
}

func (p *Info) Close() error {
	close(p.getCountersStop)
	close(p.getLinkAddrStop)
	return nil
}

func (p *Info) Prefixes(prefixes ...string) []string {
	if len(prefixes) > 0 {
		p.prefixes = prefixes
	}
	return p.prefixes
}

func (p *Info) Del(key string) error {
	return p.delLinkAddr(redis.Split(key))
}

func (p *Info) Set(key, value string) (err error) {
	a := redis.Split(key)
	switch {
	case len(a) == 1:
		err = p.newLinkAddr(a[0], value, "", "")
	case strings.ContainsAny(a[1], ".:"):
		var attr string
		if len(a) > 2 {
			attr = a[2]
		}
		err = p.newLinkAddr(a[0], a[1], attr, value)
	default:
		err = p.setLinkAttr(a[0], a[1], value)
	}
	return
}

// getLinkAddr listens to Netlink multicast groups for link info (besides
// counters) and address changes.
func (p *Info) getLinkAddr() {
	stop := make(chan struct{})
	p.getLinkAddrStop = stop
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
					p.logerr(t)
				}
			case *netlink.IfInfoMessage:
				p.ifInfo(t, withoutCounters)
			case *netlink.IfAddrMessage:
				switch t.Header.Type {
				case netlink.RTM_DELADDR:
					p.ifDelAddr(t)
				case netlink.RTM_NEWADDR:
					p.ifNewAddr(t)
				}
			default:
				fmt.Fprintln(os.Stderr, desc, "unexpected: ",
					msg)
			}
			msg.Close()
		}
	}(rxch)
}

func (p *Info) ifDelAddr(msg *netlink.IfAddrMessage) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	idx := msg.Index
	dev, found := p.dev[idx]
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

		k := fmt.Sprint(cidr, ".", p.addrNames[i])

		_, found := dev.addrs[k]
		if found {
			fullname := fmt.Sprint(dev.name, ".", k)
			info.Publish("delete", fullname)
			delete(dev.addrs, k)
		}
	}
}

func (p *Info) ifNewAddr(msg *netlink.IfAddrMessage) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	idx := msg.Index
	dev, found := p.dev[idx]
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

		k := fmt.Sprint(cidr, ".", p.addrNames[i])
		s := attr.String()

		as, found := dev.addrs[k]
		if !found || s != as {
			fullname := fmt.Sprint(dev.name, ".", k)
			info.Publish(fullname, s)
			dev.addrs[k] = s
		}
	}
}

// getCounters uses a periodic ticker to request link counters.
// The same request socket is used for SETLINK, NEWADDR and DELADDR.
func (p *Info) getCounters() {
	stop := make(chan struct{})
	p.getCountersStop = stop
	rxch := make(chan netlink.Message, 64)
	sock, err := netlink.New(rxch, netlink.NOOP_RTNLGRP)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	go sock.Listen(netlink.NoopListenReq)
	go p.getCounterTicker(sock, stop)
	go p.getCounterRx(rxch)
	return
}

func (p *Info) getCounterTicker(sock *netlink.Socket, wait <-chan struct{}) {
	defer sock.Close()
	seq := uint32(1000000)
	reqch := make(chan netlink.Message, 4)
	p.reqch = reqch
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

func (p *Info) getCounterRx(rxch <-chan netlink.Message) {
	for msg := range rxch {
		switch t := msg.(type) {
		case *netlink.DoneMessage:
		case *netlink.ErrorMessage:
			if t.Errno != 0 {
				p.logerr(t)
			}
		case *netlink.IfInfoMessage:
			p.ifInfo(t, withCounters)
		default:
			fmt.Fprintln(os.Stderr, "GETLINK unexpected: ", msg)
		}
		msg.Close()
	}
}

func (p *Info) ifInfo(msg *netlink.IfInfoMessage, counters bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

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

	dev, found := p.dev[idx]
	if !found {
		return
	}

	if len(name) > 0 && name != dev.name {
		// renamed
		delete(p.idx, dev.name)
		dev.name = name
		p.idx[dev.name] = idx
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
		fullname := fmt.Sprint(dev.name, ".", flag.name)
		if counters {
			// don't import flags from timed counters responses
		} else if dev.flags&flag.bit == flag.bit {
			if !found {
				dev.attrs[flag.name] = flag.show
				info.Publish(fullname, flag.show)
			}
		} else if found {
			delete(dev.attrs, flag.name)
			info.Publish("delete", fullname)
		}
	}
	for i, attr := range msg.Attrs {
		if attr == nil {
			continue
		}
		k := netlink.IfInfoAttrKind(i)
		aname := p.attrNames[i]
		switch k {
		case netlink.IFLA_STATS:
		case netlink.IFLA_AF_SPEC:
		case netlink.IFLA_STATS64:
			if !counters {
				continue
			}
			for i, count := range attr.(*netlink.LinkStats64) {
				if count > dev.stats[i] {
					dev.stats[i] = count
					fullname := fmt.Sprint(dev.name, ".",
						p.statNames[i])
					info.Publish(fullname, count)
				}
			}
		default:
			if counters {
				continue
			}
			s := attr.String()
			as, found := dev.attrs[aname]
			if !found || s != as {
				fullname := fmt.Sprint(dev.name, ".", aname)
				info.Publish(fullname, s)
				dev.attrs[aname] = s
				break
			}
		}
	}
}

func (p *Info) logerr(msg *netlink.ErrorMessage) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	seq := msg.Header.Sequence
	_, found := p.efilter[seq]
	if !found {
		fmt.Fprintln(os.Stderr, msg)
		p.efilter[seq] = struct{}{}
	}
}

// args: LINK, CIDR
func (p *Info) delLinkAddr(args []string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if len(args) != 2 {
		return fmt.Errorf("%v invalid", args)
	}

	idx, found := p.idx[args[0]]
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

	p.reqch <- req
	return nil
}

func (p *Info) newLinkAddr(link, cidr, attr, value string) error {
	var addrAttr netlink.Attr

	p.mutex.Lock()
	defer p.mutex.Unlock()

	idx, found := p.idx[link]
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

	p.reqch <- req
	return nil
}

func (p *Info) setLinkAttr(link, attr, value string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	idx, found := p.idx[link]
	if !found {
		return fmt.Errorf("%s not found", link)
	}

	dev := p.dev[idx]

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

	p.reqch <- req
	return nil
}
