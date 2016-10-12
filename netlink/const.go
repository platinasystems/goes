// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package netlink

import (
	"fmt"

	"github.com/platinasystems/go/elib"
)

type Empty struct{}

type MsgType uint16

// Header type field values
const (
	NLMSG_NOOP    MsgType = 0x1 /* Nothing */
	NLMSG_ERROR   MsgType = 0x2 /* Error */
	NLMSG_DONE    MsgType = 0x3 /* End of a dump */
	NLMSG_OVERRUN MsgType = 0x4 /* Data lost */

	RTM_NEWLINK MsgType = 16
	RTM_DELLINK MsgType = 17
	RTM_GETLINK MsgType = 18
	RTM_SETLINK MsgType = 19

	RTM_NEWADDR MsgType = 20
	RTM_DELADDR MsgType = 21
	RTM_GETADDR MsgType = 22

	RTM_NEWROUTE MsgType = 24
	RTM_DELROUTE MsgType = 25
	RTM_GETROUTE MsgType = 26

	RTM_NEWNEIGH MsgType = 28
	RTM_DELNEIGH MsgType = 29
	RTM_GETNEIGH MsgType = 30

	RTM_NEWRULE MsgType = 32
	RTM_DELRULE MsgType = 33
	RTM_GETRULE MsgType = 34

	RTM_NEWQDISC MsgType = 36
	RTM_DELQDISC MsgType = 37
	RTM_GETQDISC MsgType = 38

	RTM_NEWTCLASS MsgType = 40
	RTM_DELTCLASS MsgType = 41
	RTM_GETTCLASS MsgType = 42

	RTM_NEWTFILTER MsgType = 44
	RTM_DELTFILTER MsgType = 45
	RTM_GETTFILTER MsgType = 46

	RTM_NEWACTION MsgType = 48
	RTM_DELACTION MsgType = 49
	RTM_GETACTION MsgType = 50

	RTM_NEWPREFIX MsgType = 52

	RTM_GETMULTICAST MsgType = 58

	RTM_GETANYCAST MsgType = 62

	RTM_NEWNEIGHTBL MsgType = 64
	RTM_GETNEIGHTBL MsgType = 66
	RTM_SETNEIGHTBL MsgType = 67

	RTM_NEWNDUSEROPT MsgType = 68

	RTM_NEWADDRLABEL MsgType = 72
	RTM_DELADDRLABEL MsgType = 73
	RTM_GETADDRLABEL MsgType = 74

	RTM_GETDCB MsgType = 78
	RTM_SETDCB MsgType = 79

	RTM_NEWNETCONF MsgType = 80
	RTM_GETNETCONF MsgType = 82

	RTM_NEWMDB MsgType = 84
	RTM_DELMDB MsgType = 85
	RTM_GETMDB MsgType = 86

	RTM_NEWNSID MsgType = 88
	RTM_DELNSID MsgType = 89
	RTM_GETNSID MsgType = 90
)

var msgTypeNames = []string{
	NLMSG_NOOP:       "NLMSG_NOOP",
	NLMSG_ERROR:      "NLMSG_ERROR",
	NLMSG_DONE:       "NLMSG_DONE",
	NLMSG_OVERRUN:    "NLMSG_OVERRUN",
	RTM_NEWLINK:      "RTM_NEWLINK",
	RTM_DELLINK:      "RTM_DELLINK",
	RTM_GETLINK:      "RTM_GETLINK",
	RTM_SETLINK:      "RTM_SETLINK",
	RTM_NEWADDR:      "RTM_NEWADDR",
	RTM_DELADDR:      "RTM_DELADDR",
	RTM_GETADDR:      "RTM_GETADDR",
	RTM_NEWROUTE:     "RTM_NEWROUTE",
	RTM_DELROUTE:     "RTM_DELROUTE",
	RTM_GETROUTE:     "RTM_GETROUTE",
	RTM_NEWNEIGH:     "RTM_NEWNEIGH",
	RTM_DELNEIGH:     "RTM_DELNEIGH",
	RTM_GETNEIGH:     "RTM_GETNEIGH",
	RTM_NEWRULE:      "RTM_NEWRULE",
	RTM_DELRULE:      "RTM_DELRULE",
	RTM_GETRULE:      "RTM_GETRULE",
	RTM_NEWQDISC:     "RTM_NEWQDISC",
	RTM_DELQDISC:     "RTM_DELQDISC",
	RTM_GETQDISC:     "RTM_GETQDISC",
	RTM_NEWTCLASS:    "RTM_NEWTCLASS",
	RTM_DELTCLASS:    "RTM_DELTCLASS",
	RTM_GETTCLASS:    "RTM_GETTCLASS",
	RTM_NEWTFILTER:   "RTM_NEWTFILTER",
	RTM_DELTFILTER:   "RTM_DELTFILTER",
	RTM_GETTFILTER:   "RTM_GETTFILTER",
	RTM_NEWACTION:    "RTM_NEWACTION",
	RTM_DELACTION:    "RTM_DELACTION",
	RTM_GETACTION:    "RTM_GETACTION",
	RTM_NEWPREFIX:    "RTM_NEWPREFIX",
	RTM_GETMULTICAST: "RTM_GETMULTICAST",
	RTM_GETANYCAST:   "RTM_GETANYCAST",
	RTM_NEWNEIGHTBL:  "RTM_NEWNEIGHTBL",
	RTM_GETNEIGHTBL:  "RTM_GETNEIGHTBL",
	RTM_SETNEIGHTBL:  "RTM_SETNEIGHTBL",
	RTM_NEWNDUSEROPT: "RTM_NEWNDUSEROPT",
	RTM_NEWADDRLABEL: "RTM_NEWADDRLABEL",
	RTM_DELADDRLABEL: "RTM_DELADDRLABEL",
	RTM_GETADDRLABEL: "RTM_GETADDRLABEL",
	RTM_GETDCB:       "RTM_GETDCB",
	RTM_SETDCB:       "RTM_SETDCB",
	RTM_NEWNETCONF:   "RTM_NEWNETCONF",
	RTM_GETNETCONF:   "RTM_GETNETCONF",
	RTM_NEWMDB:       "RTM_NEWMDB",
	RTM_DELMDB:       "RTM_DELMDB",
	RTM_GETMDB:       "RTM_GETMDB",
	RTM_NEWNSID:      "RTM_NEWNSID",
	RTM_DELNSID:      "RTM_DELNSID",
	RTM_GETNSID:      "RTM_GETNSID",
}

func (x MsgType) String() string { return elib.Stringer(msgTypeNames, int(x)) }

type MulticastGroup uint32

// Multicast groups
const (
	RTNLGRP_LINK MulticastGroup = iota
	RTNLGRP_NOTIFY
	RTNLGRP_NEIGH
	RTNLGRP_TC
	RTNLGRP_IPV4_IFADDR
	RTNLGRP_IPV4_MROUTE
	RTNLGRP_IPV4_ROUTE
	RTNLGRP_IPV4_RULE
	RTNLGRP_IPV6_IFADDR
	RTNLGRP_IPV6_MROUTE
	RTNLGRP_IPV6_ROUTE
	RTNLGRP_IPV6_IFINFO
	RTNLGRP_DECnet_IFADDR
	RTNLGRP_NOP2
	RTNLGRP_DECnet_ROUTE
	RTNLGRP_DECnet_RULE
	RTNLGRP_NOP4
	RTNLGRP_IPV6_PREFIX
	RTNLGRP_IPV6_RULE
	RTNLGRP_ND_USEROPT
	RTNLGRP_PHONET_IFADDR
	RTNLGRP_PHONET_ROUTE
	RTNLGRP_DCB
	RTNLGRP_IPV4_NETCONF
	RTNLGRP_IPV6_NETCONF
	RTNLGRP_MDB
	RTNLGRP_MPLS_ROUTE
	RTNLGRP_NSID
	NOOP_RTNLGRP
)

type HeaderFlags uint16

