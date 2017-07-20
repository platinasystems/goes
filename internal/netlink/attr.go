// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"unsafe"

	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/internal/accumulate"
	"github.com/platinasystems/go/internal/indent"
)

type AttrType interface {
	attrType()
	IthString(int) string
}

type Attr interface {
	attr()
	Set([]byte)
	Size() int
	fmt.Stringer
	io.WriterTo
}

type Int8er interface {
	Int() int8
}
type Int32er interface {
	Uint() int32
}
type Int64er interface {
	Int() int64
}
type Uint8er interface {
	Uint() uint8
}
type Uint32er interface {
	Uint() uint32
}
type Uint64er interface {
	Uint() uint64
}

func closeAttrs(attrs []Attr) {
	for i, a := range attrs {
		if a != nil {
			if method, found := a.(io.Closer); found {
				method.Close()
			}
			attrs[i] = nil
		}
	}
}

func fprintAttrs(w io.Writer, names []string, attrs []Attr) (int64,
	error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for i, v := range attrs {
		if v == nil {
			continue
		}
		fmt.Fprint(acc, elib.Stringer(names, i), ":")
		if _, found := v.(multiliner); found {
			fmt.Fprintln(acc)
			indent.Increase(acc)
			v.WriteTo(acc)
			indent.Decrease(acc)
		} else {
			fmt.Fprint(acc, " ")
			v.WriteTo(acc)
			fmt.Fprintln(acc)
		}
	}
	return acc.Tuple()
}

func nextAttr(b []byte, i int) (n *NlAttr, v []byte, j int) {
	n = (*NlAttr)(unsafe.Pointer(&b[i]))
	v = b[i+SizeofNlAttr : i+int(n.Len)]
	j = i + attrAlignLen(int(n.Len))
	return
}

func parse_af_spec(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewAddressFamilyAttrType()
	for i := 0; i < len(b); {
		a, v, next := nextAttr(b, i)
		i = next
		af := AddressFamily(a.Kind())
		as.X.Validate(uint(af))
		switch af {
		case AF_INET:
			as.X[af] = parse_ip4_af_spec(v)
		case AF_INET6:
			as.X[af] = parse_ip6_af_spec(v)
		default:
			// For now don't panic.
			// With bridgemodeset() an AF_UNSPEC msg comes down
			// which we ignore.
			if false {
				panic("unknown address family " + af.String())
			}
		}
	}
	return as
}

func parse_ip4_af_spec(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewIp4IfAttrType()
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		t := Ip4IfAttrKind(n.Kind())
		as.X.Validate(uint(t))
		switch t {
		case IFLA_INET_UNSPEC:
		case IFLA_INET_CONF:
			as.X[t] = NewIp4DevConfBytes(v)
		default:
			as.X[t] = NewHexStringAttrBytes(v)
		}
	}
	return as
}

func parse_ip6_af_spec(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewIp6IfAttrType()
	for i := 0; i < len(b); {
		n, v, next_i := nextAttr(b, i)
		i = next_i
		t := Ip6IfAttrKind(n.Kind())
		as.X.Validate(uint(t))
		switch t {
		case IFLA_INET6_UNSPEC:
		case IFLA_INET6_FLAGS:
			flags := Ip6IfFlagsAttrBytes(v)
			if flags != 0 {
				as.X[t] = flags
			}
		case IFLA_INET6_CONF:
			as.X[t] = NewIp6DevConfBytes(v)
		case IFLA_INET6_STATS:
		case IFLA_INET6_MCAST:
		case IFLA_INET6_CACHEINFO:
		case IFLA_INET6_ICMP6STATS:
		case IFLA_INET6_TOKEN:
		case IFLA_INET6_ADDR_GEN_MODE:
		default:
			as.X[t] = NewHexStringAttrBytes(v)
		}
	}
	return as
}

const (
	IFLA_INFO_UNSPEC IfLinkInfoAttrKind = iota
	IFLA_INFO_KIND
	IFLA_INFO_DATA
	IFLA_INFO_XSTATS
	IFLA_INFO_SLAVE_KIND
	IFLA_INFO_SLAVE_DATA
)

var ifLinkInfoAttrTypeNames = []string{
	IFLA_INFO_UNSPEC:     "UNSPEC",
	IFLA_INFO_KIND:       "KIND",
	IFLA_INFO_DATA:       "DATA",
	IFLA_INFO_XSTATS:     "XSTATS",
	IFLA_INFO_SLAVE_KIND: "SLAVE_KIND",
	IFLA_INFO_SLAVE_DATA: "SLAVE_DATA",
}

func (t IfLinkInfoAttrKind) String() string { return elib.Stringer(ifLinkInfoAttrTypeNames, int(t)) }

type IfLinkInfoAttrKind int
type IfLinkInfoAttrType Empty

func NewIfLinkInfoAttrType() *IfLinkInfoAttrType {
	return (*IfLinkInfoAttrType)(pool.Empty.Get().(*Empty))
}

func (t *IfLinkInfoAttrType) attrType() {}
func (t *IfLinkInfoAttrType) Close() error {
	repool(t)
	return nil
}
func (t *IfLinkInfoAttrType) IthString(i int) string {
	return elib.Stringer(ifLinkInfoAttrTypeNames, i)
}

type InterfaceKind int

const (
	InterfaceKindUnknown InterfaceKind = iota
	InterfaceKindDummy
	InterfaceKindTun
	InterfaceKindVeth
	InterfaceKindVlan
	InterfaceKindIpip
	InterfaceKindIp6Tunnel
	InterfaceKindIp4GRE
	InterfaceKindIp4GRETap
	InterfaceKindIp6GRE
	InterfaceKindIp6GRETap
)

var kindStrings = [...]string{
	InterfaceKindUnknown:   "",
	InterfaceKindDummy:     "dummy",
	InterfaceKindTun:       "tun",
	InterfaceKindVeth:      "veth",
	InterfaceKindVlan:      "vlan",
	InterfaceKindIpip:      "ipip",
	InterfaceKindIp6Tunnel: "ip6tnl",
	InterfaceKindIp4GRE:    "gre",
	InterfaceKindIp4GRETap: "gretap",
	InterfaceKindIp6GRE:    "ip6gre",
	InterfaceKindIp6GRETap: "ip6gretap",
}

func (k InterfaceKind) String() string { return kindStrings[k] }

var kindMap = map[string]InterfaceKind{
	"dummy":     InterfaceKindDummy,
	"tun":       InterfaceKindTun,
	"veth":      InterfaceKindVeth,
	"vlan":      InterfaceKindVlan,
	"ipip":      InterfaceKindIpip,
	"ip6tnl":    InterfaceKindIp6Tunnel,
	"gre":       InterfaceKindIp4GRE,
	"gretap":    InterfaceKindIp4GRETap,
	"ip6gre":    InterfaceKindIp6GRE,
	"ip6gretap": InterfaceKindIp6GRETap,
}

