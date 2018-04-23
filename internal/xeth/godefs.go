// +build ignore

package xeth

// #define IFNAMSIZ 16
// typedef int bool;
// typedef unsigned long long u64;
// typedef unsigned u32;
// typedef unsigned short u16;
// typedef unsigned char u8;
// struct ethtool_link_ksettings {};
// #include "linux/xeth.h"
import "C"

const (
	IFNAMSIZ              = C.IFNAMSIZ
	SizeofJumboFrame      = C.XETH_SIZEOF_JUMBO_FRAME
	SizeofMsgHdr          = C.sizeof_struct_xeth_msg_hdr
	SizeofBreakMsg        = C.sizeof_struct_xeth_break_msg
	SizeofStatMsg         = C.sizeof_struct_xeth_stat_msg
	SizeofEthtoolFlagsMsg = C.sizeof_struct_xeth_ethtool_flags_msg
	SizeofEthtoolDumpMsg  = C.sizeof_struct_xeth_ethtool_dump_msg
)

type Hdr C.struct_xeth_msg_hdr
type Stat C.struct_xeth_stat