const (
	NLM_F_REQUEST   HeaderFlags = 1  /* It is request message. */
	NLM_F_MULTI     HeaderFlags = 2  /* Multipart message, terminated by NLMSG_DONE */
	NLM_F_ACK       HeaderFlags = 4  /* Reply with ack, with zero or error code */
	NLM_F_ECHO      HeaderFlags = 8  /* Echo this request 		*/
	NLM_F_DUMP_INTR HeaderFlags = 16 /* Dump was inconsistent due to sequence change */

	/* Modifiers to GET request */
	NLM_F_ROOT   HeaderFlags = 0x100
	NLM_F_MATCH  HeaderFlags = 0x200 /* return all matching	*/
	NLM_F_ATOMIC HeaderFlags = 0x400 /* atomic GET		*/
	NLM_F_DUMP   HeaderFlags = NLM_F_ROOT | NLM_F_MATCH

	/* Modifiers to NEW request */
	NLM_F_REPLACE HeaderFlags = 0x100 /* Override existing		*/
	NLM_F_EXCL    HeaderFlags = 0x200 /* Do not touch, if it exists	*/
	NLM_F_CREATE  HeaderFlags = 0x400 /* Create, if it does not exist	*/
	NLM_F_APPEND  HeaderFlags = 0x800 /* Add to end of list		*/
)

var headerFlagNames = []string{
	0: "Request",
	1: "Multipart",
	2: "ACK",
	3: "Echo",
	4: "Interrupt",
}

func (x HeaderFlags) String() string { return elib.FlagStringer(headerFlagNames, elib.Word(x)) }

const NLMSG_ALIGNTO = 4
const RTA_ALIGNTO = 4

// Round the length of a netlink message up to align it properly.
func messageAlignLen(l int) int {
	return (l + NLMSG_ALIGNTO - 1) & ^(NLMSG_ALIGNTO - 1)
}

// Round the length of a netlink route attribute up to align it
// properly.
func attrAlignLen(l int) int {
	return (l + RTA_ALIGNTO - 1) & ^(RTA_ALIGNTO - 1)
}

type NlAttr struct {
	Len  uint16
	Kind uint16
}

const SizeofNlAttr = 4

type RtScope uint8

const (
	RT_SCOPE_UNIVERSE RtScope = 0
	// User defined values
	RT_SCOPE_SITE    RtScope = 200
	RT_SCOPE_LINK    RtScope = 253
	RT_SCOPE_HOST    RtScope = 254
	RT_SCOPE_NOWHERE RtScope = 255
)

var scopeNames = []string{
	RT_SCOPE_UNIVERSE: "Universe",
	RT_SCOPE_SITE:     "Site",
	RT_SCOPE_LINK:     "Link",
	RT_SCOPE_HOST:     "Host",
	RT_SCOPE_NOWHERE:  "Nowhere",
}

func (s RtScope) String() string { return elib.Stringer(scopeNames, int(s)) }
func (s RtScope) Uint() uint8    { return uint8(s) }

type RouteType uint8

const (
	RTN_UNSPEC RouteType = iota
	RTN_UNICAST
	RTN_LOCAL
	RTN_BROADCAST
	RTN_ANYCAST
	RTN_MULTICAST
	RTN_BLACKHOLE
	RTN_UNREACHABLE
	RTN_PROHIBIT
	RTN_THROW
	RTN_NAT
	RTN_XRESOLVE
)

var routeTypeNames = []string{
	RTN_UNSPEC:      "UNSPEC",
	RTN_UNICAST:     "UNICAST",
	RTN_LOCAL:       "LOCAL",
	RTN_BROADCAST:   "BROADCAST",
	RTN_ANYCAST:     "ANYCAST",
	RTN_MULTICAST:   "MULTICAST",
	RTN_BLACKHOLE:   "DROP",
	RTN_UNREACHABLE: "UNREACHABLE",
	RTN_PROHIBIT:    "PROHIBIT",
	RTN_THROW:       "THROW",
	RTN_NAT:         "NAT",
	RTN_XRESOLVE:    "XRESOLVE",
}

func (s RouteType) String() string { return elib.Stringer(routeTypeNames, int(s)) }

type RouteProtocol uint8

const (
	RTPROT_UNSPEC   RouteProtocol = 0
	RTPROT_REDIRECT RouteProtocol = 1 /* Route installed by ICMP redirects; not used by current IPv4 */
	RTPROT_KERNEL   RouteProtocol = 2 /* Route installed by kernel */
	RTPROT_BOOT     RouteProtocol = 3 /* Route installed during boot */
	RTPROT_STATIC   RouteProtocol = 4 /* Route installed by administrator	*/

	/* Values of protocol >= RTPROT_STATIC are not interpreted by kernel;
	   they are just passed from user and back as is.
	   It will be used by hypothetical multiple routing daemons.
	   Note that protocol values should be standardized in order to
	   avoid conflicts.
	*/

	RTPROT_GATED    RouteProtocol = 8  /* Apparently, GateD */
	RTPROT_RA       RouteProtocol = 9  /* RDISC/ND router advertisements */
	RTPROT_MRT      RouteProtocol = 10 /* Merit MRT */
	RTPROT_ZEBRA    RouteProtocol = 11 /* Zebra */
	RTPROT_BIRD     RouteProtocol = 12 /* BIRD */
	RTPROT_DNROUTED RouteProtocol = 13 /* DECnet routing daemon */
	RTPROT_XORP     RouteProtocol = 14 /* XORP */
	RTPROT_NTK      RouteProtocol = 15 /* Netsukuku */
	RTPROT_DHCP     RouteProtocol = 16 /* DHCP client */
	RTPROT_MROUTED  RouteProtocol = 17 /* Multicast daemon */
	RTPROT_BABEL    RouteProtocol = 42 /* Babel daemon */
)

var routeProtocolNames = []string{
	RTPROT_UNSPEC:   "UNSPEC",
	RTPROT_REDIRECT: "REDIRECT",
	RTPROT_KERNEL:   "KERNEL",
	RTPROT_BOOT:     "BOOT",
	RTPROT_STATIC:   "STATIC",
	RTPROT_GATED:    "GATED",
	RTPROT_RA:       "RA",
	RTPROT_MRT:      "MRT",
	RTPROT_ZEBRA:    "ZEBRA",
	RTPROT_BIRD:     "BIRD",
	RTPROT_DNROUTED: "DNROUTED",
	RTPROT_XORP:     "XORP",
	RTPROT_NTK:      "NTK",
	RTPROT_DHCP:     "DHCP",
	RTPROT_MROUTED:  "MROUTED",
	RTPROT_BABEL:    "BABEL",
}

func (x RouteProtocol) String() string { return elib.Stringer(routeProtocolNames, int(x)) }

type RouteFlags uint32

const (
	RTM_F_NOTIFY   RouteFlags = 0x100 /* Notify user of route change	*/
	RTM_F_CLONED   RouteFlags = 0x200 /* This route is cloned		*/
	RTM_F_EQUALIZE RouteFlags = 0x400 /* Multipath equalizer: NI	*/
	RTM_F_PREFIX   RouteFlags = 0x800 /* Prefix addresses		*/
)

var routeFlagNames = []string{
	8:  "Notify",
	9:  "Cloned",
	10: "Multipath equalize",
	11: "Prefix",
}

func (x RouteFlags) String() string { return elib.FlagStringer(routeFlagNames, elib.Word(x)) }

type RouteAttrKind int

const (
	RTA_UNSPEC RouteAttrKind = iota
	RTA_DST
	RTA_SRC
	RTA_IIF
	RTA_OIF
	RTA_GATEWAY
	RTA_PRIORITY
	RTA_PREFSRC
	RTA_METRICS
	RTA_MULTIPATH
	RTA_PROTOINFO
	RTA_FLOW
	RTA_CACHEINFO
	RTA_SESSION
	RTA_MP_ALGO
	RTA_TABLE
	RTA_MARK
	RTA_MFC_STATS
	RTA_VIA
	RTA_NEWDST
	RTA_PREF
	RTA_ENCAP_TYPE
	RTA_ENCAP
	RTA_MAX
)