func (m *IfInfoMessage) InterfaceKind() (k InterfaceKind) {
	if a, ok := m.Attrs[IFLA_LINKINFO].(*AttrArray); ok {
		k = kindMap[a.X[IFLA_INFO_KIND].String()]
	}
	return
}

func parse_link_info(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewIfLinkInfoAttrType()
	linkKind := InterfaceKindUnknown
	for i := 0; i < len(b); {
		a, v, next := nextAttr(b, i)
		i = next
		kind := IfLinkInfoAttrKind(a.Kind())
		as.X.Validate(uint(kind))
		switch kind {
		case IFLA_INFO_KIND, IFLA_INFO_SLAVE_KIND:
			// Remove trailing 0.
			l := len(v)
			for l > 0 && v[l-1] == 0 {
				l = l - 1
			}
			as.X[kind] = StringAttrBytes(v[:l])
			linkKind = kindMap[string(v[:l])]
		case IFLA_INFO_DATA, IFLA_INFO_SLAVE_DATA:
			as.X[kind] = StringAttrBytes(v)
		default:
			panic("unknown link info attribute kind " + kind.String())
		}
	}
	switch linkKind {
	case InterfaceKindVlan:
		as.X[IFLA_INFO_DATA] = parse_vlan_info([]byte(as.X[IFLA_INFO_DATA].(StringAttr)))
	case InterfaceKindIpip, InterfaceKindIp6Tunnel:
		as.X[IFLA_INFO_DATA] = parse_iptun_info([]byte(as.X[IFLA_INFO_DATA].(StringAttr)), linkKind)
	case InterfaceKindIp4GRE, InterfaceKindIp4GRETap, InterfaceKindIp6GRE, InterfaceKindIp6GRETap:
		as.X[IFLA_INFO_DATA] = parse_gre_info([]byte(as.X[IFLA_INFO_DATA].(StringAttr)), linkKind)
	}
	return as
}

const (
	IFLA_VLAN_UNSPEC IfVlanLinkInfoDataAttrKind = iota
	IFLA_VLAN_ID
	IFLA_VLAN_FLAGS
	IFLA_VLAN_EGRESS_QOS
	IFLA_VLAN_INGRESS_QOS
	IFLA_VLAN_PROTOCOL
)

var ifVlanLinkInfoDataAttrKindNames = []string{
	IFLA_VLAN_UNSPEC:      "UNSPEC",
	IFLA_VLAN_ID:          "VLAN_ID",
	IFLA_VLAN_FLAGS:       "VLAN_FLAGS",
	IFLA_VLAN_EGRESS_QOS:  "VLAN_EGRESS_QOS",
	IFLA_VLAN_INGRESS_QOS: "VLAN_INGRESS_QOS",
	IFLA_VLAN_PROTOCOL:    "VLAN_PROTOCOL",
}

func (t IfVlanLinkInfoDataAttrKind) String() string {
	return elib.Stringer(ifVlanLinkInfoDataAttrKindNames, int(t))
}

type IfVlanLinkInfoDataAttrKind int
type IfVlanLinkInfoDataAttrType Empty

func NewIfVlanLinkInfoDataAttrType() *IfVlanLinkInfoDataAttrType {
	return (*IfVlanLinkInfoDataAttrType)(pool.Empty.Get().(*Empty))
}

func (t *IfVlanLinkInfoDataAttrType) attrType() {}
func (t *IfVlanLinkInfoDataAttrType) Close() error {
	repool(t)
	return nil
}
func (t *IfVlanLinkInfoDataAttrType) IthString(i int) string {
	return elib.Stringer(ifVlanLinkInfoDataAttrKindNames, i)
}

func parse_vlan_info(b []byte) (as *AttrArray) {
	as = pool.AttrArray.Get().(*AttrArray)
	as.Type = NewIfVlanLinkInfoDataAttrType()
	for i := 0; i < len(b); {
		a, v, next := nextAttr(b, i)
		i = next
		kind := IfVlanLinkInfoDataAttrKind(a.Kind())
		as.X.Validate(uint(kind))
		switch kind {
		case IFLA_VLAN_ID:
			as.X[kind] = Uint16AttrBytes(v)
		case IFLA_VLAN_PROTOCOL:
			as.X[kind] = VlanProtocolAttrBytes(v)
		case IFLA_VLAN_FLAGS:
			as.X[kind] = NewVlanFlagsBytes(v)
		default:
			panic("unknown vlan link data attribute kind " + kind.String())
		}
	}
	return as
}

type VlanProtocolAttr uint16

func VlanProtocolAttrBytes(b []byte) VlanProtocolAttr {
	return VlanProtocolAttr(uint16(b[0])<<8 | uint16(b[1]))
}

func (a VlanProtocolAttr) attr() {}
func (a VlanProtocolAttr) Set(v []byte) {
	v[0] = byte(a >> 8)
	v[1] = byte(a)
}
func (a VlanProtocolAttr) Size() int {
	return 2
}
func (a VlanProtocolAttr) String() string {
	return StringOf(a)
}
func (a VlanProtocolAttr) Uint() uint16 {
	return uint16(a)
}
func (a VlanProtocolAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintf(acc, "0x%04x", a.Uint())
	return acc.Tuple()
}

type VlanFlags struct {
	Flags uint32
	Mask  uint32
}

func NewVlanFlagsBytes(b []byte) *VlanFlags {
	a := pool.VlanFlags.Get().(*VlanFlags)
	a.Parse(b)
	return a
}

func (a *VlanFlags) attr() {}

func (a *VlanFlags) Close() error {
	repool(a)
	return nil
}
func (a *VlanFlags) Set(v []byte) {
	panic("should never be called")
}
func (a *VlanFlags) Size() int {
	panic("should never be called")
	return 0
}
func (a *VlanFlags) String() string {
	return StringOf(a)
}
func (a *VlanFlags) Parse(b []byte) {
	*a = *(*VlanFlags)(unsafe.Pointer(&b[0]))
}
func (a *VlanFlags) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintf(acc, "flags: %x mask: %x", a.Flags, a.Mask)
	return acc.Tuple()
}

const (
	IFLA_IPTUN_UNSPEC IfIptunLinkInfoDataAttrKind = iota
	IFLA_IPTUN_LINK
	IFLA_IPTUN_LOCAL
	IFLA_IPTUN_REMOTE
	IFLA_IPTUN_TTL
	IFLA_IPTUN_TOS
	IFLA_IPTUN_ENCAP_LIMIT
	IFLA_IPTUN_FLOWINFO
	IFLA_IPTUN_FLAGS
	IFLA_IPTUN_PROTO
	IFLA_IPTUN_PMTUDISC
	IFLA_IPTUN_6RD_PREFIX
	IFLA_IPTUN_6RD_RELAY_PREFIX
	IFLA_IPTUN_6RD_PREFIXLEN
	IFLA_IPTUN_6RD_RELAY_PREFIXLEN
	IFLA_IPTUN_ENCAP_TYPE
	IFLA_IPTUN_ENCAP_FLAGS
	IFLA_IPTUN_ENCAP_SPORT
	IFLA_IPTUN_ENCAP_DPORT
	IFLA_IPTUN_COLLECT_METADATA
	IFLA_IPTUN_FWMARK
)

