// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"unsafe"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Apropos = "route table entry"
	Man     = `
SEE ALSO
	ip man route || ip route -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	fixme = errors.New("FIXME")
)

func New(s string) *Command { return &Command{name: s} }

type Command struct {
	name string
	opt  *options.Options
	args []string

	sr *rtnl.SockReceiver

	hdr   rtnl.Hdr
	msg   rtnl.RtMsg
	attrs rtnl.Attrs

	ifindexByName map[string]int32
	vrfByName     map[string]uint32
}

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Man() lang.Alt     { return man }
func (c *Command) String() string  { return c.name }
func (c *Command) Usage() string {
	return fmt.Sprint("ip route ", c, ` NODE-SPEC [ INFO-SPEC ]

NODE-SPEC := [ TYPE ] PREFIX [ tos TOS ] [ table RTTABLE ]
	[ proto RTPROTO ] [ scope RTSCOPE ] [ metric RTMETRIC ]
	[ [no-]ttl-propagate ]

TYPE := { unicast | local | broadcast | multicast | throw | unreachable |
	prohibit | blackhole | nat }

RTTABLE:= { compate | default | main | local | all | NUMBER }

RTPROTO := { redirect | kernel | boot | static | gated | ra | mrt | zebra |
	bird | dnrouted | xorp | ntk | dhcp | mrouted | babel | NUMBER }

RTSCOPE := { global | site | link | host | NUMBER }

INFO-SPEC := NH OPTIONS [ nexthop NH ] ...

NH := [ encap ENCAP ] [ via [ FAMILY ] ADDRESS ] [ dev IFNAME ]
	[ weight WEIGHT ] [ onlink | pervasive ]

OPTIONS := [ as [ to ] ADDRESS ]
	[ mtu NUMBER ] [ advmss NUMBER ] [ expires NUMBER ]
	[ reordering NUMBER ] [ window NUMBER ] [ cwnd NUMBER ]
	[ initcwnd NUMBER ] [ initrwnd NUMBER ] [ ssthresh NUMBER ]
	[ rtt TIME ] [ rttvar TIME ] [ rto_min TIME ]
	[ realms REALM ]
	[ features FEATURES ]
	[ quickack BOOLEAN ]
	[ congctl NAME ]
	[ pref { low | medium | high } ]
	[ onlink ]

FEATURES := { ecn }

ENCAP := { ENCAP-MPLS | ENCAP-IP | ENCAP-IP6 | ENCAP-ILA | ENCAP-SEG6 |
	ENCAP-BFP }

ENCAP-MPLS := mpls [ LABEL ] [ ttl TTL ]

ENCAP-IP := ip id TUNNEL-ID dst REMOTE-IP [ tos TOS ] [ ttl TTL ]

ENCAP-IP6 := id TUNNEL-ID dst REMOTE-IP [ tc TC ] [ hoplimit HOPS ]

ENCAP-ILA := LOCATOR [ csum-mode { adj-transport | neutral-map | no-action } ]

ENCAP-SEG6 := seg6 mode [ encap | inline ] segs SEGMENTS [ hmac KEYID ]