var routeAttrKindNames = []string{
	RTA_UNSPEC:     "UNSPEC",
	RTA_DST:        "DST",
	RTA_SRC:        "SRC",
	RTA_IIF:        "IIF",
	RTA_OIF:        "OIF",
	RTA_GATEWAY:    "GATEWAY",
	RTA_PRIORITY:   "PRIORITY",
	RTA_PREFSRC:    "PREFSRC",
	RTA_METRICS:    "METRICS",
	RTA_MULTIPATH:  "MULTIPATH",
	RTA_PROTOINFO:  "PROTOINFO",
	RTA_FLOW:       "FLOW",
	RTA_CACHEINFO:  "CACHEINFO",
	RTA_SESSION:    "SESSION",
	RTA_MP_ALGO:    "MP_ALGO",
	RTA_TABLE:      "TABLE",
	RTA_MARK:       "MARK",
	RTA_MFC_STATS:  "MFC_STATS",
	RTA_VIA:        "VIA",
	RTA_NEWDST:     "NEWDST",
	RTA_PREF:       "PREF",
	RTA_ENCAP_TYPE: "ENCAP_TYPE",
	RTA_ENCAP:      "ENCAP",
}

func (x RouteAttrKind) String() string {
	return elib.Stringer(routeAttrKindNames, int(x))
}

type NeighborAttrKind int

const (
	NDA_UNSPEC NeighborAttrKind = iota
	NDA_DST
	NDA_LLADDR
	NDA_CACHEINFO
	NDA_PROBES
	NDA_VLAN
	NDA_PORT
	NDA_VNI
	NDA_IFINDEX
	NDA_MASTER
	NDA_LINK_NETNSID
	NDA_MAX
)

var neighborAttrKindNames = []string{
	NDA_UNSPEC:    "UNSPEC",
	NDA_DST:       "DST",
	NDA_LLADDR:    "LLADDR",
	NDA_CACHEINFO: "CACHEINFO",
	NDA_PROBES:    "PROBES",
	NDA_VLAN:      "VLAN",
	NDA_PORT:      "PORT",
	NDA_VNI:       "VNI",
	NDA_IFINDEX:   "IFINDEX",
	NDA_MASTER:    "MASTER",
}

func (x NeighborAttrKind) String() string {
	return elib.Stringer(neighborAttrKindNames, int(x))
}

type NeighborFlags int

const (
	NTF_USE         NeighborFlags = 0
	NTF_SELF        NeighborFlags = 1
	NTF_MASTER      NeighborFlags = 2
	NTF_PROXY       NeighborFlags = 3
	NTF_EXT_LEARNED NeighborFlags = 4
	NTF_ROUTER      NeighborFlags = 7
)

var neighborFlagNames = []string{
	NTF_USE:         "USE",
	NTF_SELF:        "SELF",
	NTF_MASTER:      "MASTER",
	NTF_PROXY:       "PROXY",
	NTF_EXT_LEARNED: "LEARNED",
	NTF_ROUTER:      "ROUTER",
}

func (x NeighborFlags) String() string {
	return elib.FlagStringer(neighborFlagNames, elib.Word(x))
}

type NeighborState uint16

const (
	NUD_INCOMPLETE_BIT, NUD_INCOMPLETE NeighborState = iota, 1 << iota
	NUD_REACHABLE_BIT, NUD_REACHABLE
	NUD_STALE_BIT, NUD_STALE
	NUD_DELAY_BIT, NUD_DELAY
	NUD_PROBE_BIT, NUD_PROBE
	NUD_FAILED_BIT, NUD_FAILED
	NUD_NOARP_BIT, NUD_NOARP
	NUD_PERMANENT_BIT, NUD_PERMANENT
	NUD_NONE NeighborState = 0
)

var neighborStateNames = []string{
	NUD_INCOMPLETE_BIT: "INCOMPLETE",
	NUD_REACHABLE_BIT:  "REACHABLE",
	NUD_STALE_BIT:      "STALE",
	NUD_DELAY_BIT:      "DELAY",
	NUD_PROBE_BIT:      "PROBE",
	NUD_FAILED_BIT:     "FAILED",
	NUD_NOARP_BIT:      "NOARP",
	NUD_PERMANENT_BIT:  "PERMANENT",
}

func (x NeighborState) String() string {
	return elib.FlagStringer(neighborStateNames, elib.Word(x))
}

type AddressFamily uint8

const SizeofAddressFamily = 1

const (
	AF_UNSPEC AddressFamily = iota
	AF_UNIX
	AF_INET
	AF_AX25
	AF_IPX
	AF_APPLETALK
	AF_NETROM
	AF_BRIDGE
	AF_ATMPVC
	AF_X25
	AF_INET6
	AF_ROSE
	AF_DECnet
	AF_NETBEUI
	AF_SECURITY
	AF_KEY
	AF_NETLINK
	AF_PACKET
	AF_ASH
	AF_ECONET
	AF_ATMSVC
	AF_RDS
	AF_SNA
	AF_IRDA
	AF_PPPOX
	AF_WANPIPE
	AF_LLC
	AF_IB
	AF_MPLS
	AF_CAN
	AF_TIPC
	AF_BLUETOOTH
	AF_IUCV
	AF_RXRPC
	AF_ISDN
	AF_PHONET
	AF_IEEE802154
	AF_CAIF
	AF_ALG
	AF_NFC
	AF_VSOCK
)

var afNames = []string{
	AF_UNSPEC:     "UNSPEC",
	AF_UNIX:       "UNIX",
	AF_INET:       "INET",
	AF_AX25:       "AX25",
	AF_IPX:        "IPX",
	AF_APPLETALK:  "APPLETALK",
	AF_NETROM:     "NETROM",
	AF_BRIDGE:     "BRIDGE",
	AF_ATMPVC:     "ATMPVC",
	AF_X25:        "X25",
	AF_INET6:      "INET6",
	AF_ROSE:       "ROSE",
	AF_DECnet:     "DECNET",
	AF_NETBEUI:    "NETBEUI",
	AF_SECURITY:   "SECURITY",
	AF_KEY:        "KEY",
	AF_NETLINK:    "NETLINK",
	AF_PACKET:     "PACKET",
	AF_ASH:        "ASH",
	AF_ECONET:     "ECONET",
	AF_ATMSVC:     "ATMSVC",
	AF_RDS:        "RDS",
	AF_SNA:        "SNA",
	AF_IRDA:       "IRDA",
	AF_PPPOX:      "PPPOX",
	AF_WANPIPE:    "WANPIPE",
	AF_LLC:        "LLC",
	AF_IB:         "IB",
	AF_MPLS:       "MPLS",
	AF_CAN:        "CAN",
	AF_TIPC:       "TIPC",
	AF_BLUETOOTH:  "BLUETOOTH",
	AF_IUCV:       "IUCV",
	AF_RXRPC:      "RXRPC",
	AF_ISDN:       "ISDN",
	AF_PHONET:     "PHONET",
	AF_IEEE802154: "IEEE802154",
	AF_CAIF:       "CAIF",
	AF_ALG:        "ALG",
	AF_NFC:        "NFC",
	AF_VSOCK:      "VSOCK",
}

func (af AddressFamily) String() string {
	return elib.Stringer(afNames, int(af))
}
func (af AddressFamily) Uint() uint8 {
	return uint8(af)
}

type AddressFamilyAttrType Empty

func NewAddressFamilyAttrType() *AddressFamilyAttrType {
	return (*AddressFamilyAttrType)(pool.Empty.Get().(*Empty))
}

func (t *AddressFamilyAttrType) attrType() {}
func (t *AddressFamilyAttrType) Close() error {
	repool(t)
	return nil
}
func (t *AddressFamilyAttrType) IthString(i int) string {
	return elib.Stringer(afNames, i)
}

type AddressFamilyAddress interface {
	String([]byte) string
}