var ifIptunLinkInfoDataAttrKindNames = []string{
	IFLA_IPTUN_UNSPEC:              "UNSPEC",
	IFLA_IPTUN_LINK:                "IPTUN_LINK",
	IFLA_IPTUN_LOCAL:               "IPTUN_LOCAL",
	IFLA_IPTUN_REMOTE:              "IPTUN_REMOTE",
	IFLA_IPTUN_TTL:                 "IPTUN_TTL",
	IFLA_IPTUN_TOS:                 "IPTUN_TOS",
	IFLA_IPTUN_ENCAP_LIMIT:         "IPTUN_ENCAP_LIMIT",
	IFLA_IPTUN_FLOWINFO:            "IPTUN_FLOWINFO",
	IFLA_IPTUN_FLAGS:               "IPTUN_FLAGS",
	IFLA_IPTUN_PROTO:               "IPTUN_PROTO",
	IFLA_IPTUN_PMTUDISC:            "IPTUN_PMTUDISC",
	IFLA_IPTUN_6RD_PREFIX:          "IPTUN_6RD_PREFIX",
	IFLA_IPTUN_6RD_RELAY_PREFIX:    "IPTUN_6RD_RELAY_PREFIX",
	IFLA_IPTUN_6RD_PREFIXLEN:       "IPTUN_6RD_PREFIXLEN",
	IFLA_IPTUN_6RD_RELAY_PREFIXLEN: "IPTUN_6RD_RELAY_PREFIXLEN",
	IFLA_IPTUN_ENCAP_TYPE:          "IPTUN_ENCAP_TYPE",
	IFLA_IPTUN_ENCAP_FLAGS:         "IPTUN_ENCAP_FLAGS",
	IFLA_IPTUN_ENCAP_SPORT:         "IPTUN_ENCAP_SPORT",
	IFLA_IPTUN_ENCAP_DPORT:         "IPTUN_ENCAP_DPORT",
	IFLA_IPTUN_COLLECT_METADATA:    "IPTUN_COLLECT_METADATA",
	IFLA_IPTUN_FWMARK:              "IPTUN_FWMARK",
}

func (t IfIptunLinkInfoDataAttrKind) String() string {
	return elib.Stringer(ifIptunLinkInfoDataAttrKindNames, int(t))
}

type IfIptunLinkInfoDataAttrKind int
type IfIptunLinkInfoDataAttrType Empty

func NewIfIptunLinkInfoDataAttrType() *IfIptunLinkInfoDataAttrType {
	return (*IfIptunLinkInfoDataAttrType)(pool.Empty.Get().(*Empty))
}

func (t *IfIptunLinkInfoDataAttrType) attrType() {}
func (t *IfIptunLinkInfoDataAttrType) Close() error {
	repool(t)
	return nil
}
func (t *IfIptunLinkInfoDataAttrType) IthString(i int) string {
	return elib.Stringer(ifIptunLinkInfoDataAttrKindNames, i)
}

func parse_iptun_info(b []byte, linkKind InterfaceKind) (as *AttrArray) {
	as = pool.AttrArray.Get().(*AttrArray)
	as.Type = NewIfIptunLinkInfoDataAttrType()
	for i := 0; i < len(b); {
		a, v, next := nextAttr(b, i)
		i = next
		kind := IfIptunLinkInfoDataAttrKind(a.Kind())
		as.X.Validate(uint(kind))
		switch kind {
		case IFLA_IPTUN_LOCAL, IFLA_IPTUN_REMOTE:
			if linkKind == InterfaceKindIp6Tunnel {
				as.X[kind] = NewIp6AddressBytes(v)
			} else {
				as.X[kind] = NewIp4AddressBytes(v)
			}
		case IFLA_IPTUN_6RD_PREFIX:
			as.X[kind] = NewIp6AddressBytes(v)
		case IFLA_IPTUN_6RD_RELAY_PREFIX:
			as.X[kind] = NewIp4AddressBytes(v)
		case IFLA_IPTUN_TTL, IFLA_IPTUN_TOS, IFLA_IPTUN_PROTO, IFLA_IPTUN_PMTUDISC, IFLA_IPTUN_ENCAP_LIMIT, IFLA_IPTUN_FWMARK, IFLA_IPTUN_FLAGS:
			as.X[kind] = Uint8Attr(v[0])
		case IFLA_IPTUN_ENCAP_TYPE, IFLA_IPTUN_ENCAP_FLAGS, IFLA_IPTUN_ENCAP_SPORT, IFLA_IPTUN_ENCAP_DPORT, IFLA_IPTUN_6RD_PREFIXLEN, IFLA_IPTUN_6RD_RELAY_PREFIXLEN:
			as.X[kind] = Uint16AttrBytes(v)
		case IFLA_IPTUN_LINK, IFLA_IPTUN_FLOWINFO:
			as.X[kind] = Uint32AttrBytes(v)
		case IFLA_IPTUN_COLLECT_METADATA:
			as.X[kind] = Uint8Attr(0)
		default:
			panic("unknown iptun link data attribute kind " + kind.String())
		}
	}
	return as
}

const (
	IFLA_GRE_UNSPEC IfGRELinkInfoDataAttrKind = iota
	IFLA_GRE_LINK
	IFLA_GRE_IFLAGS
	IFLA_GRE_OFLAGS
	IFLA_GRE_IKEY
	IFLA_GRE_OKEY
	IFLA_GRE_LOCAL
	IFLA_GRE_REMOTE
	IFLA_GRE_TTL
	IFLA_GRE_TOS
	IFLA_GRE_PMTUDISC
	IFLA_GRE_ENCAP_LIMIT
	IFLA_GRE_FLOWINFO
	IFLA_GRE_FLAGS
	IFLA_GRE_ENCAP_TYPE
	IFLA_GRE_ENCAP_FLAGS
	IFLA_GRE_ENCAP_SPORT
	IFLA_GRE_ENCAP_DPORT
	IFLA_GRE_COLLECT_METADATA
	IFLA_GRE_IGNORE_DF
	IFLA_GRE_FWMARK
)