ENCAP-BPF := bpf [ in PROG ] [ out PROG ] [ xmit PROG ] [ headroom SIZE ]`)
}

func (c *Command) Main(args ...string) error {
	var err error

	if args, err = options.Netns(args); err != nil {
		return err
	}

	c.opt, c.args = options.New(args)

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	c.sr = rtnl.NewSockReceiver(sock)

	if err = c.getifindices(); err != nil {
		return err
	}

	c.hdr.Flags = rtnl.NLM_F_REQUEST | rtnl.NLM_F_ACK

	switch c.name {
	case "add":
		c.hdr.Type = rtnl.RTM_NEWROUTE
		c.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_EXCL
	case "append":
		c.hdr.Type = rtnl.RTM_NEWROUTE
		c.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_APPEND
	case "change", "set":
		c.hdr.Type = rtnl.RTM_NEWROUTE
		c.hdr.Flags |= rtnl.NLM_F_REPLACE
	case "prepend":
		c.hdr.Type = rtnl.RTM_NEWROUTE
		c.hdr.Flags |= rtnl.NLM_F_CREATE
	case "replace":
		c.hdr.Type = rtnl.RTM_NEWROUTE
		c.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_REPLACE
	case "delete":
		c.hdr.Type = rtnl.RTM_DELROUTE
	case "test":
		c.hdr.Type = rtnl.RTM_NEWROUTE
		c.hdr.Flags |= rtnl.NLM_F_EXCL
	default:
		return fmt.Errorf("%s: unknown", c)
	}

	if c.name != "delete" {
		c.msg.Protocol = rtnl.RTPROT_BOOT
		c.msg.Scope = rtnl.RT_SCOPE_UNIVERSE
		c.msg.Type = rtnl.RTN_UNICAST
	}

	if err = c.parse(); err != nil {
		return err
	}

	req, err := rtnl.NewMessage(c.hdr, c.msg, c.attrs...)
	if err == nil {
		err = c.sr.UntilDone(req, func([]byte) {})
	}
	return err
}

func (c *Command) append(t uint16, v io.Reader) {
	c.attrs = append(c.attrs, rtnl.Attr{t, v})
}

func (c *Command) getifindices() error {
	c.ifindexByName = make(map[string]int32)
	c.vrfByName = make(map[string]uint32)

	req, err := rtnl.NewMessage(
		rtnl.Hdr{
			Type:  rtnl.RTM_GETLINK,
			Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_DUMP,
		},
		rtnl.IfInfoMsg{
			Family: rtnl.AF_UNSPEC,
		},
	)
	if err != nil {
		return err
	}
	return c.sr.UntilDone(req, func(b []byte) {
		var ifla rtnl.Ifla
		if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWLINK {
			return
		}
		msg := rtnl.IfInfoMsgPtr(b)
		ifla.Write(b)
		name := rtnl.Kstring(ifla[rtnl.IFLA_IFNAME])
		c.ifindexByName[name] = msg.Index
		if rtnl.Kstring(ifla[rtnl.IFLA_INFO_KIND]) == "vrf" {
			c.vrfByName[name] =
				rtnl.Uint32(ifla[rtnl.IFLA_VRF_TABLE])
		}
	})
}

func (c *Command) parse() error {
	var (
		err     error
		mxlock  rtnl.Uint32Attr
		mxattrs rtnl.Attrs
	)
	mxappend := func(t uint16, v io.Reader) {
		mxattrs = append(mxattrs, rtnl.Attr{t, v})
	}
	if s := c.opt.Parms.ByName["-f"]; len(s) > 0 {
		if v, ok := rtnl.AfByName[s]; ok {
			c.msg.Family = v
		} else {
			return fmt.Errorf("family: %q unknown", s)
		}
	}
	if len(c.args) > 0 {
		if v, ok := rtnl.RtnByName[c.args[0]]; ok {
			c.msg.Type = v
			c.args = c.args[1:]
		}
	}

	prefix, err := c.parsePrefix(c.msg.Family)
	if err != nil {
		return err
	}
	c.msg.Family = prefix.Family()
	c.msg.Dst_len = prefix.Len()
	c.append(rtnl.RTA_DST, prefix)
	c.msg.Scope = rtnl.RT_SCOPE_UNIVERSE

	for err == nil && len(c.args) > 0 {
		arg0 := c.args[0]
		c.args = c.args[1:]
		switch arg0 {
		case "src":
			if v, e := c.parseAddress(c.msg.Family); e == nil {
				c.append(rtnl.RTA_PREFSRC, v)
			} else {
				err = e
			}
		case "as":
			if v, e := c.parseAs(); err == nil {
				c.append(rtnl.RTA_NEWDST, v)
			} else {
				err = e
			}
		case "via":
			if v, e := c.parseVia(); e == nil {
				if c.msg.Family == v.Family() {
					c.append(rtnl.RTA_GATEWAY, v)
				} else {
					c.append(rtnl.RTA_VIA,
						rtnl.RtVia{uint16(v.Family()),
							v.Bytes()})
				}
			} else {
				err = e
			}
		case "from":
			if v, e := c.parsePrefix(c.msg.Family); e == nil {
				if c.msg.Family == rtnl.AF_UNSPEC {
					c.msg.Family = v.Family()
				}
				c.append(rtnl.RTA_SRC, v)
			} else {
				err = e
			}
		case "tos", "dstfield":
			err = c.parseTos()
		case "scope":
			err = c.parseScope()
		case "expires", "metric", "priority":
			t := map[string]uint16{
				"expires":  rtnl.RTA_EXPIRES,
				"metric":   rtnl.RTA_PRIORITY,
				"priority": rtnl.RTA_PRIORITY,
			}[arg0]
			if v, e := c.parseNumber(); e == nil {
				c.append(t, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "mtu", "hoplimit", "advmss", "reordering",
			"window", "cwnd", "initcwnd", "initrwnd":
			// [ lock ] NUMBER
			t := map[string]uint16{
				"mtu":        rtnl.RTAX_MTU,
				"hoplimit":   rtnl.RTAX_HOPLIMIT,
				"advmss":     rtnl.RTAX_ADVMSS,
				"reordering": rtnl.RTAX_REORDERING,
				"window":     rtnl.RTAX_WINDOW,
				"cwnd":       rtnl.RTAX_CWND,
				"initcwnd":   rtnl.RTAX_INITCWND,
				"initrwnd":   rtnl.RTAX_INITRWND,
				"ssthresh":   rtnl.RTAX_SSTHRESH,
			}[arg0]
			if len(c.args) > 0 && c.args[0] == "lock" {
				mxlock |= 1 << t
				c.args = c.args[1:]
			}
			if v, e := c.parseNumber(); e == nil {
				mxappend(t, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "congctl": // [ lock ] STRING
			if len(c.args) > 0 && c.args[0] == "lock" {
				mxlock |= 1 << rtnl.RTAX_CC_ALGO
				c.args = c.args[1:]
			}
			if v, e := c.parseString(); e == nil {
				mxappend(rtnl.RTAX_CC_ALGO,
					rtnl.KstringAttr(v))
			} else {
				err = e
			}
		case "rtt":
			if len(c.args) > 0 && c.args[0] == "lock" {
				mxlock |= 1 << rtnl.RTAX_RTT
				c.args = c.args[1:]
			}
			if v, e := c.parseRtt(8); e == nil {
				mxappend(rtnl.RTAX_RTT, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "rto-min", "rto_min":
			if v, e := c.parseRtt(1); e == nil {
				mxappend(rtnl.RTAX_RTO_MIN, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "rttvar":
			if len(c.args) > 0 && c.args[0] == "lock" {
				mxlock |= 1 << rtnl.RTAX_RTTVAR
				c.args = c.args[1:]
			}
			if v, e := c.parseRtt(4); e == nil {
				mxappend(rtnl.RTAX_RTTVAR, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "quickack":
			var v rtnl.Uint32Attr
			if len(c.args) > 0 {
				if c.args[0] == "1" ||
					c.args[0] == "t" ||
					c.args[0] == "true" {
					v = 1
				}
			} else {
				err = fmt.Errorf("missing BOOLEAN")
			}
			if err == nil {
				mxappend(rtnl.RTAX_QUICKACK,
					rtnl.Uint32Attr(v))
			}
		case "features":
			var features uint32
			if len(c.args) > 0 {
				switch c.args[0] {
				case "ecn:":
					features |= rtnl.RTAX_FEATURE_ECN
				default:
					err = fmt.Errorf("feature: %q unknown",
						c.args[0])
				}
				c.args = c.args[1:]
			}
			mxappend(rtnl.RTAX_FEATURES, rtnl.Uint32Attr(features))
		case "realms":
			if v, e := c.parseRealm(); e == nil {
				c.append(rtnl.RTA_FLOW, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "onlink":
			c.msg.Flags |= uint32(rtnl.RTNH_F_ONLINK)
		case "nexthop":
			if nhs, e := c.parseNextHops(); e == nil {
				c.append(rtnl.RTA_MULTIPATH, nhs)
			} else {
				err = e
			}
		case "prot", "protocol":
			err = c.parseProtocol()
		case "table":
			err = c.parseTable()
		case "vrf":
			err = c.parseVrf()
		case "dev", "oif":
			if v, e := c.parseIfname(); e == nil {
				c.append(rtnl.RTA_OIF, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "pref", "preference":
			if v, e := c.parsePreference(); e == nil {
				c.append(rtnl.RTA_PREF, rtnl.Uint8Attr(v))
			} else {
				err = e
			}
		case "encap":
			if v, e := c.parseEncap(); e == nil {
				c.append(rtnl.RTA_ENCAP, v)
			} else {
				err = e
			}
		case "ttl-propagate", "+ttl-propagate":
			c.append(rtnl.RTA_TTL_PROPAGATE, rtnl.Uint8Attr(1))
		case "no-ttl-propagate", "-ttl-propagate":
			c.append(rtnl.RTA_TTL_PROPAGATE, rtnl.Uint8Attr(0))
		default:
			err = fmt.Errorf("unexpected")
		}
		if err != nil {
			err = fmt.Errorf("%s: %s", arg0, err)
		}
	}
	if mxlock != 0 {
		c.append(rtnl.RTAX_LOCK, rtnl.Uint32Attr(mxlock))
	}
	if len(mxattrs) > 0 {
		c.append(rtnl.RTA_METRICS, mxattrs)
	}
	return err
}

func (c *Command) parseString() (string, error) {
	var v string
	if len(c.args) == 0 {
		return v, fmt.Errorf("missing STRING")
	}
	v = c.args[0]
	c.args = c.args[1:]
	return v, nil
}

func (c *Command) parseNumber() (int64, error) {
	var v int64
	if len(c.args) == 0 {
		return v, fmt.Errorf("missing NUMBER")
	}
	if _, err := fmt.Sscan(c.args[0], &v); err != nil {
		return v, fmt.Errorf("%q %v", c.args[0], err)
	}
	c.args = c.args[1:]
	return v, nil
}

func (c *Command) parseRtt(rawfactor int64) (int64, error) {
	var v int64
	if len(c.args) == 0 {
		return v, fmt.Errorf("missing NUMBER")
	}
	arg0 := c.args[0]
	c.args = c.args[1:]
	factor := rawfactor
	for _, suffix := range []string{"secs", "sec", "s"} {
		if strings.HasSuffix(arg0, suffix) {
			strings.TrimSuffix(arg0, suffix)
			factor = 1000
			break
		}
	}
	for _, suffix := range []string{"msecs", "msec", "ms"} {
		if strings.HasSuffix(arg0, suffix) {
			strings.TrimSuffix(arg0, suffix)
			factor = 1
			break
		}
	}
	if _, err := fmt.Sscan(c.args[0], &v); err != nil {
		return v, fmt.Errorf("%q %v", c.args[0], err)
	}
	return v * factor, nil
}

func (c *Command) parsePrefix(family uint8) (rtnl.Prefixer, error) {
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing PREFIX")
	}
	prefix, err := rtnl.Prefix(c.args[0], family)
	c.args = c.args[1:]
	return prefix, err
}

func (c *Command) parseAddress(family uint8) (rtnl.Addresser, error) {
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing ADDRESS")
	}
	addr, err := rtnl.Address(c.args[0], family)
	c.args = c.args[1:]
	return addr, err
}

func (c *Command) parseTos() error {
	if len(c.args) == 0 {
		return fmt.Errorf("missing TOS")
	}
	if _, err := fmt.Sscan(c.args[0], &c.msg.Tos); err != nil {
		return fmt.Errorf("%q %v", c.args[0], err)
	}
	c.args = c.args[1:]
	return nil
}

func (c *Command) parseTable() error {
	var t rtnl.Uint32Attr
	if len(c.args) == 0 {
		return fmt.Errorf("missing RTTABLE")
	}
	if v, ok := rtnl.RtTableByName[c.args[0]]; ok {
		c.msg.Table = uint8(v)
	} else if _, err := fmt.Sscan(c.args[0], &t); err != nil {
		return fmt.Errorf("%q %v", c.args[0], err)
	} else if t < 256 {
		c.msg.Table = uint8(t)
	} else {
		c.msg.Table = uint8(rtnl.RT_TABLE_UNSPEC)
		c.append(rtnl.RTA_TABLE, t)
	}
	c.args = c.args[1:]
	return nil
}

func (c *Command) parseVrf() error {
	if len(c.args) == 0 {
		return fmt.Errorf("missing VRF")
	}
	if vrf, found := c.vrfByName[c.args[0]]; !found {
		return fmt.Errorf("%q no found", c.args[0])
	} else if vrf < 256 {
		c.msg.Table = uint8(vrf)
	} else {
		c.msg.Table = uint8(rtnl.RT_TABLE_UNSPEC)
		c.append(rtnl.RTA_TABLE, rtnl.Uint32Attr(vrf))
	}
	c.args = c.args[1:]
	return nil
}

func (c *Command) parsePreference() (uint8, error) {
	if len(c.args) == 0 {
		return 0, fmt.Errorf("missing { low | medium | high }")
	}
	pref, ok := map[string]uint8{
		"low":    rtnl.ICMPV6_ROUTER_PREF_LOW,
		"med":    rtnl.ICMPV6_ROUTER_PREF_MEDIUM,
		"medium": rtnl.ICMPV6_ROUTER_PREF_MEDIUM,
		"hi":     rtnl.ICMPV6_ROUTER_PREF_HIGH,
		"high":   rtnl.ICMPV6_ROUTER_PREF_HIGH,
	}[c.args[0]]
	if !ok {
		return 0, fmt.Errorf("%q invalid", c.args[0])
	}
	c.args = c.args[1:]
	return pref, nil
}

func (c *Command) parseProtocol() error {
	if len(c.args) == 0 {
		return fmt.Errorf("missing RTPROTOCOL")
	}
	if v, ok := rtnl.RtProtByName[c.args[0]]; ok {
		c.msg.Protocol = v
	} else if _, err := fmt.Sscan(c.args[0], &c.msg.Protocol); err != nil {
		return fmt.Errorf("%q %v", c.args[0], err)
	}
	c.args = c.args[1:]
	return nil
}

func (c *Command) parseScope() error {
	if len(c.args) == 0 {
		return fmt.Errorf("missing RTSCOPE")
	}
	if v, ok := rtnl.RtScopeByName[c.args[0]]; ok {
		c.msg.Scope = v
	} else {
		return fmt.Errorf("%q unknown", c.args[0])
	}
	c.args = c.args[1:]
	return nil
}

func (c *Command) parseEncap() (io.Reader, error) {
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing ENCAP")
	}
	arg0 := c.args[0]
	c.args = c.args[1:]
	switch arg0 {
	case "mpls":
		return c.parseEncapMpls()
	case "ip":
		return c.parseEncapIp()
	case "ip6":
		return c.parseEncapIp6()
	case "ila":
		return c.parseEncapIla()
	case "bpf":
		return c.parseEncapBpf()
	case "seg6":
		return c.parseEncapSeg6()
	}
	return nil, fmt.Errorf("%q unknown", arg0)
}

func (c *Command) parseVia() (rtnl.Addresser, error) {
	family := c.msg.Family
	mia := fmt.Errorf("missing ADDRESS")
	if len(c.args) == 0 {
		return nil, mia
	}
	if viaFamily, ok := rtnl.AfByName[c.args[0]]; ok {
		family = viaFamily
		c.args = c.args[1:]
	}
	if len(c.args) == 0 {
		return nil, mia
	}
	addr, err := rtnl.Address(c.args[0], family)
	if err != nil {
		return nil, err
	}
	c.args = c.args[1:]
	if c.msg.Family == rtnl.AF_UNSPEC {
		c.msg.Family = addr.Family()
	}
	return addr, nil
}

func (c *Command) parseIfname() (int, error) {
	if len(c.args) == 0 {
		return -1, fmt.Errorf("missing IFNAME")
	}
	ifindex, found := c.ifindexByName[c.args[0]]
	if !found {
		return -1, fmt.Errorf("%q not found", c.args[0])
	}
	c.args = c.args[1:]
	return int(ifindex), nil
}

func (c *Command) parseWeight() (uint8, error) {
	var u8 uint8
	if len(c.args) == 0 {
		return u8, fmt.Errorf("missing WEIGHT")
	}
	if _, err := fmt.Sscan(c.args[0], &u8); err != nil {
		return 0, err
	} else if u8 < 1 {
		return 0, fmt.Errorf("must be >= 1")
	}
	c.args = c.args[1:]
	return u8 - 1, nil
}

func (c *Command) parseRealm() (uint32, error) {
	if len(c.args) == 0 {
		return 0, fmt.Errorf("missing REALM")
	}
	arg0 := c.args[0]
	c.args = c.args[1:]
	u8, found := rtnl.RtnByName[arg0]
	if found {
		return uint32(u8), nil
	}
	var realm uint32
	if _, err := fmt.Sscan(arg0, &realm); err != nil {
		return 0, err
	}
	return realm, nil
}

// [to] ADDRESS
func (c *Command) parseAs() (rtnl.Addresser, error) {
	if len(c.args) >= 1 && c.args[0] == "to" {
		c.args = c.args[1:]
	}
	return c.parseAddress(c.msg.Family)
}

// NH [ nexthop NH... ]
// NH := [ encap ENCAP ] [ via [ FAMILY ] ADDRESS ] [ dev IFNAME ]
//	[ weight WEIGHT ] [ onlink | pervasive ]
func (c *Command) parseNextHops() (rtnl.RtnhAttrsList, error) {
	var (
		err error
		nh  rtnl.RtnhAttrs
	)
	nhappend := func(t uint16, v io.Reader) {
		nh.Attrs = append(nh.Attrs, rtnl.Attr{t, v})
	}
	nhs := rtnl.RtnhAttrsList{nh}
nhloop:
	for err == nil && len(c.args) > 0 {
		arg0 := c.args[0]
		c.args = c.args[1:]
		switch arg0 {
		case "nexthop":
			// recurse
			if more, e := c.parseNextHops(); e == nil {
				nhs = append(nhs, more...)
				break nhloop
			} else {
				err = e
			}
		case "via":
			if v, e := c.parseVia(); e == nil {
				if c.msg.Family == v.Family() {
					nhappend(rtnl.RTA_GATEWAY, v)
				} else {
					nhappend(rtnl.RTA_VIA,
						rtnl.RtVia{uint16(v.Family()),
							v.Bytes()})
				}
			} else {
				err = e
			}
		case "dev", "oif":
			if len(c.args) == 0 {
				err = fmt.Errorf("missing IFNAME")
			} else if i, ok := c.ifindexByName[c.args[0]]; !ok {
				err = fmt.Errorf("%q not found", c.args[0])
			} else {
				nh.Ifindex = int(i)
				c.args = c.args[1:]
			}
		case "weight":
			if v, e := c.parseWeight(); e == nil {
				nh.Rtnh.Hops = v
			} else {
				err = e
			}
		case "onlink":
			nh.Rtnh.Flags |= rtnl.RTNH_F_ONLINK
		case "realm":
			if v, e := c.parseRealm(); e == nil {
				nhappend(rtnl.RTA_FLOW, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "as":
			if v, e := c.parseAs(); e == nil {
				nhappend(rtnl.RTA_NEWDST, v)
			} else {
				err = e
			}
		default:
			err = fmt.Errorf("unexpected")
		}
		if err != nil {
			err = fmt.Errorf("%s: %v", arg0, err)
			break
		}
	}
	return nhs, err
}

// LABEL [ ttl TTL ]
func (c *Command) parseEncapMpls() (io.Reader, error) {
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing LABEL")
	}
	addr, err := rtnl.Address(c.args[0], rtnl.AF_MPLS)
	if err != nil {
		return nil, err
	}
	attrs := rtnl.Attrs{rtnl.Attr{rtnl.MPLS_IPTUNNEL_DST, addr}}
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return attrs, nil
	}
	if c.args[0] != "ttl" {
		return nil, fmt.Errorf("%s: unexpected", c.args[0])
	}
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing TTL")
	}
	var ttl uint8
	if _, err = fmt.Sscan(c.args[0], &ttl); err == nil {
		return nil, fmt.Errorf("ttl: %v", err)
	}
	attrs = append(attrs,
		rtnl.Attr{rtnl.MPLS_IPTUNNEL_TTL, rtnl.Uint8Attr(ttl)})
	c.args = c.args[1:]
	return attrs, nil
}

// id TUNNEL-ID dst REMOTE-IP [ tos TOS ] [ ttl TTL ]
func (c *Command) parseEncapIp() (io.Reader, error) {
	var attrs rtnl.Attrs
	appendAttr := func(t uint16, v io.Reader) {
		attrs = append(attrs, rtnl.Attr{t, v})
	}
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing id")
	}
	if c.args[0] != "id" {
		return nil, fmt.Errorf("%s: unexpected", c.args[0])
	}
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing TUNNEL-ID")
	}
	var id uint64
	if _, err := fmt.Sscan(c.args[0], &id); err == nil {
		return nil, fmt.Errorf("id: %v", err)
	}
	appendAttr(rtnl.LWTUNNEL_IP_ID, rtnl.Uint64Attr(id))
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing dst")
	}
	if c.args[0] != "dst" {
		return nil, fmt.Errorf("%s: unexpected", c.args[0])
	}
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing REMOTE-IP")
	}
	if addr, err := rtnl.Address(c.args[0], rtnl.AF_INET); err != nil {
		return nil, fmt.Errorf("dst: %v", err)
	} else {
		appendAttr(rtnl.LWTUNNEL_IP_DST, addr)
	}
	c.args = c.args[1:]
	for len(c.args) > 0 {
		switch c.args[0] {
		case "tos":
			var tos uint32
			c.args = c.args[1:]
			if len(c.args) == 0 {
				return nil, fmt.Errorf("missing TOS")
			}
			// FIXME symbolic TOS
			if _, err := fmt.Sscan(c.args[0], &tos); err != nil {
				return nil, fmt.Errorf("tos: %v", err)
			}
			appendAttr(rtnl.LWTUNNEL_IP_TOS, rtnl.Uint32Attr(tos))
			c.args = c.args[1:]
		case "ttl":
			var ttl uint8
			c.args = c.args[1:]
			if len(c.args) == 0 {
				return nil, fmt.Errorf("missing TTL")
			}
			if _, err := fmt.Sscan(c.args[0], &ttl); err != nil {
				return nil, fmt.Errorf("ttl: %v", err)
			}
			appendAttr(rtnl.LWTUNNEL_IP_TTL, rtnl.Uint8Attr(ttl))
			c.args = c.args[1:]
		default:
			return attrs, nil
		}
	}
	return attrs, nil
}

// id TUNNEL-ID dst REMOTE-IP [ tc TC ] [ hoplimit HOPS ]
func (c *Command) parseEncapIp6() (io.Reader, error) {
	var attrs rtnl.Attrs
	appendAttr := func(t uint16, v io.Reader) {
		attrs = append(attrs, rtnl.Attr{t, v})
	}
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing id")
	}
	if c.args[0] != "id" {
		return nil, fmt.Errorf("%s: unexpected", c.args[0])
	}
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing TUNNEL-ID")
	}
	var id uint64
	if _, err := fmt.Sscan(c.args[0], &id); err == nil {
		return nil, fmt.Errorf("id: %v", err)
	}
	appendAttr(rtnl.LWTUNNEL_IP6_ID, rtnl.Uint64Attr(id))
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing dst")
	}
	if c.args[0] != "dst" {
		return nil, fmt.Errorf("%s: unexpected", c.args[0])
	}
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing REMOTE-IP6")
	}
	if addr, err := rtnl.Address(c.args[0], rtnl.AF_INET6); err != nil {
		return nil, fmt.Errorf("dst: %v", err)
	} else {
		appendAttr(rtnl.LWTUNNEL_IP6_DST, addr)
	}
	c.args = c.args[1:]
	for len(c.args) > 0 {
		switch c.args[0] {
		case "tc":
			var tc uint8
			c.args = c.args[1:]
			if len(c.args) == 0 {
				return nil, fmt.Errorf("missing TC")
			}
			// FIXME symbolic TOS
			if _, err := fmt.Sscan(c.args[0], &tc); err != nil {
				return nil, fmt.Errorf("tc: %v", err)
			}
			appendAttr(rtnl.LWTUNNEL_IP6_TC, rtnl.Uint8Attr(tc))
			c.args = c.args[1:]
		case "ttl":
			var hops uint8
			c.args = c.args[1:]
			if len(c.args) == 0 {
				return nil, fmt.Errorf("missing HOPS")
			}
			if _, err := fmt.Sscan(c.args[0], &hops); err != nil {
				return nil, fmt.Errorf("tops: %v", err)
			}
			appendAttr(rtnl.LWTUNNEL_IP6_HOPLIMIT,
				rtnl.Uint8Attr(hops))
			c.args = c.args[1:]
		default:
			return attrs, nil
		}
	}
	return attrs, nil
}

// LOCATOR [ csum-mode { adj-transport | neutral-map | no-action } ]
func (c *Command) parseEncapIla() (io.Reader, error) {
	var attrs rtnl.Attrs
	appendAttr := func(t uint16, v io.Reader) {
		attrs = append(attrs, rtnl.Attr{t, v})
	}
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing LOCATOR")
	}
	var locator uint64
	if _, err := fmt.Sscan(c.args[0], &locator); err == nil {
		return nil, fmt.Errorf("locator: %v", err)
	}
	appendAttr(rtnl.ILA_ATTR_LOCATOR, rtnl.Uint64Attr(locator))
	c.args = c.args[1:]
	if len(c.args) > 0 && c.args[0] == "csum-mode" {
		c.args = c.args[1:]
		if len(c.args) == 0 {
			return nil, fmt.Errorf("missing CSUM-MODE")
		}
		mode, found := rtnl.IlaCsumModeByName[c.args[0]]
		if !found {
			return nil, fmt.Errorf("csum-mode: %q invalid",
				c.args[0])
		}
		appendAttr(rtnl.ILA_ATTR_CSUM_MODE, rtnl.Uint8Attr(mode))
		c.args = c.args[1:]
	}
	return attrs, nil
}

// mode { encap | inline } segs SEGMENTS [ hmac KEYID ]
func (c *Command) parseEncapSeg6() (io.Reader, error) {
	var segs []net.IP
	var hmac uint32
	var flags uint8
	mode := rtnl.SEG6_IPTUN_MODE_UNSPEC
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing mode")
	}
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing MODE")
	}
	switch c.args[0] {
	case "encap":
		mode = rtnl.SEG6_IPTUN_MODE_ENCAP
	case "inline":
		mode = rtnl.SEG6_IPTUN_MODE_INLINE
	default:
		return nil, fmt.Errorf("mode: %q invalid", c.args[0])
	}
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing segs")
	}
	c.args = c.args[1:]
	if len(c.args) == 0 {
		return nil, fmt.Errorf("missing SEGMENTS")
	}
	for _, s := range strings.Split(c.args[0], ",") {
		seg := net.ParseIP(s)
		if seg.To16() == nil {
			return nil, fmt.Errorf("segment: %q invalid", s)
		}
		segs = append(segs, seg)
	}
	srhlen := 8 + (16 * len(segs))
	c.args = c.args[1:]
	if len(c.args) > 0 && c.args[0] == "hmac" {
		c.args = c.args[1:]
		if len(c.args) == 0 {
			return nil, fmt.Errorf("missing KEYID")
		}
		if _, err := fmt.Sscan(c.args[0], &hmac); err != nil {
			return nil, fmt.Errorf("hmac: %q %v", c.args[0], err)
		}
		flags |= rtnl.SR6_TLV_HMAC
		srhlen += 40
	}

	b := make([]byte, rtnl.SizeofSeg6IpTunnelEncap+srhlen)
	ti := (*rtnl.Seg6IpTunnelEncap)(unsafe.Pointer(&b[0]))
	ti.Mode = mode
	ti.HdrLen = uint8((srhlen >> 3) - 1)
	ti.Type = 4
	ti.SegmentsLeft = uint8(len(segs) - 1)
	ti.FirstSegment = uint8(len(segs) - 1)
	ti.Flags = flags

	for i, seg := range segs {
		copy(b[rtnl.SizeofSeg6IpTunnelEncap+(16*i):], seg)
	}

	if hmac != 0 {
		tlv := (*rtnl.Sr6TlvHmac)(unsafe.Pointer(&b[len(b)-40]))
		tlv.Type = rtnl.SR6_TLV_HMAC
		tlv.Len = 38
		tlv.HmacKeyId.Store(hmac)
	}

	return rtnl.Attr{rtnl.SEG6_IPTUNNEL_SRH, rtnl.BytesAttr(b)}, nil
}

// [ in PROG ] [ out PROG ] [ xmit PROG ] [ headroom SIZE ]
func (c *Command) parseEncapBpf() (io.Reader, error) {
	return nil, fixme
}