func afAddr(af AddressFamily, b []byte) Attr {
	switch {
	case af == AF_INET || len(b) == 4:
		return NewIp4AddressBytes(b)
	case af == AF_INET6 || len(b) == 16:
		return NewIp6AddressBytes(b)
	case af == AF_UNSPEC || len(b) == 6:
		return NewEthernetAddressBytes(b)
	default:
		panic(fmt.Errorf("unrecognized addr: %v", b))
	}
}

type L2IfType int

const (
	/* ARP protocol HARDWARE identifiers. */
	ARPHRD_NETROM     L2IfType = 0  /* From KA9Q: NET/ROM pseudo. */
	ARPHRD_ETHER      L2IfType = 1  /* Ethernet 10/100Mbps.  */
	ARPHRD_EETHER     L2IfType = 2  /* Experimental Ethernet.  */
	ARPHRD_AX25       L2IfType = 3  /* AX.25 Level 2.  */
	ARPHRD_PRONET     L2IfType = 4  /* PROnet token ring.  */
	ARPHRD_CHAOS      L2IfType = 5  /* Chaosnet.  */
	ARPHRD_IEEE802    L2IfType = 6  /* IEEE 802.2 Ethernet/TR/TB.  */
	ARPHRD_ARCNET     L2IfType = 7  /* ARCnet.  */
	ARPHRD_APPLETLK   L2IfType = 8  /* APPLEtalk.  */
	ARPHRD_DLCI       L2IfType = 15 /* Frame Relay DLCI.  */
	ARPHRD_ATM        L2IfType = 19 /* ATM.  */
	ARPHRD_METRICOM   L2IfType = 23 /* Metricom STRIP (new IANA id).  */
	ARPHRD_IEEE1394   L2IfType = 24 /* IEEE 1394 IPv4 - RFC 2734.  */
	ARPHRD_EUI64      L2IfType = 27 /* EUI-64.  */
	ARPHRD_INFINIBAND L2IfType = 32 /* InfiniBand.  */

	/* Dummy types for non ARP hardware */
	ARPHRD_SLIP    L2IfType = 256
	ARPHRD_CSLIP   L2IfType = 257
	ARPHRD_SLIP6   L2IfType = 258
	ARPHRD_CSLIP6  L2IfType = 259
	ARPHRD_RSRVD   L2IfType = 260 /* Notional KISS type.  */
	ARPHRD_ADAPT   L2IfType = 264
	ARPHRD_ROSE    L2IfType = 270
	ARPHRD_X25     L2IfType = 271 /* CCITT X.25.  */
	ARPHRD_HWX25   L2IfType = 272 /* Boards with X.25 in firmware.  */
	ARPHRD_PPP     L2IfType = 512
	ARPHRD_CISCO   L2IfType = 513 /* Cisco HDLC.  */
	ARPHRD_HDLC    L2IfType = ARPHRD_CISCO
	ARPHRD_LAPB    L2IfType = 516 /* LAPB.  */
	ARPHRD_DDCMP   L2IfType = 517 /* Digital's DDCMP.  */
	ARPHRD_RAWHDLC L2IfType = 518 /* Raw HDLC.  */

	ARPHRD_TUNNEL             L2IfType = 768 /* IPIP tunnel.  */
	ARPHRD_TUNNEL6            L2IfType = 769 /* IPIP6 tunnel.  */
	ARPHRD_FRAD               L2IfType = 770 /* Frame Relay Access Device.  */
	ARPHRD_SKIP               L2IfType = 771 /* SKIP vif.  */
	ARPHRD_LOOPBACK           L2IfType = 772 /* Loopback device.  */
	ARPHRD_LOCALTLK           L2IfType = 773 /* Localtalk device.  */
	ARPHRD_FDDI               L2IfType = 774 /* Fiber Distributed Data Interface. */
	ARPHRD_BIF                L2IfType = 775 /* AP1000 BIF.  */
	ARPHRD_SIT                L2IfType = 776 /* sit0 device - IPv6-in-IPv4.  */
	ARPHRD_IPDDP              L2IfType = 777 /* IP-in-DDP tunnel.  */
	ARPHRD_IPGRE              L2IfType = 778 /* GRE over IP.  */
	ARPHRD_PIMREG             L2IfType = 779 /* PIMSM register interface.  */
	ARPHRD_HIPPI              L2IfType = 780 /* High Performance Parallel I'face. */
	ARPHRD_ASH                L2IfType = 781 /* (Nexus Electronics) Ash.  */
	ARPHRD_ECONET             L2IfType = 782 /* Acorn Econet.  */
	ARPHRD_IRDA               L2IfType = 783 /* Linux-IrDA.  */
	ARPHRD_FCPP               L2IfType = 784 /* Point to point fibrechanel.  */
	ARPHRD_FCAL               L2IfType = 785 /* Fibrechanel arbitrated loop.  */
	ARPHRD_FCPL               L2IfType = 786 /* Fibrechanel public loop.  */
	ARPHRD_FCFABRIC           L2IfType = 787 /* Fibrechanel fabric.  */
	ARPHRD_IEEE802_TR         L2IfType = 800 /* Magic type ident for TR.  */
	ARPHRD_IEEE80211          L2IfType = 801 /* IEEE 802.11.  */
	ARPHRD_IEEE80211_PRISM    L2IfType = 802 /* IEEE 802.11 + Prism2 header.  */
	ARPHRD_IEEE80211_RADIOTAP L2IfType = 803 /* IEEE 802.11 + radiotap header.  */
	ARPHRD_IEEE802154         L2IfType = 804 /* IEEE 802.15.4 header.  */
	ARPHRD_IEEE802154_PHY     L2IfType = 805 /* IEEE 802.15.4 PHY header.  */
)

var l2IfTypeNames = []string{
	ARPHRD_NETROM:             "NETROM",
	ARPHRD_ETHER:              "ETHER",
	ARPHRD_EETHER:             "EETHER",
	ARPHRD_AX25:               "AX25",
	ARPHRD_PRONET:             "PRONET",
	ARPHRD_CHAOS:              "CHAOS",
	ARPHRD_IEEE802:            "IEEE802",
	ARPHRD_ARCNET:             "ARCNET",
	ARPHRD_APPLETLK:           "APPLETLK",
	ARPHRD_DLCI:               "DLCI",
	ARPHRD_ATM:                "ATM",
	ARPHRD_METRICOM:           "METRICOM",
	ARPHRD_IEEE1394:           "IEEE1394",
	ARPHRD_EUI64:              "EUI64",
	ARPHRD_INFINIBAND:         "INFINIBAND",
	ARPHRD_SLIP:               "SLIP",
	ARPHRD_CSLIP:              "CSLIP",
	ARPHRD_SLIP6:              "SLIP6",
	ARPHRD_CSLIP6:             "CSLIP6",
	ARPHRD_RSRVD:              "RSRVD",
	ARPHRD_ADAPT:              "ADAPT",
	ARPHRD_ROSE:               "ROSE",
	ARPHRD_X25:                "X25",
	ARPHRD_HWX25:              "HWX25",
	ARPHRD_PPP:                "PPP",
	ARPHRD_HDLC:               "HDLC",
	ARPHRD_LAPB:               "LAPB",
	ARPHRD_DDCMP:              "DDCMP",
	ARPHRD_RAWHDLC:            "RAWHDLC",
	ARPHRD_TUNNEL:             "TUNNEL",
	ARPHRD_TUNNEL6:            "TUNNEL6",
	ARPHRD_FRAD:               "FRAD",
	ARPHRD_SKIP:               "SKIP",
	ARPHRD_LOOPBACK:           "LOOPBACK",
	ARPHRD_LOCALTLK:           "LOCALTLK",
	ARPHRD_FDDI:               "FDDI",
	ARPHRD_BIF:                "BIF",
	ARPHRD_SIT:                "SIT",
	ARPHRD_IPDDP:              "IPDDP",
	ARPHRD_IPGRE:              "IPGRE",
	ARPHRD_PIMREG:             "PIMREG",
	ARPHRD_HIPPI:              "HIPPI",
	ARPHRD_ASH:                "ASH",
	ARPHRD_ECONET:             "ECONET",
	ARPHRD_IRDA:               "IRDA",
	ARPHRD_FCPP:               "FCPP",
	ARPHRD_FCAL:               "FCAL",
	ARPHRD_FCPL:               "FCPL",
	ARPHRD_FCFABRIC:           "FCFABRIC",
	ARPHRD_IEEE802_TR:         "IEEE802_TR",
	ARPHRD_IEEE80211:          "IEEE80211",
	ARPHRD_IEEE80211_PRISM:    "IEEE80211_PRISM",
	ARPHRD_IEEE80211_RADIOTAP: "IEEE80211_RADIOTAP",
	ARPHRD_IEEE802154:         "IEEE802154",
	ARPHRD_IEEE802154_PHY:     "IEEE802154_PHY",
}

