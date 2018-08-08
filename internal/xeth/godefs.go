// +build ignore

package xeth

// #include <linux/types.h>
// #include <asm/byteorder.h>
// #define IFNAMSIZ 16
// #define ETH_ALEN 6
// #include "linux/xeth.h"
import "C"

const (
	IFNAMSIZ                 = C.IFNAMSIZ
	ETH_ALEN                 = C.ETH_ALEN
	SizeofJumboFrame         = C.XETH_SIZEOF_JUMBO_FRAME
	SizeofMsg                = C.sizeof_struct_xeth_msg
	SizeofMsgBreak           = C.sizeof_struct_xeth_msg_break
	SizeofMsgStat            = C.sizeof_struct_xeth_msg_stat
	SizeofMsgEthtoolFlags    = C.sizeof_struct_xeth_msg_ethtool_flags
	SizeofMsgEthtoolSettings = C.sizeof_struct_xeth_msg_ethtool_settings
	SizeofMsgDumpIfinfo      = C.sizeof_struct_xeth_msg_dump_ifinfo
	SizeofMsgCarrier         = C.sizeof_struct_xeth_msg_carrier
	SizeofMsgSpeed           = C.sizeof_struct_xeth_msg_speed
	SizeofMsgIfinfo          = C.sizeof_struct_xeth_msg_ifinfo
	SizeofMsgIfa             = C.sizeof_struct_xeth_msg_ifa
	SizeofMsgIfdel           = C.sizeof_struct_xeth_msg_ifdel
	SizeofMsgIfvid           = C.sizeof_struct_xeth_msg_ifvid
	SizeofMsgFibentry        = C.sizeof_struct_xeth_msg_fibentry
	SizeofMsgNeighUpdate     = C.sizeof_struct_xeth_msg_neigh_update
	SizeofNextHop            = C.sizeof_struct_xeth_next_hop
)

type Msg C.struct_xeth_msg
type Ifmsg C.struct_xeth_ifmsg
type MsgBreak C.struct_xeth_msg_break
type MsgStat C.struct_xeth_msg_stat
type MsgEthtoolFlags C.struct_xeth_msg_ethtool_flags
type MsgEthtoolSettings C.struct_xeth_msg_ethtool_settings
type MsgCarrier C.struct_xeth_msg_carrier
type MsgSpeed C.struct_xeth_msg_speed
type MsgIfinfo C.struct_xeth_msg_ifinfo
type MsgIfa C.struct_xeth_msg_ifa
type MsgIfdel C.struct_xeth_msg_ifdel
type MsgIfvid C.struct_xeth_msg_ifvid
type MsgFibentry C.struct_xeth_msg_fibentry
type NextHop C.struct_xeth_next_hop
type MsgNeighUpdate C.struct_xeth_msg_neigh_update
