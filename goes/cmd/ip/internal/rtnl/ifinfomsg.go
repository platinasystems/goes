// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"syscall"
	"unsafe"
)

const SizeofIfInfoMsg = syscall.SizeofIfInfomsg

type IfInfoMsg syscall.IfInfomsg

func IfInfoMsgPtr(b []byte) *IfInfoMsg {
	if len(b) < SizeofHdr+SizeofIfInfoMsg {
		return nil
	}
	return (*IfInfoMsg)(unsafe.Pointer(&b[SizeofHdr]))
}

func (msg IfInfoMsg) Read(b []byte) (int, error) {
	*(*IfInfoMsg)(unsafe.Pointer(&b[0])) = msg
	return SizeofIfInfoMsg, nil
}

const (
	IFF_UP uint32 = 1 << iota
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

const (
	IFLA_UNSPEC uint16 = iota
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
	IFLA_WIRELESS
	IFLA_PROTINFO
	IFLA_TXQLEN
	IFLA_MAP
	IFLA_WEIGHT
	IFLA_OPERSTATE
	IFLA_LINKMODE
	IFLA_LINKINFO
	IFLA_NET_NS_PID
	IFLA_IFALIAS
	IFLA_NUM_VF
	IFLA_VFINFO_LIST
	IFLA_STATS64
	IFLA_VF_PORTS
	IFLA_PORT_SELF
	IFLA_AF_SPEC
	IFLA_GROUP
	IFLA_NET_NS_FD
	IFLA_EXT_MASK
	IFLA_PROMISCUITY
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
	N_IFLA
)

const IFLA_MAX = N_IFLA - 1

type Ifla [N_IFLA][]byte

func (ifla *Ifla) Write(b []byte) (int, error) {
	i := NLMSG.Align(SizeofHdr + SizeofIfInfoMsg)
	if i >= len(b) {
		IndexAttrByType(ifla[:], Empty)
		return 0, nil
	}
	IndexAttrByType(ifla[:], b[i:])
	return len(b) - i, nil
}

const (
	IFLA_INFO_UNSPEC uint16 = iota
	IFLA_INFO_KIND
	IFLA_INFO_DATA
	IFLA_INFO_XSTATS
	IFLA_INFO_SLAVE_KIND
	IFLA_INFO_SLAVE_DATA
	N_IFLA_INFO
)

const IFLA_INFO_MAX = N_IFLA_INFO - 1

const (
	IF_LINK_MODE_DEFAULT uint8 = iota
	IF_LINK_MODE_DORMANT
)

var IfLinkModeByName = map[string]uint8{
	"default": IF_LINK_MODE_DEFAULT,
	"dormant": IF_LINK_MODE_DORMANT,
}

var IfLinkModeName = map[uint8]string{
	IF_LINK_MODE_DEFAULT: "default",
	IF_LINK_MODE_DORMANT: "dormant",
}

const (
	IF_OPER_UNKNOWN uint8 = iota
	IF_OPER_NOTPRESENT
	IF_OPER_DOWN
	IF_OPER_LOWERLAYERDOWN
	IF_OPER_TESTING
	IF_OPER_DORMANT
	IF_OPER_UP
)

var IfOperByName = map[string]uint8{
	"unknown":     IF_OPER_UNKNOWN,
	"not-present": IF_OPER_NOTPRESENT,
	"down":        IF_OPER_DOWN,
	"lower-down":  IF_OPER_LOWERLAYERDOWN,
	"testing":     IF_OPER_TESTING,
	"dormant":     IF_OPER_DORMANT,
	"up":          IF_OPER_UP,
}

var IfOperName = map[uint8]string{
	IF_OPER_UNKNOWN:        "unknown",
	IF_OPER_NOTPRESENT:     "not-present",
	IF_OPER_DOWN:           "down",
	IF_OPER_LOWERLAYERDOWN: "lower-down",
	IF_OPER_TESTING:        "testing",
	IF_OPER_DORMANT:        "dormant",
	IF_OPER_UP:             "up",
}

const (
	Rx_packets = iota
	Tx_packets /* total packets transmitted	*/
	Rx_bytes   /* total bytes received 	*/
	Tx_bytes   /* total bytes transmitted	*/
	Rx_errors  /* bad packets received		*/
	Tx_errors  /* packet transmit problems	*/
	Rx_dropped /* no space in linux buffers	*/
	Tx_dropped /* no space available in linux	*/
	Multicast  /* multicast packets received	*/
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

const SizeofIfStats = N_link_stat * 4
const SizeofIfStats64 = N_link_stat * 8

func IfStatsAttr(b []byte) *IfStats {
	return (*IfStats)(unsafe.Pointer(&b[0]))
}

func IfStats64Attr(b []byte) *IfStats64 {
	return (*IfStats64)(unsafe.Pointer(&b[0]))
}

type IfStats [N_link_stat]uint32
type IfStats64 [N_link_stat]uint64

const (
	IFLA_INET6_UNSPEC uint16 = iota
	IFLA_INET6_FLAGS
	IFLA_INET6_CONF
	IFLA_INET6_STATS
	IFLA_INET6_MCAST
	IFLA_INET6_CACHEINFO
	IFLA_INET6_ICMP6STATS
	IFLA_INET6_TOKEN
	IFLA_INET6_ADDR_GEN_MODE
	N_IFLA_INET6
)

const IFLA_INET6_MAX = N_IFLA_INET6 - 1

const (
	IN6_ADDR_GEN_MODE_EUI64 uint8 = iota
	IN6_ADDR_GEN_MODE_NONE
	IN6_ADDR_GEN_MODE_STABLE_PRIVACY
	IN6_ADDR_GEN_MODE_RANDOM
)

var In6AddrGenModeByName = map[string]uint8{
	"eui64":  IN6_ADDR_GEN_MODE_EUI64,
	"none":   IN6_ADDR_GEN_MODE_NONE,
	"stable": IN6_ADDR_GEN_MODE_STABLE_PRIVACY,
	"random": IN6_ADDR_GEN_MODE_RANDOM,
}