func (x L2IfType) String() string { return elib.Stringer(l2IfTypeNames, int(x)) }

type IfInfoFlags uint32

const (
	IFF_UP IfInfoFlags = 1 << iota
	IFF_BROADCAST
	IFF_DEBUG
	IFF_LOOPBACK
	IFF_POINTOPOINT
	IFF_NOTRAILERS
	IFF_RUNNING
	IFF_NOARP
	IFF_PROMISC
	IFF_ALLMULTI
	IFF_MASTER
	IFF_SLAVE
	IFF_MULTICAST
	IFF_PORTSEL
	IFF_AUTOMEDIA
	IFF_DYNAMIC
	IFF_LOWER_UP
	IFF_DORMANT
	IFF_ECHO
)

var ifInfoFlagNames = []string{
	"Admin Up", "Broadcast", "Debug", "Loopback",
	"Point To Point", "No Trailers", "Running", "No ARP",
	"Promiscuous", "All Multicast", "Master", "Slave",
	"Multicast", "Portsel", "Automedia", "Dynamic",
	"Link Up", "Dormant", "Echo",
}

func (x IfInfoFlags) String() string { return elib.FlagStringer(ifInfoFlagNames, elib.Word(x)) }

type IfAddrFlags uint32

const (
	IFA_F_SECONDARY IfAddrFlags = 1 << iota
	IFA_F_NODAD
	IFA_F_OPTIMISTIC
	IFA_F_DADFAILED
	IFA_F_HOMEADDRESS
	IFA_F_DEPRECATED
	IFA_F_TENTATIVE
	IFA_F_PERMANENT
	IFA_F_MANAGETEMPADDR
	IFA_F_NOPREFIXROUTE
	IFA_F_MCAUTOJOIN
	IFA_F_STABLE_PRIVACY

	IFA_F_TEMPORARY = IFA_F_SECONDARY
)

var ifAddrFlagNames = []string{
	"SECONDARY", "NO_DAD", "OPTIMISTIC", "DAD_FAILED",
	"HOME_ADDRESS", "DEPRECATED", "TENTATIVE", "PERMANENT",
	"MANAGETEMPADDR", "NO_PREFIX_ROUTE", "MC_AUTO_JOIN", "STABLE_PRIVACY",
}

func (x IfAddrFlags) String() string { return elib.FlagStringer(ifAddrFlagNames, elib.Word(x)) }
func (x IfAddrFlags) Uint() uint8    { return uint8(x) }

type MessageType int

func (t MessageType) String() string {
	names := []string{
		NLMSG_NOOP:       "NLMSG_NOOP",
		NLMSG_DONE:       "NLMSG_DONE",
		NLMSG_ERROR:      "NLMSG_ERROR",
		NLMSG_OVERRUN:    "NLMSG_OVERRUN",
		RTM_NEWLINK:      "RTM_NEWLINK",
		RTM_DELLINK:      "RTM_DELLINK",
		RTM_GETLINK:      "RTM_GETLINK",
		RTM_SETLINK:      "RTM_SETLINK",
		RTM_NEWADDR:      "RTM_NEWADDR",
		RTM_DELADDR:      "RTM_DELADDR",
		RTM_GETADDR:      "RTM_GETADDR",
		RTM_NEWROUTE:     "RTM_NEWROUTE",
		RTM_DELROUTE:     "RTM_DELROUTE",
		RTM_GETROUTE:     "RTM_GETROUTE",
		RTM_NEWNEIGH:     "RTM_NEWNEIGH",
		RTM_DELNEIGH:     "RTM_DELNEIGH",
		RTM_GETNEIGH:     "RTM_GETNEIGH",
		RTM_NEWRULE:      "RTM_NEWRULE",
		RTM_DELRULE:      "RTM_DELRULE",
		RTM_GETRULE:      "RTM_GETRULE",
		RTM_NEWQDISC:     "RTM_NEWQDISC",
		RTM_DELQDISC:     "RTM_DELQDISC",
		RTM_GETQDISC:     "RTM_GETQDISC",
		RTM_NEWTCLASS:    "RTM_NEWTCLASS",
		RTM_DELTCLASS:    "RTM_DELTCLASS",
		RTM_GETTCLASS:    "RTM_GETTCLASS",
		RTM_NEWTFILTER:   "RTM_NEWTFILTER",
		RTM_DELTFILTER:   "RTM_DELTFILTER",
		RTM_GETTFILTER:   "RTM_GETTFILTER",
		RTM_NEWACTION:    "RTM_NEWACTION",
		RTM_DELACTION:    "RTM_DELACTION",
		RTM_GETACTION:    "RTM_GETACTION",
		RTM_NEWPREFIX:    "RTM_NEWPREFIX",
		RTM_GETMULTICAST: "RTM_GETMULTICAST",
		RTM_GETANYCAST:   "RTM_GETANYCAST",
		RTM_NEWNEIGHTBL:  "RTM_NEWNEIGHTBL",
		RTM_GETNEIGHTBL:  "RTM_GETNEIGHTBL",
		RTM_SETNEIGHTBL:  "RTM_SETNEIGHTBL",
		RTM_NEWNDUSEROPT: "RTM_NEWNDUSEROPT",
		RTM_NEWADDRLABEL: "RTM_NEWADDRLABEL",
		RTM_DELADDRLABEL: "RTM_DELADDRLABEL",
		RTM_GETADDRLABEL: "RTM_GETADDRLABEL",
		RTM_GETDCB:       "RTM_GETDCB",
		RTM_SETDCB:       "RTM_SETDCB",
	}
	i := int(t)
	if i < len(names) && len(names[i]) > 0 {
		return names[i]
	} else {
		panic(fmt.Errorf("unknown message type: %d", i))
	}
}

const (
	IFLA_UNSPEC IfInfoAttrKind = iota
	IFLA_ADDRESS
	IFLA_BROADCAST
	IFLA_IFNAME
	IFLA_MTU
	IFLA_LINK
	IFLA_QDISC
	IFLA_STATS
	IFLA_COST
	IFLA_PRIORITY
	IFLA_MASTER
	IFLA_WIRELESS /* Wireless Extension event - see wireless.h */
	IFLA_PROTINFO /* Protocol specific information for a link */
	IFLA_TXQLEN
	IFLA_MAP
	IFLA_WEIGHT
	IFLA_OPERSTATE
	IFLA_LINKMODE
	IFLA_LINKINFO
	IFLA_NET_NS_PID
	IFLA_IFALIAS
	IFLA_NUM_VF /* Number of VFs if device is SR-IOV PF */
	IFLA_VFINFO_LIST
	IFLA_STATS64
	IFLA_VF_PORTS
	IFLA_PORT_SELF
	IFLA_AF_SPEC
	IFLA_GROUP /* Group the device belongs to */
	IFLA_NET_NS_FD
	IFLA_EXT_MASK    /* Extended info mask VFs etc */
	IFLA_PROMISCUITY /* Promiscuity count: > 0 means acts PROMISC */
	IFLA_NUM_TX_QUEUES
	IFLA_NUM_RX_QUEUES
	IFLA_CARRIER
	IFLA_PHYS_PORT_ID
	IFLA_CARRIER_CHANGES
	IFLA_PHYS_SWITCH_ID
	IFLA_LINK_NETNSID
	IFLA_PHYS_PORT_NAME
	IFLA_PROTO_DOWN
	IFLA_GSO_MAX_SEGS
	IFLA_GSO_MAX_SIZE
	IFLA_MAX
)