var ifGRELinkInfoDataAttrKindNames = []string{
	IFLA_GRE_UNSPEC:           "GRE_UNSPEC",
	IFLA_GRE_LINK:             "GRE_LINK",
	IFLA_GRE_IFLAGS:           "GRE_IFLAGS",
	IFLA_GRE_OFLAGS:           "GRE_OFLAGS",
	IFLA_GRE_IKEY:             "GRE_IKEY",
	IFLA_GRE_OKEY:             "GRE_OKEY",
	IFLA_GRE_LOCAL:            "GRE_LOCAL",
	IFLA_GRE_REMOTE:           "GRE_REMOTE",
	IFLA_GRE_TTL:              "GRE_TTL",
	IFLA_GRE_TOS:              "GRE_TOS",
	IFLA_GRE_PMTUDISC:         "GRE_PMTUDISC",
	IFLA_GRE_ENCAP_LIMIT:      "GRE_ENCAP_LIMIT",
	IFLA_GRE_FLOWINFO:         "GRE_FLOWINFO",
	IFLA_GRE_FLAGS:            "GRE_FLAGS",
	IFLA_GRE_ENCAP_TYPE:       "GRE_ENCAP_TYPE",
	IFLA_GRE_ENCAP_FLAGS:      "GRE_ENCAP_FLAGS",
	IFLA_GRE_ENCAP_SPORT:      "GRE_ENCAP_SPORT",
	IFLA_GRE_ENCAP_DPORT:      "GRE_ENCAP_DPORT",
	IFLA_GRE_COLLECT_METADATA: "GRE_COLLECT_METADATA",
	IFLA_GRE_IGNORE_DF:        "GRE_IGNORE_DF",
	IFLA_GRE_FWMARK:           "GRE_FWMARK",
}

func (t IfGRELinkInfoDataAttrKind) String() string {
	return elib.Stringer(ifGRELinkInfoDataAttrKindNames, int(t))
}

type IfGRELinkInfoDataAttrKind int
type IfGRELinkInfoDataAttrType Empty

func NewIfGRELinkInfoDataAttrType() *IfGRELinkInfoDataAttrType {
	return (*IfGRELinkInfoDataAttrType)(pool.Empty.Get().(*Empty))
}

func (t *IfGRELinkInfoDataAttrType) attrType() {}
func (t *IfGRELinkInfoDataAttrType) Close() error {
	repool(t)
	return nil
}
func (t *IfGRELinkInfoDataAttrType) IthString(i int) string {
	return elib.Stringer(ifGRELinkInfoDataAttrKindNames, i)
}

func parse_gre_info(b []byte, linkKind InterfaceKind) (as *AttrArray) {
	as = pool.AttrArray.Get().(*AttrArray)
	as.Type = NewIfGRELinkInfoDataAttrType()
	for i := 0; i < len(b); {
		a, v, next := nextAttr(b, i)
		i = next
		kind := IfGRELinkInfoDataAttrKind(a.Kind())
		as.X.Validate(uint(kind))
		switch kind {
		case IFLA_GRE_LOCAL, IFLA_GRE_REMOTE:
			switch linkKind {
			case InterfaceKindIp6GRE, InterfaceKindIp6GRETap:
				as.X[kind] = NewIp6AddressBytes(v)
			default:
				as.X[kind] = NewIp4AddressBytes(v)
			}
		case IFLA_GRE_LINK, IFLA_GRE_IKEY, IFLA_GRE_OKEY, IFLA_GRE_FLOWINFO,
			IFLA_GRE_FLAGS, IFLA_GRE_FWMARK:
			as.X[kind] = Uint32AttrBytes(v)
		case IFLA_GRE_IFLAGS, IFLA_GRE_OFLAGS,
			IFLA_GRE_ENCAP_TYPE, IFLA_GRE_ENCAP_FLAGS,
			IFLA_GRE_ENCAP_SPORT, IFLA_GRE_ENCAP_DPORT:
			as.X[kind] = Uint16AttrBytes(v)
		case IFLA_GRE_TTL, IFLA_GRE_TOS, IFLA_GRE_ENCAP_LIMIT,
			IFLA_GRE_PMTUDISC, IFLA_GRE_IGNORE_DF:
			as.X[kind] = Uint8Attr(v[0])
		case IFLA_GRE_COLLECT_METADATA:
			as.X[kind] = Uint8Attr(0)
		default:
			panic("unknown GRE link data attribute kind " + kind.String())
		}
	}
	return as
}

//go:generate gentemplate -d Package=netlink -id Attr -d VecType=AttrVec -d Type=Attr github.com/platinasystems/go/elib/vec.tmpl

func (a AttrVec) Size() (l int) {
	for i := range a {
		if a[i] != nil {
			l += SizeofNlAttr + attrAlignLen(a[i].Size())
		}
	}
	return
}

func (a AttrVec) Set(v []byte) (vi int) {
	for i := range a {
		if a[i] == nil {
			continue
		}

		s := a[i].Size()

		// Fill in attribute header.
		nla := (*NlAttr)(unsafe.Pointer(&v[vi]))
		nla.kind = uint16(i)
		nla.Len = uint16(SizeofNlAttr + s)

		// Fill in attribute value.
		a[i].Set(v[vi+SizeofNlAttr : vi+SizeofNlAttr+s])
		vi += SizeofNlAttr + attrAlignLen(s)
	}
	return
}

type AttrArray struct {
	X    AttrVec
	Type AttrType
}

func (a *AttrArray) attr() {}

func (a *AttrArray) multiline() {}

func (a *AttrArray) Close() error {
	for i, x := range a.X {
		if x != nil {
			if method, found := x.(io.Closer); found {
				method.Close()
			}
			a.X[i] = nil
		}
	}
	if method, found := a.Type.(io.Closer); found {
		method.Close()
	}
	repool(a)
	return nil
}
func (a *AttrArray) Set(v []byte) {
	a.X.Set(v)
}
func (a *AttrArray) Size() int {
	return a.X.Size()
}
func (a *AttrArray) String() string {
	return StringOf(a)
}
func (a *AttrArray) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for i, v := range a.X {
		if v == nil {
			continue
		}
		fmt.Fprint(acc, a.Type.IthString(i), ":")
		if _, found := v.(multiliner); found {
			fmt.Fprintln(acc)
			indent.Increase(acc)
			v.WriteTo(acc)
			indent.Decrease(acc)
		} else {
			fmt.Fprint(acc, " ")
			v.WriteTo(acc)
			fmt.Fprintln(acc)
		}
	}
	return acc.Tuple()
}

type EthernetAddress [6]byte

func NewEthernetAddressBytes(b []byte) *EthernetAddress {
	a := pool.EthernetAddress.Get().(*EthernetAddress)
	a.Parse(b)
	return a
}

