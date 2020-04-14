// +build ignore

package internal

/*
#include <stdint.h>
#include <linux/types.h>
#include <errno.h>
#include "internal/xeth.h"
*/
import "C"

type MsgHeader C.struct_xeth_msg_header
type MsgBreak C.struct_xeth_msg_break
type MsgCarrier C.struct_xeth_msg_carrier
type MsgChangeUpperXid C.struct_xeth_msg_change_upper_xid
type MsgEthtoolFlags C.struct_xeth_msg_ethtool_flags
type MsgEthtoolSettings C.struct_xeth_msg_ethtool_settings
type MsgEthtoolLinkModes C.struct_xeth_msg_ethtool_link_modes
type NextHop C.struct_xeth_next_hop
type MsgFibEntry C.struct_xeth_msg_fibentry
type NextHop6 C.struct_xeth_next_hop6
type MsgFib6Entry C.struct_xeth_msg_fib6entry
type MsgIfa C.struct_xeth_msg_ifa
type MsgIfa6 C.struct_xeth_msg_ifa6
type MsgIfInfo C.struct_xeth_msg_ifinfo
type MsgNeighUpdate C.struct_xeth_msg_neigh_update
type MsgSpeed C.struct_xeth_msg_speed
type MsgStat C.struct_xeth_msg_stat

const (
	MsgKindBreak                         = C.XETH_MSG_KIND_BREAK
	MsgKindLinkStat                      = C.XETH_MSG_KIND_LINK_STAT
	MsgKindEthtoolStat                   = C.XETH_MSG_KIND_ETHTOOL_STAT
	MsgKindEthtoolFlags                  = C.XETH_MSG_KIND_ETHTOOL_FLAGS
	MsgKindEthtoolSettings               = C.XETH_MSG_KIND_ETHTOOL_SETTINGS
	MsgKindEthtoolLinkModesSupported     = C.XETH_MSG_KIND_ETHTOOL_LINK_MODES_SUPPORTED
	MsgKindEthtoolLinkModesAdvertising   = C.XETH_MSG_KIND_ETHTOOL_LINK_MODES_ADVERTISING
	MsgKindEthtoolLinkModesLPAdvertising = C.XETH_MSG_KIND_ETHTOOL_LINK_MODES_LP_ADVERTISING
	MsgKindDumpIfInfo                    = C.XETH_MSG_KIND_DUMP_IFINFO
	MsgKindCarrier                       = C.XETH_MSG_KIND_CARRIER
	MsgKindSpeed                         = C.XETH_MSG_KIND_SPEED
	MsgKindIfInfo                        = C.XETH_MSG_KIND_IFINFO
	MsgKindIfa                           = C.XETH_MSG_KIND_IFA
	MsgKindIfa6                          = C.XETH_MSG_KIND_IFA6
	MsgKindDumpFibInfo                   = C.XETH_MSG_KIND_DUMP_FIBINFO
	MsgKindFibEntry                      = C.XETH_MSG_KIND_FIBENTRY
	MsgKindFib6Entry                     = C.XETH_MSG_KIND_FIB6ENTRY
	MsgKindNeighUpdate                   = C.XETH_MSG_KIND_NEIGH_UPDATE
	MsgKindChangeUpperXid                = C.XETH_MSG_KIND_CHANGE_UPPER_XID
)

const (
	SizeofMsg                 = C.sizeof_struct_xeth_msg
	SizeofMsgBreak            = C.sizeof_struct_xeth_msg_break
	SizeofMsgCarrier          = C.sizeof_struct_xeth_msg_carrier
	SizeofMsgChangeUpperXid   = C.sizeof_struct_xeth_msg_change_upper_xid
	SizeofMsgDumpFibInfo      = C.sizeof_struct_xeth_msg_dump_fibinfo
	SizeofMsgDumpIfInfo       = C.sizeof_struct_xeth_msg_dump_ifinfo
	SizeofMsgEthtoolFlags     = C.sizeof_struct_xeth_msg_ethtool_flags
	SizeofMsgEthtoolSettings  = C.sizeof_struct_xeth_msg_ethtool_settings
	SizeofMsgEthtoolLinkModes = C.sizeof_struct_xeth_msg_ethtool_link_modes
	SizeofMsgIfa              = C.sizeof_struct_xeth_msg_ifa
	SizeofMsgIfa6             = C.sizeof_struct_xeth_msg_ifa6
	SizeofMsgIfInfo           = C.sizeof_struct_xeth_msg_ifinfo
	SizeofNextHop             = C.sizeof_struct_xeth_next_hop
	SizeofNextHop6            = C.sizeof_struct_xeth_next_hop6
	SizeofMsgFibEntry         = C.sizeof_struct_xeth_msg_fibentry
	SizeofMsgFib6Entry        = C.sizeof_struct_xeth_msg_fib6entry
	SizeofMsgNeighUpdate      = C.sizeof_struct_xeth_msg_neigh_update
	SizeofMsgSpeed            = C.sizeof_struct_xeth_msg_speed
	SizeofMsgStat             = C.sizeof_struct_xeth_msg_stat
)

const MsgVersion = C.XETH_MSG_VERSION

const (
	SizeofIfName     = C.XETH_IFNAMSIZ
	SizeofEthAddr    = C.XETH_ALEN
	SizeofJumboFrame = C.XETH_SIZEOF_JUMBO_FRAME
)

const (
	NETDEV_UP = 1 + iota
	NETDEV_DOWN
	NETDEV_REBOOT
	NETDEV_CHANGE
	NETDEV_REGISTER
	NETDEV_UNREGISTER
	NETDEV_CHANGEMTU
	NETDEV_CHANGEADDR
	NETDEV_GOING_DOWN
	NETDEV_CHANGENAME
	NETDEV_FEAT_CHANGE
	NETDEV_BONDING_FAILOVER
	NETDEV_PRE_UP
	NETDEV_PRE_TYPE_CHANGE
	NETDEV_POST_TYPE_CHANGE
	NETDEV_POST_INIT
	NETDEV_UNREGISTER_FINAL
	NETDEV_RELEASE
	NETDEV_NOTIFY_PEERS
	NETDEV_JOIN
	NETDEV_CHANGEUPPER
	NETDEV_RESEND_IGMP
	NETDEV_PRECHANGEMTU
	NETDEV_CHANGEINFODATA
	NETDEV_BONDING_INFO
	NETDEV_PRECHANGEUPPER
	NETDEV_CHANGELOWERSTATE
	NETDEV_UDP_TUNNEL_PUSH_INFO
	NETDEV_CHANGE_TX_QUEUE_LEN
)

const (
	IfInfoReasonNew   = C.XETH_IFINFO_REASON_NEW
	IfInfoReasonDel   = C.XETH_IFINFO_REASON_DEL
	IfInfoReasonUp    = C.XETH_IFINFO_REASON_UP
	IfInfoReasonDown  = C.XETH_IFINFO_REASON_DOWN
	IfInfoReasonDump  = C.XETH_IFINFO_REASON_DUMP
	IfInfoReasonReg   = C.XETH_IFINFO_REASON_REG
	IfInfoReasonUnreg = C.XETH_IFINFO_REASON_UNREG
)

const (
	CarrierOff = C.XETH_CARRIER_OFF
	CarrierOn  = C.XETH_CARRIER_ON
)

const (
	IFA_ADD = NETDEV_UP
	IFA_DEL = NETDEV_DOWN
)

const (
	VlanIflaUnspec = C.XETH_VLAN_IFLA_UNSPEC
	VlanIflaVid    = C.XETH_VLAN_IFLA_VID
)