const (
	IFA_UNSPEC IfAddrAttrKind = iota
	IFA_ADDRESS
	IFA_LOCAL
	IFA_LABEL
	IFA_BROADCAST
	IFA_ANYCAST
	IFA_CACHEINFO
	IFA_MULTICAST
	IFA_FLAGS
	IFA_MAX
)

var ifInfoAttrKindNames = []string{
	IFLA_UNSPEC:          "UNSPEC",
	IFLA_ADDRESS:         "ADDRESS",
	IFLA_BROADCAST:       "BROADCAST",
	IFLA_IFNAME:          "IFNAME",
	IFLA_MTU:             "MTU",
	IFLA_LINK:            "LINK",
	IFLA_QDISC:           "QDISC",
	IFLA_STATS:           "STATS",
	IFLA_COST:            "COST",
	IFLA_PRIORITY:        "PRIORITY",
	IFLA_MASTER:          "MASTER",
	IFLA_WIRELESS:        "WIRELESS",
	IFLA_PROTINFO:        "PROTINFO",
	IFLA_TXQLEN:          "TXQLEN",
	IFLA_MAP:             "MAP",
	IFLA_WEIGHT:          "WEIGHT",
	IFLA_OPERSTATE:       "OPERSTATE",
	IFLA_LINKMODE:        "LINKMODE",
	IFLA_LINKINFO:        "LINKINFO",
	IFLA_NET_NS_PID:      "NET_NS_PID",
	IFLA_IFALIAS:         "IFALIAS",
	IFLA_NUM_VF:          "NUM_VF",
	IFLA_VFINFO_LIST:     "VFINFO_LIST",
	IFLA_STATS64:         "STATS64",
	IFLA_VF_PORTS:        "VF_PORTS",
	IFLA_PORT_SELF:       "PORT_SELF",
	IFLA_AF_SPEC:         "AF_SPEC",
	IFLA_GROUP:           "GROUP",
	IFLA_NET_NS_FD:       "NET_NS_FD",
	IFLA_EXT_MASK:        "EXT_MASK",
	IFLA_PROMISCUITY:     "PROMISCUITY",
	IFLA_NUM_TX_QUEUES:   "NUM_TX_QUEUES",
	IFLA_NUM_RX_QUEUES:   "NUM_RX_QUEUES",
	IFLA_CARRIER:         "CARRIER",
	IFLA_PHYS_PORT_ID:    "PHYS_PORT_ID",
	IFLA_CARRIER_CHANGES: "CARRIER_CHANGES",
	IFLA_PHYS_SWITCH_ID:  "PHYS_SWITCH_ID",
	IFLA_LINK_NETNSID:    "LINK_NETNSID",
	IFLA_PHYS_PORT_NAME:  "PHYS_PORT_NAME",
	IFLA_PROTO_DOWN:      "PROTO_DOWN",
	IFLA_GSO_MAX_SEGS:    "GSO_MAX_SEGS",
	IFLA_GSO_MAX_SIZE:    "GSO_MAX_SIZE",
}

type IfInfoAttrKind int

func (t IfInfoAttrKind) String() string {
	return elib.Stringer(ifInfoAttrKindNames, int(t))
}

var ifAddrAttrKindNames = []string{
	IFA_UNSPEC:    "UNSPEC",
	IFA_ADDRESS:   "ADDRESS",
	IFA_LOCAL:     "LOCAL",
	IFA_LABEL:     "LABEL",
	IFA_BROADCAST: "BROADCAST",
	IFA_ANYCAST:   "ANYCAST",
	IFA_CACHEINFO: "CACHEINFO",
	IFA_MULTICAST: "MULTICAST",
	IFA_FLAGS:     "FLAGS",
}

type IfAddrAttrKind int

func (t IfAddrAttrKind) String() string {
	return elib.Stringer(ifAddrAttrKindNames, int(t))
}

const (
	IF_OPER_UNKNOWN IfOperState = iota
	IF_OPER_NOTPRESENT
	IF_OPER_DOWN
	IF_OPER_LOWERLAYERDOWN
	IF_OPER_TESTING
	IF_OPER_DORMANT
	IF_OPER_UP
)

var ifOperStates = []string{
	"unknown",
	"not present",
	"down",
	"lower layer down",
	"testing",
	"dormant",
	"up",
}

const (
	IFLA_INET_UNSPEC Ip4IfAttrKind = iota
	IFLA_INET_CONF
)

var ip4IfAttrTypeNames = []string{
	IFLA_INET_UNSPEC: "UNSPEC",
	IFLA_INET_CONF:   "CONF",
}

type Ip4IfAttrKind int
type Ip4IfAttrType Empty

func NewIp4IfAttrType() *Ip4IfAttrType {
	return (*Ip4IfAttrType)(pool.Empty.Get().(*Empty))
}

func (t Ip4IfAttrKind) String() string {
	return elib.Stringer(ip4IfAttrTypeNames, int(t))
}

func (t *Ip4IfAttrType) attrType() {}
func (t *Ip4IfAttrType) Close() error {
	repool(t)
	return nil
}
func (t *Ip4IfAttrType) IthString(i int) string {
	return elib.Stringer(ip4IfAttrTypeNames, i)
}

const (
	IFLA_INET6_UNSPEC        Ip6IfAttrKind = iota
	IFLA_INET6_FLAGS                       /* link flags			*/
	IFLA_INET6_CONF                        /* sysctl parameters		*/
	IFLA_INET6_STATS                       /* statistics			*/
	IFLA_INET6_MCAST                       /* MC things. What of them?	*/
	IFLA_INET6_CACHEINFO                   /* time values and max reasm size */
	IFLA_INET6_ICMP6STATS                  /* statistics (icmpv6)		*/
	IFLA_INET6_TOKEN                       /* device token			*/
	IFLA_INET6_ADDR_GEN_MODE               /* implicit address generator mode */
)

var ip6IfAttrTypeNames = []string{
	IFLA_INET6_UNSPEC:        "UNSPEC",
	IFLA_INET6_FLAGS:         "FLAGS",
	IFLA_INET6_CONF:          "CONF",
	IFLA_INET6_STATS:         "STATS",
	IFLA_INET6_MCAST:         "MULTICAST",
	IFLA_INET6_CACHEINFO:     "CACHEINFO",
	IFLA_INET6_ICMP6STATS:    "ICMP6STATS",
	IFLA_INET6_TOKEN:         "TOKEN",
	IFLA_INET6_ADDR_GEN_MODE: "ADDR_GEN_MODE",
}

type Ip6IfAttrKind int
type Ip6IfAttrType Empty

func NewIp6IfAttrType() *Ip6IfAttrType {
	return (*Ip6IfAttrType)(pool.Empty.Get().(*Empty))
}

func (t Ip6IfAttrKind) String() string {
	return elib.Stringer(ip6IfAttrTypeNames, int(t))
}

func (t *Ip6IfAttrType) attrType() {}
func (t *Ip6IfAttrType) Close() error {
	repool(t)
	return nil
}
func (t *Ip6IfAttrType) IthString(i int) string {
	return elib.Stringer(ip6IfAttrTypeNames, i)
}

const (
	INET6_IF_PREFIX_ONLINK   = Ip6IfFlagsAttr(0x01)
	INET6_IF_PREFIX_AUTOCONF = Ip6IfFlagsAttr(0x02)
	INET6_IF_RA_OTHERCONF    = Ip6IfFlagsAttr(0x80)
	INET6_IF_RA_MANAGED      = Ip6IfFlagsAttr(0x40)
	INET6_IF_RA_RCVD         = Ip6IfFlagsAttr(0x20)
	INET6_IF_RS_SENT         = Ip6IfFlagsAttr(0x10)
	INET6_IF_READY           = Ip6IfFlagsAttr(0x80000000)
)