func (a *EthernetAddress) attr() {}
func (a *EthernetAddress) Bytes() []byte {
	return a[:]
}
func (a *EthernetAddress) Close() error {
	repool(a)
	return nil
}
func (a *EthernetAddress) Parse(b []byte) {
	copy(a[:], b[:6])
}
func (a *EthernetAddress) Set(v []byte) {
	copy(v, a[:])
}
func (a *EthernetAddress) Size() int {
	return len(a)
}
func (a *EthernetAddress) String() string {
	return StringOf(a)
}
func (a *EthernetAddress) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintf(acc, "%02x:%02x:%02x:%02x:%02x:%02x",
		a[0], a[1], a[2], a[3], a[4], a[5])
	return acc.Tuple()
}

type HexStringAttr bytes.Buffer

func NewHexStringAttrBytes(b []byte) *HexStringAttr {
	h := (*HexStringAttr)(pool.Bytes.Get().(*bytes.Buffer))
	h.Parse(b)
	return h
}

func (a *HexStringAttr) attr() {}
func (a *HexStringAttr) Buffer() *bytes.Buffer {
	return (*bytes.Buffer)(a)
}
func (a *HexStringAttr) Close() error {
	repool(a.Buffer())
	return nil
}
func (a *HexStringAttr) Parse(b []byte) {
	a.Buffer().Write(b)
}
func (a *HexStringAttr) Set(v []byte) {
	copy(v, a.Buffer().Bytes())
}
func (a *HexStringAttr) Size() int {
	return a.Buffer().Len()
}
func (a *HexStringAttr) String() string {
	return StringOf(a)
}
func (a *HexStringAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, hex.EncodeToString(a.Buffer().Bytes()))
	return acc.Tuple()
}

type IfAddrCacheInfo struct {
	Prefered         uint32
	Valid            uint32
	CreatedTimestamp uint32 /* created timestamp, hundredths of seconds */
	UpdatedTimestamp uint32 /* updated timestamp, hundredths of seconds */
}

func NewIfAddrCacheInfoBytes(b []byte) *IfAddrCacheInfo {
	a := pool.IfAddrCacheInfo.Get().(*IfAddrCacheInfo)
	a.Parse(b)
	return a
}

func (a *IfAddrCacheInfo) attr() {}

func (a *IfAddrCacheInfo) multiline() {}

func (a *IfAddrCacheInfo) Close() error {
	repool(a)
	return nil
}
func (a *IfAddrCacheInfo) Set(v []byte) {
	panic("should never be called")
}
func (a *IfAddrCacheInfo) Size() int {
	panic("should never be called")
	return 0
}
func (a *IfAddrCacheInfo) String() string {
	return StringOf(a)
}
func (a *IfAddrCacheInfo) Parse(b []byte) {
	*a = *(*IfAddrCacheInfo)(unsafe.Pointer(&b[0]))
}
func (a *IfAddrCacheInfo) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintln(acc, "prefered:", a.Prefered)
	fmt.Fprintln(acc, "valid:", a.Valid)
	fmt.Fprintln(acc, "created:", a.CreatedTimestamp)
	fmt.Fprintln(acc, "updated:", a.UpdatedTimestamp)
	return acc.Tuple()
}

type IfAddrFlagAttr uint32

func IfAddrFlagAttrBytes(b []byte) IfAddrFlagAttr {
	return *(*IfAddrFlagAttr)(unsafe.Pointer(&b[0]))
}

func (a IfAddrFlagAttr) attr() {}
func (a IfAddrFlagAttr) Size() int {
	return 4
}
func (a IfAddrFlagAttr) Set(v []byte) {
	*(*IfAddrFlagAttr)(unsafe.Pointer(&v[0])) = a
}
func (a IfAddrFlagAttr) String() string {
	return IfAddrFlags(a).String()
}
func (a IfAddrFlagAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, IfAddrFlags(a))
	return acc.Tuple()
}

type IfOperState uint8

func (a IfOperState) attr() {}
func (a IfOperState) Set(v []byte) {
	v[0] = byte(a)
}
func (a IfOperState) Size() int {
	return 1
}
func (a IfOperState) String() string {
	if int(a) >= len(ifOperStates) {
		a = 0
	}
	return ifOperStates[a]
}
func (a IfOperState) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a)
	return acc.Tuple()
}

type Int8Attr int8

func (a Int8Attr) attr() {}
func (a Int8Attr) Rune() rune {
	return rune(a)
}
func (a Int8Attr) Set(v []byte) {
	v[0] = byte(a)
}
func (a Int8Attr) Size() int {
	return 1
}
func (a Int8Attr) String() string {
	return StringOf(a)
}
func (a Int8Attr) Int() int8 {
	return int8(a)
}
func (a Int8Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Int())
	return acc.Tuple()
}

type Int16Attr int16

func Int16AttrBytes(b []byte) Int16Attr {
	return Int16Attr(*(*int16)(unsafe.Pointer(&b[0])))
}

func (a Int16Attr) attr() {}
func (a Int16Attr) Set(v []byte) {
	*(*Int16Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Int16Attr) Size() int {
	return 2
}
func (a Int16Attr) String() string {
	return StringOf(a)
}
func (a Int16Attr) Int() int16 {
	return int16(a)
}
func (a Int16Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Int())
	return acc.Tuple()
}

type Int32Attr int32

func Int32AttrBytes(b []byte) Int32Attr {
	return Int32Attr(*(*int32)(unsafe.Pointer(&b[0])))
}

func (a Int32Attr) attr() {}
func (a Int32Attr) Set(v []byte) {
	*(*Int32Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Int32Attr) Size() int {
	return 4
}
func (a Int32Attr) String() string {
	return StringOf(a)
}
func (a Int32Attr) Int() int32 {
	return int32(a)
}
func (a Int32Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Int())
	return acc.Tuple()
}

type Int64Attr int64

func Int64AttrBytes(b []byte) Int64Attr {
	return Int64Attr(*(*int64)(unsafe.Pointer(&b[0])))
}

func (a Int64Attr) attr() {}
func (a Int64Attr) Set(v []byte) {
	*(*Int64Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Int64Attr) Size() int {
	return 8
}
func (a Int64Attr) String() string {
	return StringOf(a)
}
func (a Int64Attr) Int() int64 {
	return int64(a)
}
func (a Int64Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Int())
	return acc.Tuple()
}

type Ip4Address [4]byte

func NewIp4AddressBytes(b []byte) *Ip4Address {
	a := pool.Ip4Address.Get().(*Ip4Address)
	a.Parse(b)
	return a
}

func (a *Ip4Address) attr() {}
func (a *Ip4Address) Bytes() []byte {
	return a[:]
}
func (a *Ip4Address) Close() error {
	repool(a)
	return nil
}
func (a *Ip4Address) Parse(b []byte) {
	copy(a[:], b[:4])
}
func (a *Ip4Address) Set(v []byte) {
	copy(v, a[:])
}
func (a *Ip4Address) Size() int {
	return len(a)
}
func (a *Ip4Address) String() string {
	return StringOf(a)
}
func (a *Ip4Address) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintf(acc, "%d.%d.%d.%d", a[0], a[1], a[2], a[3])
	return acc.Tuple()
}

