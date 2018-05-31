// +build ignore

package xeth

// #include <linux/types.h>
// #include <asm/byteorder.h>
// #define IFNAMSIZ 16
// #include "linux/xeth.h"
import "C"

const (
	IFNAMSIZ                 = C.IFNAMSIZ
	SizeofJumboFrame         = C.XETH_SIZEOF_JUMBO_FRAME
	SizeofMsg                = C.sizeof_struct_xeth_msg
	SizeofMsgBreak           = C.sizeof_struct_xeth_msg_break
	SizeofMsgStat            = C.sizeof_struct_xeth_msg_stat
	SizeofMsgEthtoolFlags    = C.sizeof_struct_xeth_msg_ethtool_flags
	SizeofMsgEthtoolSettings = C.sizeof_struct_xeth_msg_ethtool_settings
	SizeofMsgDumpIfinfo      = C.sizeof_struct_xeth_msg_dump_ifinfo
	SizeofMsgCarrier         = C.sizeof_struct_xeth_msg_carrier
	SizeofMsgSpeed           = C.sizeof_struct_xeth_msg_speed
	SizeofMsgIfindex         = C.sizeof_struct_xeth_msg_ifindex
	SizeofMsgIfa             = C.sizeof_struct_xeth_msg_ifa
)

type Msg C.struct_xeth_msg
type Ifmsg C.struct_xeth_ifmsg
type MsgBreak C.struct_xeth_msg_break
type MsgStat C.struct_xeth_msg_stat
type MsgEthtoolFlags C.struct_xeth_msg_ethtool_flags
type MsgEthtoolSettings C.struct_xeth_msg_ethtool_settings
type MsgCarrier C.struct_xeth_msg_carrier
type MsgSpeed C.struct_xeth_msg_speed
type MsgIfindex C.struct_xeth_msg_ifindex
type MsgIfa C.struct_xeth_msg_ifa
