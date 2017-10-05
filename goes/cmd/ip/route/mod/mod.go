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

func New(s string) Command { return Command(s) }

type Command string

type mod struct {
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

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (c Command) String() string  { return string(c) }
func (c Command) Usage() string {
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

func (c Command) Main(args ...string) error {
	var err error
	var m mod

	if args, err = options.Netns(args); err != nil {
		return err
	}

	m.opt, m.args = options.New(args)

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	m.sr = rtnl.NewSockReceiver(sock)

	if err = m.getifindices(); err != nil {
		return err
	}

	m.hdr.Flags = rtnl.NLM_F_REQUEST | rtnl.NLM_F_ACK

	switch c {
	case "add":
		m.hdr.Type = rtnl.RTM_NEWROUTE
		m.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_EXCL
	case "append":
		m.hdr.Type = rtnl.RTM_NEWROUTE
		m.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_APPEND
	case "change", "set":
		m.hdr.Type = rtnl.RTM_NEWROUTE
		m.hdr.Flags |= rtnl.NLM_F_REPLACE
	case "prepend":
		m.hdr.Type = rtnl.RTM_NEWROUTE
		m.hdr.Flags |= rtnl.NLM_F_CREATE
	case "replace":
		m.hdr.Type = rtnl.RTM_NEWROUTE
		m.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_REPLACE
	case "delete":
		m.hdr.Type = rtnl.RTM_DELROUTE
	case "test":
		m.hdr.Type = rtnl.RTM_NEWROUTE
		m.hdr.Flags |= rtnl.NLM_F_EXCL
	default:
		return fmt.Errorf("%s: unknown", c)
	}

	if c != "delete" {
		m.msg.Protocol = rtnl.RTPROT_BOOT
		m.msg.Scope = rtnl.RT_SCOPE_UNIVERSE
		m.msg.Type = rtnl.RTN_UNICAST
	}

	if err = m.parse(); err != nil {
		return err
	}

	req, err := rtnl.NewMessage(m.hdr, m.msg, m.attrs...)
	if err == nil {
		err = m.sr.UntilDone(req, func([]byte) {})
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
	cpv["tos"] = options.NoComplete
	cpv["table"] = options.NoComplete
	cpv["protocol"] = rtnl.CompleteRtProt
	cpv["scope"] = rtnl.CompleteRtScope
	cpv["metric"] = options.NoComplete
	cpv["encap"] = completeEncap
	cpv["via"] = options.NoComplete
	cpv["dev"] = options.CompleteIfName
	cpv["weight"] = options.NoComplete
	cpv["as"] = options.NoComplete
	cpv["mtu"] = options.NoComplete
	cpv["advmss"] = options.NoComplete
	cpv["expires"] = options.NoComplete
	cpv["reordering"] = options.NoComplete
	cpv["window"] = options.NoComplete
	cpv["cwnd"] = options.NoComplete
	cpv["initcwnd"] = options.NoComplete
	cpv["initrwnd"] = options.NoComplete
	cpv["ssthresh"] = options.NoComplete
	cpv["rtt"] = options.NoComplete
	cpv["rttvar"] = options.NoComplete
	cpv["rto_min"] = options.NoComplete
	cpv["realms"] = options.NoComplete
	cpv["features"] = options.NoComplete
	cpv["quickack"] = options.NoComplete
	cpv["congctl"] = options.NoComplete
	cpv["pref"] = options.NoComplete
	if method, found := cpv[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"tos",
			"table",
			"protocol",
			"scope",
			"metric",
			"ttl-propagate",
			"no-ttl-propagate",
			"encap",
			"via",
			"dev",
			"weight",
			"onlink",
			"pervasive",
			"as",
			"mtu",
			"advmss",
			"expires",
			"reordering",
			"window",
			"cwnd",
			"initcwnd",
			"initrwnd",
			"ssthresh",
			"rtt",
			"rttvar",
			"rto_min",
			"realms",
			"features",
			"quickack",
			"congctl",
			"pref",
		) {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	return
}

func (m *mod) append(t uint16, v io.Reader) {
	m.attrs = append(m.attrs, rtnl.Attr{t, v})
}

func (m *mod) getifindices() error {
	m.ifindexByName = make(map[string]int32)
	m.vrfByName = make(map[string]uint32)

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
	return m.sr.UntilDone(req, func(b []byte) {
		var ifla rtnl.Ifla
		if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWLINK {
			return
		}
		msg := rtnl.IfInfoMsgPtr(b)
		ifla.Write(b)
		name := rtnl.Kstring(ifla[rtnl.IFLA_IFNAME])
		m.ifindexByName[name] = msg.Index
		if rtnl.Kstring(ifla[rtnl.IFLA_INFO_KIND]) == "vrf" {
			m.vrfByName[name] =
				rtnl.Uint32(ifla[rtnl.IFLA_VRF_TABLE])
		}
	})
}

func (m *mod) parse() error {
	var (
		err     error
		mxlock  rtnl.Uint32Attr
		mxattrs rtnl.Attrs
	)
	mxappend := func(t uint16, v io.Reader) {
		mxattrs = append(mxattrs, rtnl.Attr{t, v})
	}
	if s := m.opt.Parms.ByName["-f"]; len(s) > 0 {
		if v, ok := rtnl.AfByName[s]; ok {
			m.msg.Family = v
		} else {
			return fmt.Errorf("family: %q unknown", s)
		}
	}
	if len(m.args) > 0 {
		if v, ok := rtnl.RtnByName[m.args[0]]; ok {
			m.msg.Type = v
			m.args = m.args[1:]
		}
	}

	prefix, err := m.parsePrefix(m.msg.Family)
	if err != nil {
		return err
	}
	m.msg.Family = prefix.Family()
	m.msg.Dst_len = prefix.Len()
	m.append(rtnl.RTA_DST, prefix)
	m.msg.Scope = rtnl.RT_SCOPE_UNIVERSE

	for err == nil && len(m.args) > 0 {
		arg0 := m.args[0]
		m.args = m.args[1:]
		switch arg0 {
		case "src":
			if v, e := m.parseAddress(m.msg.Family); e == nil {
				m.append(rtnl.RTA_PREFSRC, v)
			} else {
				err = e
			}
		case "as":
			if v, e := m.parseAs(); err == nil {
				m.append(rtnl.RTA_NEWDST, v)
			} else {
				err = e
			}
		case "via":
			if v, e := m.parseVia(); e == nil {
				if m.msg.Family == v.Family() {
					m.append(rtnl.RTA_GATEWAY, v)
				} else {
					m.append(rtnl.RTA_VIA,
						rtnl.RtVia{uint16(v.Family()),
							v.Bytes()})
				}
			} else {
				err = e
			}
		case "from":
			if v, e := m.parsePrefix(m.msg.Family); e == nil {
				if m.msg.Family == rtnl.AF_UNSPEC {
					m.msg.Family = v.Family()
				}
				m.append(rtnl.RTA_SRC, v)
			} else {
				err = e
			}
		case "tos", "dstfield":
			err = m.parseTos()
		case "scope":
			err = m.parseScope()
		case "expires", "metric", "priority":
			t := map[string]uint16{
				"expires":  rtnl.RTA_EXPIRES,
				"metric":   rtnl.RTA_PRIORITY,
				"priority": rtnl.RTA_PRIORITY,
			}[arg0]
			if v, e := m.parseNumber(); e == nil {
				m.append(t, rtnl.Uint32Attr(v))
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
			if len(m.args) > 0 && m.args[0] == "lock" {
				mxlock |= 1 << t
				m.args = m.args[1:]
			}
			if v, e := m.parseNumber(); e == nil {
				mxappend(t, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "congctl": // [ lock ] STRING
			if len(m.args) > 0 && m.args[0] == "lock" {
				mxlock |= 1 << rtnl.RTAX_CC_ALGO
				m.args = m.args[1:]
			}
			if v, e := m.parseString(); e == nil {
				mxappend(rtnl.RTAX_CC_ALGO,
					rtnl.KstringAttr(v))
			} else {
				err = e
			}
		case "rtt":
			if len(m.args) > 0 && m.args[0] == "lock" {
				mxlock |= 1 << rtnl.RTAX_RTT
				m.args = m.args[1:]
			}
			if v, e := m.parseRtt(8); e == nil {
				mxappend(rtnl.RTAX_RTT, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "rto-min", "rto_min":
			if v, e := m.parseRtt(1); e == nil {
				mxappend(rtnl.RTAX_RTO_MIN, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "rttvar":
			if len(m.args) > 0 && m.args[0] == "lock" {
				mxlock |= 1 << rtnl.RTAX_RTTVAR
				m.args = m.args[1:]
			}
			if v, e := m.parseRtt(4); e == nil {
				mxappend(rtnl.RTAX_RTTVAR, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "quickack":
			var v rtnl.Uint32Attr
			if len(m.args) > 0 {
				if m.args[0] == "1" ||
					m.args[0] == "t" ||
					m.args[0] == "true" {
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
			if len(m.args) > 0 {
				switch m.args[0] {
				case "ecn:":
					features |= rtnl.RTAX_FEATURE_ECN
				default:
					err = fmt.Errorf("feature: %q unknown",
						m.args[0])
				}
				m.args = m.args[1:]
			}
			mxappend(rtnl.RTAX_FEATURES, rtnl.Uint32Attr(features))
		case "realms":
			if v, e := m.parseRealm(); e == nil {
				m.append(rtnl.RTA_FLOW, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "onlink":
			m.msg.Flags |= uint32(rtnl.RTNH_F_ONLINK)
		case "nexthop":
			if nhs, e := m.parseNextHops(); e == nil {
				m.append(rtnl.RTA_MULTIPATH, nhs)
			} else {
				err = e
			}
		case "prot", "protocol":
			err = m.parseProtocol()
		case "table":
			err = m.parseTable()
		case "vrf":
			err = m.parseVrf()
		case "dev", "oif":
			if v, e := m.parseIfname(); e == nil {
				m.append(rtnl.RTA_OIF, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "pref", "preference":
			if v, e := m.parsePreference(); e == nil {
				m.append(rtnl.RTA_PREF, rtnl.Uint8Attr(v))
			} else {
				err = e
			}
		case "encap":
			if v, e := m.parseEncap(); e == nil {
				m.append(rtnl.RTA_ENCAP, v)
			} else {
				err = e
			}
		case "ttl-propagate", "+ttl-propagate":
			m.append(rtnl.RTA_TTL_PROPAGATE, rtnl.Uint8Attr(1))
		case "no-ttl-propagate", "-ttl-propagate":
			m.append(rtnl.RTA_TTL_PROPAGATE, rtnl.Uint8Attr(0))
		default:
			err = fmt.Errorf("unexpected")
		}
		if err != nil {
			err = fmt.Errorf("%s: %s", arg0, err)
		}
	}
	if mxlock != 0 {
		m.append(rtnl.RTAX_LOCK, rtnl.Uint32Attr(mxlock))
	}
	if len(mxattrs) > 0 {
		m.append(rtnl.RTA_METRICS, mxattrs)
	}
	return err
}

func (m *mod) parseString() (string, error) {
	var v string
	if len(m.args) == 0 {
		return v, fmt.Errorf("missing STRING")
	}
	v = m.args[0]
	m.args = m.args[1:]
	return v, nil
}

func (m *mod) parseNumber() (int64, error) {
	var v int64
	if len(m.args) == 0 {
		return v, fmt.Errorf("missing NUMBER")
	}
	if _, err := fmt.Sscan(m.args[0], &v); err != nil {
		return v, fmt.Errorf("%q %v", m.args[0], err)
	}
	m.args = m.args[1:]
	return v, nil
}

func (m *mod) parseRtt(rawfactor int64) (int64, error) {
	var v int64
	if len(m.args) == 0 {
		return v, fmt.Errorf("missing NUMBER")
	}
	arg0 := m.args[0]
	m.args = m.args[1:]
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
	if _, err := fmt.Sscan(m.args[0], &v); err != nil {
		return v, fmt.Errorf("%q %v", m.args[0], err)
	}
	return v * factor, nil
}

func (m *mod) parsePrefix(family uint8) (rtnl.Prefixer, error) {
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing PREFIX")
	}
	prefix, err := rtnl.Prefix(m.args[0], family)
	m.args = m.args[1:]
	return prefix, err
}

func (m *mod) parseAddress(family uint8) (rtnl.Addresser, error) {
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing ADDRESS")
	}
	addr, err := rtnl.Address(m.args[0], family)
	m.args = m.args[1:]
	return addr, err
}

func (m *mod) parseTos() error {
	if len(m.args) == 0 {
		return fmt.Errorf("missing TOS")
	}
	if _, err := fmt.Sscan(m.args[0], &m.msg.Tos); err != nil {
		return fmt.Errorf("%q %v", m.args[0], err)
	}
	m.args = m.args[1:]
	return nil
}

func (m *mod) parseTable() error {
	var t rtnl.Uint32Attr
	if len(m.args) == 0 {
		return fmt.Errorf("missing RTTABLE")
	}
	if v, ok := rtnl.RtTableByName[m.args[0]]; ok {
		m.msg.Table = uint8(v)
	} else if _, err := fmt.Sscan(m.args[0], &t); err != nil {
		return fmt.Errorf("%q %v", m.args[0], err)
	} else if t < 256 {
		m.msg.Table = uint8(t)
	} else {
		m.msg.Table = uint8(rtnl.RT_TABLE_UNSPEC)
		m.append(rtnl.RTA_TABLE, t)
	}
	m.args = m.args[1:]
	return nil
}

func (m *mod) parseVrf() error {
	if len(m.args) == 0 {
		return fmt.Errorf("missing VRF")
	}
	if vrf, found := m.vrfByName[m.args[0]]; !found {
		return fmt.Errorf("%q no found", m.args[0])
	} else if vrf < 256 {
		m.msg.Table = uint8(vrf)
	} else {
		m.msg.Table = uint8(rtnl.RT_TABLE_UNSPEC)
		m.append(rtnl.RTA_TABLE, rtnl.Uint32Attr(vrf))
	}
	m.args = m.args[1:]
	return nil
}

func (m *mod) parsePreference() (uint8, error) {
	if len(m.args) == 0 {
		return 0, fmt.Errorf("missing { low | medium | high }")
	}
	pref, ok := map[string]uint8{
		"low":    rtnl.ICMPV6_ROUTER_PREF_LOW,
		"med":    rtnl.ICMPV6_ROUTER_PREF_MEDIUM,
		"medium": rtnl.ICMPV6_ROUTER_PREF_MEDIUM,
		"hi":     rtnl.ICMPV6_ROUTER_PREF_HIGH,
		"high":   rtnl.ICMPV6_ROUTER_PREF_HIGH,
	}[m.args[0]]
	if !ok {
		return 0, fmt.Errorf("%q invalid", m.args[0])
	}
	m.args = m.args[1:]
	return pref, nil
}

func (m *mod) parseProtocol() error {
	if len(m.args) == 0 {
		return fmt.Errorf("missing RTPROTOCOL")
	}
	if v, ok := rtnl.RtProtByName[m.args[0]]; ok {
		m.msg.Protocol = v
	} else if _, err := fmt.Sscan(m.args[0], &m.msg.Protocol); err != nil {
		return fmt.Errorf("%q %v", m.args[0], err)
	}
	m.args = m.args[1:]
	return nil
}

func (m *mod) parseScope() error {
	if len(m.args) == 0 {
		return fmt.Errorf("missing RTSCOPE")
	}
	if v, ok := rtnl.RtScopeByName[m.args[0]]; ok {
		m.msg.Scope = v
	} else {
		return fmt.Errorf("%q unknown", m.args[0])
	}
	m.args = m.args[1:]
	return nil
}

func (m *mod) parseEncap() (io.Reader, error) {
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing ENCAP")
	}
	arg0 := m.args[0]
	m.args = m.args[1:]
	switch arg0 {
	case "mpls":
		return m.parseEncapMpls()
	case "ip":
		return m.parseEncapIp()
	case "ip6":
		return m.parseEncapIp6()
	case "ila":
		return m.parseEncapIla()
	case "bpf":
		return m.parseEncapBpf()
	case "seg6":
		return m.parseEncapSeg6()
	}
	return nil, fmt.Errorf("%q unknown", arg0)
}

func completeEncap(s string) (list []string) {
	for _, encap := range []string{
		"mpls",
		"ip",
		"ip6",
		"ila",
		"bpf",
		"seg6",
	} {
		if len(s) == 0 || strings.HasPrefix(encap, s) {
			list = append(list, encap)
		}
	}
	return
}

func (m *mod) parseVia() (rtnl.Addresser, error) {
	family := m.msg.Family
	mia := fmt.Errorf("missing ADDRESS")
	if len(m.args) == 0 {
		return nil, mia
	}
	if viaFamily, ok := rtnl.AfByName[m.args[0]]; ok {
		family = viaFamily
		m.args = m.args[1:]
	}
	if len(m.args) == 0 {
		return nil, mia
	}
	addr, err := rtnl.Address(m.args[0], family)
	if err != nil {
		return nil, err
	}
	m.args = m.args[1:]
	if m.msg.Family == rtnl.AF_UNSPEC {
		m.msg.Family = addr.Family()
	}
	return addr, nil
}

func (m *mod) parseIfname() (int, error) {
	if len(m.args) == 0 {
		return -1, fmt.Errorf("missing IFNAME")
	}
	ifindex, found := m.ifindexByName[m.args[0]]
	if !found {
		return -1, fmt.Errorf("%q not found", m.args[0])
	}
	m.args = m.args[1:]
	return int(ifindex), nil
}

func (m *mod) parseWeight() (uint8, error) {
	var u8 uint8
	if len(m.args) == 0 {
		return u8, fmt.Errorf("missing WEIGHT")
	}
	if _, err := fmt.Sscan(m.args[0], &u8); err != nil {
		return 0, err
	} else if u8 < 1 {
		return 0, fmt.Errorf("must be >= 1")
	}
	m.args = m.args[1:]
	return u8 - 1, nil
}

func (m *mod) parseRealm() (uint32, error) {
	if len(m.args) == 0 {
		return 0, fmt.Errorf("missing REALM")
	}
	arg0 := m.args[0]
	m.args = m.args[1:]
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
func (m *mod) parseAs() (rtnl.Addresser, error) {
	if len(m.args) >= 1 && m.args[0] == "to" {
		m.args = m.args[1:]
	}
	return m.parseAddress(m.msg.Family)
}

// NH [ nexthop NH... ]
// NH := [ encap ENCAP ] [ via [ FAMILY ] ADDRESS ] [ dev IFNAME ]
//	[ weight WEIGHT ] [ onlink | pervasive ]
func (m *mod) parseNextHops() (rtnl.RtnhAttrsList, error) {
	var (
		err error
		nh  rtnl.RtnhAttrs
	)
	nhappend := func(t uint16, v io.Reader) {
		nh.Attrs = append(nh.Attrs, rtnl.Attr{t, v})
	}
	nhs := rtnl.RtnhAttrsList{nh}
nhloop:
	for err == nil && len(m.args) > 0 {
		arg0 := m.args[0]
		m.args = m.args[1:]
		switch arg0 {
		case "nexthop":
			// recurse
			if more, e := m.parseNextHops(); e == nil {
				nhs = append(nhs, more...)
				break nhloop
			} else {
				err = e
			}
		case "via":
			if v, e := m.parseVia(); e == nil {
				if m.msg.Family == v.Family() {
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
			if len(m.args) == 0 {
				err = fmt.Errorf("missing IFNAME")
			} else if i, ok := m.ifindexByName[m.args[0]]; !ok {
				err = fmt.Errorf("%q not found", m.args[0])
			} else {
				nh.Ifindex = int(i)
				m.args = m.args[1:]
			}
		case "weight":
			if v, e := m.parseWeight(); e == nil {
				nh.Rtnh.Hops = v
			} else {
				err = e
			}
		case "onlink":
			nh.Rtnh.Flags |= rtnl.RTNH_F_ONLINK
		case "realm":
			if v, e := m.parseRealm(); e == nil {
				nhappend(rtnl.RTA_FLOW, rtnl.Uint32Attr(v))
			} else {
				err = e
			}
		case "as":
			if v, e := m.parseAs(); e == nil {
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
func (m *mod) parseEncapMpls() (io.Reader, error) {
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing LABEL")
	}
	addr, err := rtnl.Address(m.args[0], rtnl.AF_MPLS)
	if err != nil {
		return nil, err
	}
	attrs := rtnl.Attrs{rtnl.Attr{rtnl.MPLS_IPTUNNEL_DST, addr}}
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return attrs, nil
	}
	if m.args[0] != "ttl" {
		return nil, fmt.Errorf("%s: unexpected", m.args[0])
	}
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing TTL")
	}
	var ttl uint8
	if _, err = fmt.Sscan(m.args[0], &ttl); err == nil {
		return nil, fmt.Errorf("ttl: %v", err)
	}
	attrs = append(attrs,
		rtnl.Attr{rtnl.MPLS_IPTUNNEL_TTL, rtnl.Uint8Attr(ttl)})
	m.args = m.args[1:]
	return attrs, nil
}

// id TUNNEL-ID dst REMOTE-IP [ tos TOS ] [ ttl TTL ]
func (m *mod) parseEncapIp() (io.Reader, error) {
	var attrs rtnl.Attrs
	appendAttr := func(t uint16, v io.Reader) {
		attrs = append(attrs, rtnl.Attr{t, v})
	}
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing id")
	}
	if m.args[0] != "id" {
		return nil, fmt.Errorf("%s: unexpected", m.args[0])
	}
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing TUNNEL-ID")
	}
	var id uint64
	if _, err := fmt.Sscan(m.args[0], &id); err == nil {
		return nil, fmt.Errorf("id: %v", err)
	}
	appendAttr(rtnl.LWTUNNEL_IP_ID, rtnl.Uint64Attr(id))
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing dst")
	}
	if m.args[0] != "dst" {
		return nil, fmt.Errorf("%s: unexpected", m.args[0])
	}
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing REMOTE-IP")
	}
	if addr, err := rtnl.Address(m.args[0], rtnl.AF_INET); err != nil {
		return nil, fmt.Errorf("dst: %v", err)
	} else {
		appendAttr(rtnl.LWTUNNEL_IP_DST, addr)
	}
	m.args = m.args[1:]
	for len(m.args) > 0 {
		switch m.args[0] {
		case "tos":
			var tos uint32
			m.args = m.args[1:]
			if len(m.args) == 0 {
				return nil, fmt.Errorf("missing TOS")
			}
			// FIXME symbolic TOS
			if _, err := fmt.Sscan(m.args[0], &tos); err != nil {
				return nil, fmt.Errorf("tos: %v", err)
			}
			appendAttr(rtnl.LWTUNNEL_IP_TOS, rtnl.Uint32Attr(tos))
			m.args = m.args[1:]
		case "ttl":
			var ttl uint8
			m.args = m.args[1:]
			if len(m.args) == 0 {
				return nil, fmt.Errorf("missing TTL")
			}
			if _, err := fmt.Sscan(m.args[0], &ttl); err != nil {
				return nil, fmt.Errorf("ttl: %v", err)
			}
			appendAttr(rtnl.LWTUNNEL_IP_TTL, rtnl.Uint8Attr(ttl))
			m.args = m.args[1:]
		default:
			return attrs, nil
		}
	}
	return attrs, nil
}

// id TUNNEL-ID dst REMOTE-IP [ tc TC ] [ hoplimit HOPS ]
func (m *mod) parseEncapIp6() (io.Reader, error) {
	var attrs rtnl.Attrs
	appendAttr := func(t uint16, v io.Reader) {
		attrs = append(attrs, rtnl.Attr{t, v})
	}
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing id")
	}
	if m.args[0] != "id" {
		return nil, fmt.Errorf("%s: unexpected", m.args[0])
	}
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing TUNNEL-ID")
	}
	var id uint64
	if _, err := fmt.Sscan(m.args[0], &id); err == nil {
		return nil, fmt.Errorf("id: %v", err)
	}
	appendAttr(rtnl.LWTUNNEL_IP6_ID, rtnl.Uint64Attr(id))
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing dst")
	}
	if m.args[0] != "dst" {
		return nil, fmt.Errorf("%s: unexpected", m.args[0])
	}
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing REMOTE-IP6")
	}
	if addr, err := rtnl.Address(m.args[0], rtnl.AF_INET6); err != nil {
		return nil, fmt.Errorf("dst: %v", err)
	} else {
		appendAttr(rtnl.LWTUNNEL_IP6_DST, addr)
	}
	m.args = m.args[1:]
	for len(m.args) > 0 {
		switch m.args[0] {
		case "tc":
			var tc uint8
			m.args = m.args[1:]
			if len(m.args) == 0 {
				return nil, fmt.Errorf("missing TC")
			}
			// FIXME symbolic TOS
			if _, err := fmt.Sscan(m.args[0], &tc); err != nil {
				return nil, fmt.Errorf("tc: %v", err)
			}
			appendAttr(rtnl.LWTUNNEL_IP6_TC, rtnl.Uint8Attr(tc))
			m.args = m.args[1:]
		case "ttl":
			var hops uint8
			m.args = m.args[1:]
			if len(m.args) == 0 {
				return nil, fmt.Errorf("missing HOPS")
			}
			if _, err := fmt.Sscan(m.args[0], &hops); err != nil {
				return nil, fmt.Errorf("tops: %v", err)
			}
			appendAttr(rtnl.LWTUNNEL_IP6_HOPLIMIT,
				rtnl.Uint8Attr(hops))
			m.args = m.args[1:]
		default:
			return attrs, nil
		}
	}
	return attrs, nil
}

// LOCATOR [ csum-mode { adj-transport | neutral-map | no-action } ]
func (m *mod) parseEncapIla() (io.Reader, error) {
	var attrs rtnl.Attrs
	appendAttr := func(t uint16, v io.Reader) {
		attrs = append(attrs, rtnl.Attr{t, v})
	}
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing LOCATOR")
	}
	var locator uint64
	if _, err := fmt.Sscan(m.args[0], &locator); err == nil {
		return nil, fmt.Errorf("locator: %v", err)
	}
	appendAttr(rtnl.ILA_ATTR_LOCATOR, rtnl.Uint64Attr(locator))
	m.args = m.args[1:]
	if len(m.args) > 0 && m.args[0] == "csum-mode" {
		m.args = m.args[1:]
		if len(m.args) == 0 {
			return nil, fmt.Errorf("missing CSUM-MODE")
		}
		mode, found := rtnl.IlaCsumModeByName[m.args[0]]
		if !found {
			return nil, fmt.Errorf("csum-mode: %q invalid",
				m.args[0])
		}
		appendAttr(rtnl.ILA_ATTR_CSUM_MODE, rtnl.Uint8Attr(mode))
		m.args = m.args[1:]
	}
	return attrs, nil
}

// mode { encap | inline } segs SEGMENTS [ hmac KEYID ]
func (m *mod) parseEncapSeg6() (io.Reader, error) {
	var segs []net.IP
	var hmac uint32
	var flags uint8
	mode := rtnl.SEG6_IPTUN_MODE_UNSPEC
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing mode")
	}
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing MODE")
	}
	switch m.args[0] {
	case "encap":
		mode = rtnl.SEG6_IPTUN_MODE_ENCAP
	case "inline":
		mode = rtnl.SEG6_IPTUN_MODE_INLINE
	default:
		return nil, fmt.Errorf("mode: %q invalid", m.args[0])
	}
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing segs")
	}
	m.args = m.args[1:]
	if len(m.args) == 0 {
		return nil, fmt.Errorf("missing SEGMENTS")
	}
	for _, s := range strings.Split(m.args[0], ",") {
		seg := net.ParseIP(s)
		if seg.To16() == nil {
			return nil, fmt.Errorf("segment: %q invalid", s)
		}
		segs = append(segs, seg)
	}
	srhlen := 8 + (16 * len(segs))
	m.args = m.args[1:]
	if len(m.args) > 0 && m.args[0] == "hmac" {
		m.args = m.args[1:]
		if len(m.args) == 0 {
			return nil, fmt.Errorf("missing KEYID")
		}
		if _, err := fmt.Sscan(m.args[0], &hmac); err != nil {
			return nil, fmt.Errorf("hmac: %q %v", m.args[0], err)
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
func (m *mod) parseEncapBpf() (io.Reader, error) {
	return nil, fixme
}