type Ip4DevConf [IPV4_DEVCONF_MAX]uint32

func NewIp4DevConfBytes(b []byte) *Ip4DevConf {
	a := pool.Ip4DevConf.Get().(*Ip4DevConf)
	a.Parse(b)
	return a
}

func (a *Ip4DevConf) attr() {}

func (a *Ip4DevConf) multiline() {}

func (a *Ip4DevConf) Close() error {
	repool(a)
	return nil
}
func (a *Ip4DevConf) Parse(b []byte) {
	*a = *(*Ip4DevConf)(unsafe.Pointer(&b[0]))
}
func (a *Ip4DevConf) Set(v []byte) {
	panic("not implemented")
}
func (a *Ip4DevConf) Size() int {
	panic("not implemented")
	return 0
}
func (a *Ip4DevConf) String() string {
	return StringOf(a)
}
func (a *Ip4DevConf) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	indent.Increase(acc)
	for i, v := range a {
		if v != 0 {
			fmt.Fprint(acc, Ip4DevConfKind(i), ": ", v, "\n")
		}
	}
	indent.Decrease(acc)
	return acc.Tuple()
}

type Ip6Address [16]byte

func NewIp6AddressBytes(b []byte) *Ip6Address {
	a := pool.Ip6Address.Get().(*Ip6Address)
	a.Parse(b)
	return a
}

func (a *Ip6Address) attr() {}
func (a *Ip6Address) Bytes() []byte {
	return a[:]
}
func (a *Ip6Address) Close() error {
	repool(a)
	return nil
}
func (a *Ip6Address) Parse(b []byte) {
	copy(a[:], b[:16])
}
func (a *Ip6Address) Set(v []byte) {
	copy(v, a[:])
}
func (a *Ip6Address) Size() int {
	return len(a)
}
func (a *Ip6Address) String() string {
	return StringOf(a)
}
func (a *Ip6Address) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, net.IP(a[:]))
	return acc.Tuple()
}

type Ip6DevConf [IPV6_DEVCONF_MAX]uint32

func NewIp6DevConfBytes(b []byte) *Ip6DevConf {
	a := pool.Ip6DevConf.Get().(*Ip6DevConf)
	a.Parse(b)
	return a
}

func (a *Ip6DevConf) attr() {}

func (a *Ip6DevConf) multiline() {}

func (a *Ip6DevConf) Close() error {
	repool(a)
	return nil
}
func (a *Ip6DevConf) Parse(b []byte) {
	*a = *(*Ip6DevConf)(unsafe.Pointer(&b[0]))
}
func (a *Ip6DevConf) Set(v []byte) {
	panic("not implemented")
}
func (a *Ip6DevConf) Size() int {
	panic("not implemented")
	return 0
}
func (a *Ip6DevConf) String() string {
	return StringOf(a)
}
func (a *Ip6DevConf) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	indent.Increase(acc)
	for i, v := range a {
		if v != 0 {
			fmt.Fprint(acc, Ip6DevConfKind(i), ": ", v, "\n")
		}
	}
	indent.Decrease(acc)
	return acc.Tuple()
}

type Ip6IfFlagsAttr uint32

func Ip6IfFlagsAttrBytes(b []byte) Ip6IfFlagsAttr {
	return Ip6IfFlagsAttr(*(*uint32)(unsafe.Pointer(&b[0])))
}

func (a Ip6IfFlagsAttr) attr() {}
func (a Ip6IfFlagsAttr) Set(v []byte) {
	*(*Ip6IfFlagsAttr)(unsafe.Pointer(&v[0])) = a
}
func (a Ip6IfFlagsAttr) Size() int {
	return 4
}
func (a Ip6IfFlagsAttr) String() string {
	return StringOf(a)
}
func (a Ip6IfFlagsAttr) Uint() uint32 {
	return uint32(a)
}
func (a Ip6IfFlagsAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for _, match := range []struct {
		bit  Ip6IfFlagsAttr
		name string
	}{
		{INET6_IF_PREFIX_ONLINK, "ONLINK"},
		{INET6_IF_PREFIX_AUTOCONF, "AUTOCONF"},
		{INET6_IF_RA_OTHERCONF, "OTHERCONF"},
		{INET6_IF_RA_MANAGED, "MANAGED"},
		{INET6_IF_RA_RCVD, "RCVD"},
		{INET6_IF_RS_SENT, "SENT"},
		{INET6_IF_READY, "READY"},
	} {
		if a&match.bit == match.bit {
			if acc.Total() > 0 {
				fmt.Fprint(acc, " | ")
			}
			fmt.Fprint(acc, match.name)
		}
	}
	return acc.Tuple()
}

type LinkStats [N_link_stat]uint32

func NewLinkStatsBytes(b []byte) *LinkStats {
	a := pool.LinkStats.Get().(*LinkStats)
	a.Parse(b)
	return a
}

func (a *LinkStats) attr() {}

func (a *LinkStats) multiline() {}

func (a *LinkStats) Close() error {
	repool(a)
	return nil
}
func (a *LinkStats) Parse(b []byte) {
	*a = *(*LinkStats)(unsafe.Pointer(&b[0]))
}
func (a *LinkStats) Set(v []byte) {
	*(*LinkStats)(unsafe.Pointer(&v[0])) = *a
}
func (a *LinkStats) Size() int {
	return int(N_link_stat) * 4
}
func (a *LinkStats) String() string {
	return StringOf(a)
}
func (a *LinkStats) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for i, v := range a {
		t := LinkStatType(i)
		if v != 0 || t == Rx_packets || t == Tx_packets {
			fmt.Fprint(acc, t, ": ", v, "\n")
		}
	}
	return acc.Tuple()
}

type LinkStats64 [N_link_stat]uint64

func NewLinkStats64Bytes(b []byte) *LinkStats64 {
	a := pool.LinkStats64.Get().(*LinkStats64)
	a.Parse(b)
	return a
}

func (a *LinkStats64) attr() {}

func (a *LinkStats64) multiline() {}

func (a *LinkStats64) Close() error {
	repool(a)
	return nil
}
func (a *LinkStats64) Parse(b []byte) {
	*a = *(*LinkStats64)(unsafe.Pointer(&b[0]))
}
func (a *LinkStats64) Set(v []byte) {
	*(*LinkStats64)(unsafe.Pointer(&v[0])) = *a
}
func (a *LinkStats64) Size() int {
	return int(N_link_stat) * 8
}
func (a *LinkStats64) String() string {
	return StringOf(a)
}
func (a *LinkStats64) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	for i, v := range a {
		t := LinkStatType(i)
		if v != 0 || t == Rx_packets || t == Tx_packets {
			fmt.Fprint(acc, t, ": ", v, "\n")
		}
	}
	return acc.Tuple()
}