const (
	IPV4_DEVCONF_FORWARDING Ip4DevConfKind = iota + 1
	IPV4_DEVCONF_MC_FORWARDING
	IPV4_DEVCONF_PROXY_ARP
	IPV4_DEVCONF_ACCEPT_REDIRECTS
	IPV4_DEVCONF_SECURE_REDIRECTS
	IPV4_DEVCONF_SEND_REDIRECTS
	IPV4_DEVCONF_SHARED_MEDIA
	IPV4_DEVCONF_RP_FILTER
	IPV4_DEVCONF_ACCEPT_SOURCE_ROUTE
	IPV4_DEVCONF_BOOTP_RELAY
	IPV4_DEVCONF_LOG_MARTIANS
	IPV4_DEVCONF_TAG
	IPV4_DEVCONF_ARPFILTER
	IPV4_DEVCONF_MEDIUM_ID
	IPV4_DEVCONF_NOXFRM
	IPV4_DEVCONF_NOPOLICY
	IPV4_DEVCONF_FORCE_IGMP_VERSION
	IPV4_DEVCONF_ARP_ANNOUNCE
	IPV4_DEVCONF_ARP_IGNORE
	IPV4_DEVCONF_PROMOTE_SECONDARIES
	IPV4_DEVCONF_ARP_ACCEPT
	IPV4_DEVCONF_ARP_NOTIFY
	IPV4_DEVCONF_ACCEPT_LOCAL
	IPV4_DEVCONF_SRC_VMARK
	IPV4_DEVCONF_PROXY_ARP_PVLAN
	IPV4_DEVCONF_ROUTE_LOCALNET
	IPV4_DEVCONF_IGMPV2_UNSOLICITED_REPORT_INTERVAL
	IPV4_DEVCONF_IGMPV3_UNSOLICITED_REPORT_INTERVAL
	IPV4_DEVCONF_IGNORE_ROUTES_WITH_LINKDOWN
	IPV4_DEVCONF_MAX
)

var ip4DevConfKindNames = []string{
	IPV4_DEVCONF_FORWARDING:                         "Forwarding",
	IPV4_DEVCONF_MC_FORWARDING:                      "Multicast Forwarding",
	IPV4_DEVCONF_PROXY_ARP:                          "Proxy ARP",
	IPV4_DEVCONF_ACCEPT_REDIRECTS:                   "Accept Redirects",
	IPV4_DEVCONF_SECURE_REDIRECTS:                   "Secure Redirects",
	IPV4_DEVCONF_SEND_REDIRECTS:                     "Send Redirects",
	IPV4_DEVCONF_SHARED_MEDIA:                       "Shared Media",
	IPV4_DEVCONF_RP_FILTER:                          "Rp Filter",
	IPV4_DEVCONF_ACCEPT_SOURCE_ROUTE:                "Accept Source Route",
	IPV4_DEVCONF_BOOTP_RELAY:                        "BOOTP Relay",
	IPV4_DEVCONF_LOG_MARTIANS:                       "Log Martians",
	IPV4_DEVCONF_TAG:                                "Tag",
	IPV4_DEVCONF_ARPFILTER:                          "ARP Filter",
	IPV4_DEVCONF_MEDIUM_ID:                          "Medium ID",
	IPV4_DEVCONF_NOXFRM:                             "No Xfrm",
	IPV4_DEVCONF_NOPOLICY:                           "No Policy",
	IPV4_DEVCONF_FORCE_IGMP_VERSION:                 "Force IGMP Version",
	IPV4_DEVCONF_ARP_ANNOUNCE:                       "ARP Announce",
	IPV4_DEVCONF_ARP_IGNORE:                         "ARP Ignore",
	IPV4_DEVCONF_PROMOTE_SECONDARIES:                "Promote Secondaries",
	IPV4_DEVCONF_ARP_ACCEPT:                         "ARP Accept",
	IPV4_DEVCONF_ARP_NOTIFY:                         "ARP Notify",
	IPV4_DEVCONF_ACCEPT_LOCAL:                       "Accept Local",
	IPV4_DEVCONF_SRC_VMARK:                          "Src Vmark",
	IPV4_DEVCONF_PROXY_ARP_PVLAN:                    "Proxy ARP Pvlan",
	IPV4_DEVCONF_ROUTE_LOCALNET:                     "Route Localnet",
	IPV4_DEVCONF_IGMPV2_UNSOLICITED_REPORT_INTERVAL: "IGMPV2 Unsolicited Report Interval",
	IPV4_DEVCONF_IGMPV3_UNSOLICITED_REPORT_INTERVAL: "IGMPV3 Unsolicited Report Interval",
	IPV4_DEVCONF_IGNORE_ROUTES_WITH_LINKDOWN:        "Ignore Routes With Linkdown",
}

type Ip4DevConfKind int
type Ip4DevConfType Empty

func (t Ip4DevConfKind) String() string {
	return elib.Stringer(ip4DevConfKindNames, int(t))
}

func (t *Ip4DevConfType) attrType() {}
func (t *Ip4DevConfType) IthString(i int) string {
	return elib.Stringer(ip4DevConfKindNames, i)
}

const (
	IPV6_DEVCONF_FORWARDING Ip6DevConfKind = iota
	IPV6_DEVCONF_HOPLIMIT
	IPV6_DEVCONF_MTU6
	IPV6_DEVCONF_ACCEPT_RA
	IPV6_DEVCONF_ACCEPT_REDIRECTS
	IPV6_DEVCONF_AUTOCONF
	IPV6_DEVCONF_DAD_TRANSMITS
	IPV6_DEVCONF_RTR_SOLICITS
	IPV6_DEVCONF_RTR_SOLICIT_INTERVAL
	IPV6_DEVCONF_RTR_SOLICIT_DELAY
	IPV6_DEVCONF_USE_TEMPADDR
	IPV6_DEVCONF_TEMP_VALID_LFT
	IPV6_DEVCONF_TEMP_PREFERED_LFT
	IPV6_DEVCONF_REGEN_MAX_RETRY
	IPV6_DEVCONF_MAX_DESYNC_FACTOR
	IPV6_DEVCONF_MAX_ADDRESSES
	IPV6_DEVCONF_FORCE_MLD_VERSION
	IPV6_DEVCONF_ACCEPT_RA_DEFRTR
	IPV6_DEVCONF_ACCEPT_RA_PINFO
	IPV6_DEVCONF_ACCEPT_RA_RTR_PREF
	IPV6_DEVCONF_RTR_PROBE_INTERVAL
	IPV6_DEVCONF_ACCEPT_RA_RT_INFO_MAX_PLEN
	IPV6_DEVCONF_PROXY_NDP
	IPV6_DEVCONF_OPTIMISTIC_DAD
	IPV6_DEVCONF_ACCEPT_SOURCE_ROUTE
	IPV6_DEVCONF_MC_FORWARDING
	IPV6_DEVCONF_DISABLE_IPV6
	IPV6_DEVCONF_ACCEPT_DAD
	IPV6_DEVCONF_FORCE_TLLAO
	IPV6_DEVCONF_NDISC_NOTIFY
	IPV6_DEVCONF_MLDV1_UNSOLICITED_REPORT_INTERVAL
	IPV6_DEVCONF_MLDV2_UNSOLICITED_REPORT_INTERVAL
	IPV6_DEVCONF_SUPPRESS_FRAG_NDISC
	IPV6_DEVCONF_ACCEPT_RA_FROM_LOCAL
	IPV6_DEVCONF_USE_OPTIMISTIC
	IPV6_DEVCONF_ACCEPT_RA_MTU
	IPV6_DEVCONF_STABLE_SECRET
	IPV6_DEVCONF_USE_OIF_ADDRS_ONLY
	IPV6_DEVCONF_ACCEPT_RA_MIN_HOP_LIMIT
	IPV6_DEVCONF_IGNORE_ROUTES_WITH_LINKDOWN
	IPV6_DEVCONF_MAX
)