type NdaCacheInfo struct {
	Confirmed uint32
	Used      uint32
	Updated   uint32
	RefCnt    uint32
}

func NewNdaCacheInfoBytes(b []byte) *NdaCacheInfo {
	a := pool.NdaCacheInfo.Get().(*NdaCacheInfo)
	a.Parse(b)
	return a
}

func (a *NdaCacheInfo) attr() {}

func (a *NdaCacheInfo) multiline() {}

func (a *NdaCacheInfo) Close() error {
	repool(a)
	return nil
}
func (a *NdaCacheInfo) Parse(b []byte) {
	*a = *(*NdaCacheInfo)(unsafe.Pointer(&b[0]))
}
func (a *NdaCacheInfo) Set(v []byte) {
	panic("should never be called")
}
func (a *NdaCacheInfo) Size() int {
	panic("should never be called")
	return 0
}
func (a *NdaCacheInfo) String() string {
	return StringOf(a)
}
func (a *NdaCacheInfo) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintln(acc, "confirmed:", a.Confirmed)
	fmt.Fprintln(acc, "used:", a.Used)
	fmt.Fprintln(acc, "updated:", a.Updated)
	fmt.Fprintln(acc, "refcnt:", a.RefCnt)
	return acc.Tuple()
}

type RtaCacheInfo struct {
	ClntRef uint32
	LastUse uint32
	Expires uint32
	Error   uint32
	Used    uint32
}

func NewRtaCacheInfoBytes(b []byte) *RtaCacheInfo {
	a := pool.RtaCacheInfo.Get().(*RtaCacheInfo)
	a.Parse(b)
	return a
}

func (a *RtaCacheInfo) attr() {}

func (a *RtaCacheInfo) multiline() {}

func (a *RtaCacheInfo) Close() error {
	repool(a)
	return nil
}
func (a *RtaCacheInfo) Set(v []byte) {
	panic("should never be called")
}
func (a *RtaCacheInfo) Size() int {
	panic("should never be called")
	return 0
}
func (a *RtaCacheInfo) String() string {
	return StringOf(a)
}
func (a *RtaCacheInfo) Parse(b []byte) {
	*a = *(*RtaCacheInfo)(unsafe.Pointer(&b[0]))
}
func (a *RtaCacheInfo) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprintln(acc, "clntref:", a.ClntRef)
	fmt.Fprintln(acc, "lastuse:", a.LastUse)
	fmt.Fprintln(acc, "expires:", a.Expires)
	fmt.Fprintln(acc, "error:", a.Error)
	fmt.Fprintln(acc, "used:", a.Used)
	return acc.Tuple()
}

type StringAttr string

func StringAttrBytes(b []byte) StringAttr {
	return StringAttr(string(b))
}
func (a StringAttr) attr() {}
func (a StringAttr) Size() int {
	return len(a) + 1
}
func (a StringAttr) Set(v []byte) {
	copy(v, a)
	v = append(v, 0)
}
func (a StringAttr) String() string {
	return string(a)
}
func (a StringAttr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a)
	return acc.Tuple()
}

type Uint8Attr uint8

func (a Uint8Attr) attr() {}
func (a Uint8Attr) Rune() rune {
	return rune(a)
}
func (a Uint8Attr) Set(v []byte) {
	v[0] = byte(a)
}
func (a Uint8Attr) Size() int {
	return 1
}
func (a Uint8Attr) String() string {
	return StringOf(a)
}
func (a Uint8Attr) Uint() uint8 {
	return uint8(a)
}
func (a Uint8Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Uint())
	return acc.Tuple()
}

type Uint16Attr uint16

func Uint16AttrBytes(b []byte) Uint16Attr {
	return Uint16Attr(*(*uint16)(unsafe.Pointer(&b[0])))
}

func (a Uint16Attr) attr() {}
func (a Uint16Attr) Set(v []byte) {
	*(*Uint16Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Uint16Attr) Size() int {
	return 2
}
func (a Uint16Attr) String() string {
	return StringOf(a)
}
func (a Uint16Attr) Uint() uint16 {
	return uint16(a)
}
func (a Uint16Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Uint())
	return acc.Tuple()
}

type Uint32Attr uint32

func Uint32AttrBytes(b []byte) Uint32Attr {
	return Uint32Attr(*(*uint32)(unsafe.Pointer(&b[0])))
}

func (a Uint32Attr) attr() {}
func (a Uint32Attr) Set(v []byte) {
	*(*Uint32Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Uint32Attr) Size() int {
	return 4
}
func (a Uint32Attr) String() string {
	return StringOf(a)
}
func (a Uint32Attr) Uint() uint32 {
	return uint32(a)
}
func (a Uint32Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Uint())
	return acc.Tuple()
}

type Uint64Attr uint64

func Uint64AttrBytes(b []byte) Uint64Attr {
	return Uint64Attr(*(*uint64)(unsafe.Pointer(&b[0])))
}

func (a Uint64Attr) attr() {}
func (a Uint64Attr) Set(v []byte) {
	*(*Uint64Attr)(unsafe.Pointer(&v[0])) = a
}
func (a Uint64Attr) Size() int {
	return 8
}
func (a Uint64Attr) String() string {
	return StringOf(a)
}
func (a Uint64Attr) Uint() uint64 {
	return uint64(a)
}
func (a Uint64Attr) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, a.Uint())
	return acc.Tuple()
}

type LwtunnelEncapType uint8

const (
	LWTUNNEL_ENCAP_NONE = iota
	LWTUNNEL_ENCAP_MPLS
	LWTUNNEL_ENCAP_IP
	LWTUNNEL_ENCAP_ILA
	LWTUNNEL_ENCAP_IP6
	LWTUNNEL_ENCAP_SEG6
	LWTUNNEL_ENCAP_BPF
)

var LwtunnelEncapNames = []string{
	LWTUNNEL_ENCAP_NONE: "NONE",
	LWTUNNEL_ENCAP_MPLS: "MPLS",
	LWTUNNEL_ENCAP_IP:   "IP",
	LWTUNNEL_ENCAP_ILA:  "ILA",
	LWTUNNEL_ENCAP_IP6:  "IP6",
	LWTUNNEL_ENCAP_SEG6: "SEG6",
	LWTUNNEL_ENCAP_BPF:  "BPF",
}

func (x LwtunnelEncapType) attr()          {}
func (x LwtunnelEncapType) String() string { return elib.Stringer(LwtunnelEncapNames, int(x)) }
func (x LwtunnelEncapType) Uint() uint8    { return uint8(x) }
func (x LwtunnelEncapType) Size() int      { return 1 }
func (x LwtunnelEncapType) Set(v []byte)   { *(*LwtunnelEncapType)(unsafe.Pointer(&v[0])) = x }
func (x LwtunnelEncapType) WriteTo(w io.Writer) (int64, error) {
	acc := accumulate.New(w)
	defer acc.Fini()
	fmt.Fprint(acc, x.String())
	return acc.Tuple()
}