var ip6DevConfKindNames = []string{
	IPV6_DEVCONF_FORWARDING:                        "Forwarding",
	IPV6_DEVCONF_HOPLIMIT:                          "Hop Limit",
	IPV6_DEVCONF_MTU6:                              "MTU",
	IPV6_DEVCONF_ACCEPT_RA:                         "Accept RA",
	IPV6_DEVCONF_ACCEPT_REDIRECTS:                  "Accept Redirects",
	IPV6_DEVCONF_AUTOCONF:                          "Autoconf",
	IPV6_DEVCONF_DAD_TRANSMITS:                     "DAD Transmits",
	IPV6_DEVCONF_RTR_SOLICITS:                      "Router Solicits",
	IPV6_DEVCONF_RTR_SOLICIT_INTERVAL:              "Router Solicit Interval",
	IPV6_DEVCONF_RTR_SOLICIT_DELAY:                 "Router Solicit Delay",
	IPV6_DEVCONF_USE_TEMPADDR:                      "Use Temp Address",
	IPV6_DEVCONF_TEMP_VALID_LFT:                    "Temp Valid Left",
	IPV6_DEVCONF_TEMP_PREFERED_LFT:                 "Temp Preferred Left",
	IPV6_DEVCONF_REGEN_MAX_RETRY:                   "Regen Max Retry",
	IPV6_DEVCONF_MAX_DESYNC_FACTOR:                 "Max Desync Factor",
	IPV6_DEVCONF_MAX_ADDRESSES:                     "Max Addresses",
	IPV6_DEVCONF_FORCE_MLD_VERSION:                 "Force MLD Version",
	IPV6_DEVCONF_ACCEPT_RA_DEFRTR:                  "Accept RA Default Router",
	IPV6_DEVCONF_ACCEPT_RA_PINFO:                   "Accept RA Pinfo",
	IPV6_DEVCONF_ACCEPT_RA_RTR_PREF:                "Accept RA Router Pref",
	IPV6_DEVCONF_RTR_PROBE_INTERVAL:                "Router Probe Interval",
	IPV6_DEVCONF_ACCEPT_RA_RT_INFO_MAX_PLEN:        "Accept Ra Rt Info Max Plen",
	IPV6_DEVCONF_PROXY_NDP:                         "Proxy Ndp",
	IPV6_DEVCONF_OPTIMISTIC_DAD:                    "Optimistic DAD",
	IPV6_DEVCONF_ACCEPT_SOURCE_ROUTE:               "Accept Source Route",
	IPV6_DEVCONF_MC_FORWARDING:                     "Multicast Forwarding",
	IPV6_DEVCONF_DISABLE_IPV6:                      "Disable IPV6",
	IPV6_DEVCONF_ACCEPT_DAD:                        "Accept Dad",
	IPV6_DEVCONF_FORCE_TLLAO:                       "Force Tllao",
	IPV6_DEVCONF_NDISC_NOTIFY:                      "Ndisc Notify",
	IPV6_DEVCONF_MLDV1_UNSOLICITED_REPORT_INTERVAL: "Mldv1 Unsolicited Report Interval",
	IPV6_DEVCONF_MLDV2_UNSOLICITED_REPORT_INTERVAL: "Mldv2 Unsolicited Report Interval",
	IPV6_DEVCONF_SUPPRESS_FRAG_NDISC:               "Suppress Frag Ndisc",
	IPV6_DEVCONF_ACCEPT_RA_FROM_LOCAL:              "Accept Ra From Local",
	IPV6_DEVCONF_USE_OPTIMISTIC:                    "Use Optimistic",
	IPV6_DEVCONF_ACCEPT_RA_MTU:                     "Accept Ra MTU",
	IPV6_DEVCONF_STABLE_SECRET:                     "Stable Secret",
	IPV6_DEVCONF_USE_OIF_ADDRS_ONLY:                "Use OIF Addrs Only",
	IPV6_DEVCONF_ACCEPT_RA_MIN_HOP_LIMIT:           "Accept RA Min Hop Limit",
	IPV6_DEVCONF_IGNORE_ROUTES_WITH_LINKDOWN:       "Ignore Routes With Linkdown",
}

type Ip6DevConfKind int
type Ip6DevConfType Empty

func (t Ip6DevConfKind) String() string {
	return elib.Stringer(ip6DevConfKindNames, int(t))
}

func (t *Ip6DevConfType) attrType() {}
func (t *Ip6DevConfType) IthString(i int) string {
	return elib.Stringer(ip6DevConfKindNames, i)
}

const (
	Rx_packets LinkStatType = iota
	Tx_packets              /* total packets transmitted	*/
	Rx_bytes                /* total bytes received 	*/
	Tx_bytes                /* total bytes transmitted	*/
	Rx_errors               /* bad packets received		*/
	Tx_errors               /* packet transmit problems	*/
	Rx_dropped              /* no space in linux buffers	*/
	Tx_dropped              /* no space available in linux	*/
	Multicast               /* multicast packets received	*/
	Collisions
	Rx_length_errors
	Rx_over_errors   /* receiver ring buff overflow	*/
	Rx_crc_errors    /* recved pkt with crc error	*/
	Rx_frame_errors  /* recv'd frame alignment error */
	Rx_fifo_errors   /* recv'r fifo overrun		*/
	Rx_missed_errors /* receiver missed packet	*/
	Tx_aborted_errors
	Tx_carrier_errors
	Tx_fifo_errors
	Tx_heartbeat_errors
	Tx_window_errors
	Rx_compressed
	Tx_compressed
	N_link_stat
)

type LinkStatType int

func (t LinkStatType) String() string {
	names := []string{
		Rx_packets:          "Rx Packets",
		Tx_packets:          "Tx Packets",
		Rx_bytes:            "Rx Bytes",
		Tx_bytes:            "Tx Bytes",
		Rx_errors:           "Rx Errors",
		Tx_errors:           "Tx Errors",
		Rx_dropped:          "Rx Drops",
		Tx_dropped:          "Tx Drops",
		Multicast:           "Rx Multicast Packets",
		Collisions:          "Collisions",
		Rx_length_errors:    "Rx Length Errors",
		Rx_over_errors:      "Rx Overrun Errors",
		Rx_crc_errors:       "Rx CRC Errors",
		Rx_frame_errors:     "Rx Frame Errors",
		Rx_fifo_errors:      "Rx Fifo Errors",
		Rx_missed_errors:    "Rx Missed Errors",
		Tx_aborted_errors:   "Tx Aborts",
		Tx_carrier_errors:   "Tx Carrier Errors",
		Tx_fifo_errors:      "Tx Fifo Errors",
		Tx_heartbeat_errors: "Tx Heartbeat Errors",
		Tx_window_errors:    "Tx Window Errors",
		Rx_compressed:       "Rx Compressed Packets",
		Tx_compressed:       "Tx Compressed Packets",
	}
	i := int(t)
	if i < len(names) && len(names[i]) > 0 {
		return names[i]
	} else {
		panic(fmt.Errorf("unknown link stat type: %d", i))
	}
}

type NetnsAttrKind int

const NETNSA_NSID_NOT_ASSIGNED = NetnsAttrKind(-1)
const (
	NETNSA_NONE NetnsAttrKind = iota
	NETNSA_NSID
	NETNSA_PID
	NETNSA_FD
	NETNSA_MAX
)

var netnsAttrKindNames = []string{
	NETNSA_NONE: "NONE",
	NETNSA_NSID: "NSID",
	NETNSA_PID:  "PID",
	NETNSA_FD:   "FD",
}

func (x NetnsAttrKind) String() string {
	return elib.Stringer(netnsAttrKindNames, int(x))
}