type LwtunnelIp4AttrKind uint8

const (
	LWTUNNEL_IP_UNSPEC LwtunnelIp4AttrKind = iota
	LWTUNNEL_IP_ID
	LWTUNNEL_IP_DST
	LWTUNNEL_IP_SRC
	LWTUNNEL_IP_TTL
	LWTUNNEL_IP_TOS
	LWTUNNEL_IP_FLAGS
)

var LwtunnelIp4AttrKindNames = []string{
	LWTUNNEL_IP_UNSPEC: "UNSPEC",
	LWTUNNEL_IP_ID:     "ID",
	LWTUNNEL_IP_DST:    "DST",
	LWTUNNEL_IP_SRC:    "SRC",
	LWTUNNEL_IP_TTL:    "TTL",
	LWTUNNEL_IP_TOS:    "TOS",
	LWTUNNEL_IP_FLAGS:  "FLAGS",
}

func (x LwtunnelIp4AttrKind) String() string { return elib.Stringer(LwtunnelIp4AttrKindNames, int(x)) }

type LwtunnelIp4EncapKindAttrType Empty

func NewLwtunnelIp4EncapKindAttrType() *LwtunnelIp4EncapKindAttrType {
	return (*LwtunnelIp4EncapKindAttrType)(pool.Empty.Get().(*Empty))
}

func (t *LwtunnelIp4EncapKindAttrType) attrType() {}
func (t *LwtunnelIp4EncapKindAttrType) Close() error {
	repool(t)
	return nil
}
func (t *LwtunnelIp4EncapKindAttrType) IthString(i int) string {
	return elib.Stringer(LwtunnelIp4AttrKindNames, i)
}

type LwtunnelIp4EncapAttrType Empty

func NewLwtunnelIp4EncapAttrType() *LwtunnelIp4EncapAttrType {
	return (*LwtunnelIp4EncapAttrType)(pool.Empty.Get().(*Empty))
}

func (t *LwtunnelIp4EncapAttrType) attrType() {}
func (t *LwtunnelIp4EncapAttrType) Close() error {
	repool(t)
	return nil
}
func (t *LwtunnelIp4EncapAttrType) IthString(i int) string {
	return elib.Stringer(LwtunnelIp4AttrKindNames, i)
}

func parse_lwtunnel_ip4_encap(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewLwtunnelIp4EncapAttrType()
	for i := 0; i < len(b); {
		a, v, next := nextAttr(b, i)
		i = next
		kind := LwtunnelIp4AttrKind(a.Kind())
		as.X.Validate(uint(kind))
		switch kind {
		case LWTUNNEL_IP_ID:
			as.X[kind] = Uint64AttrBytes(v[:])
		case LWTUNNEL_IP_DST, LWTUNNEL_IP_SRC:
			as.X[kind] = NewIp4AddressBytes(v)
		case LWTUNNEL_IP_TTL, LWTUNNEL_IP_TOS:
			as.X[kind] = Uint8Attr(v[0])
		case LWTUNNEL_IP_FLAGS:
			as.X[kind] = Uint16Attr(v[0])
		default:
			panic("unknown ip tunnel encap kind " + kind.String())
		}
	}
	return as
}

type LwtunnelIp6AttrKind uint8

const (
	LWTUNNEL_IP6_UNSPEC LwtunnelIp6AttrKind = iota
	LWTUNNEL_IP6_ID
	LWTUNNEL_IP6_DST
	LWTUNNEL_IP6_SRC
	LWTUNNEL_IP6_HOPLIMIT
	LWTUNNEL_IP6_TC
	LWTUNNEL_IP6_FLAGS
)

var LwtunnelIp6AttrKindNames = []string{
	LWTUNNEL_IP6_UNSPEC:   "UNSPEC",
	LWTUNNEL_IP6_ID:       "ID",
	LWTUNNEL_IP6_DST:      "DST",
	LWTUNNEL_IP6_SRC:      "SRC",
	LWTUNNEL_IP6_HOPLIMIT: "HOP LIMIT",
	LWTUNNEL_IP6_TC:       "TC",
	LWTUNNEL_IP6_FLAGS:    "FLAGS",
}

func (x LwtunnelIp6AttrKind) String() string { return elib.Stringer(LwtunnelIp6AttrKindNames, int(x)) }

type LwtunnelIp6EncapKindAttrType Empty

func NewLwtunnelIp6EncapKindAttrType() *LwtunnelIp6EncapKindAttrType {
	return (*LwtunnelIp6EncapKindAttrType)(pool.Empty.Get().(*Empty))
}

func (t *LwtunnelIp6EncapKindAttrType) attrType() {}
func (t *LwtunnelIp6EncapKindAttrType) Close() error {
	repool(t)
	return nil
}
func (t *LwtunnelIp6EncapKindAttrType) IthString(i int) string {
	return elib.Stringer(LwtunnelIp6AttrKindNames, i)
}

type LwtunnelIp6EncapAttrType Empty

func NewLwtunnelIp6EncapAttrType() *LwtunnelIp6EncapAttrType {
	return (*LwtunnelIp6EncapAttrType)(pool.Empty.Get().(*Empty))
}

func (t *LwtunnelIp6EncapAttrType) attrType() {}
func (t *LwtunnelIp6EncapAttrType) Close() error {
	repool(t)
	return nil
}
func (t *LwtunnelIp6EncapAttrType) IthString(i int) string {
	return elib.Stringer(LwtunnelIp6AttrKindNames, i)
}

func parse_lwtunnel_ip6_encap(b []byte) *AttrArray {
	as := pool.AttrArray.Get().(*AttrArray)
	as.Type = NewLwtunnelIp6EncapAttrType()
	for i := 0; i < len(b); {
		a, v, next := nextAttr(b, i)
		i = next
		kind := LwtunnelIp6AttrKind(a.Kind())
		as.X.Validate(uint(kind))
		switch kind {
		case LWTUNNEL_IP6_ID:
			as.X[kind] = Uint64AttrBytes(v[:])
		case LWTUNNEL_IP6_DST, LWTUNNEL_IP6_SRC:
			as.X[kind] = NewIp6AddressBytes(v)
		case LWTUNNEL_IP6_HOPLIMIT, LWTUNNEL_IP6_TC:
			as.X[kind] = Uint8Attr(v[0])
		case LWTUNNEL_IP6_FLAGS:
			as.X[kind] = Uint16Attr(v[0])
		default:
			panic("unknown ip6 tunnel encap kind " + kind.String())
		}
	}
	return as
}
