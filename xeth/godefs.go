// +build ignore

package xeth

// #define IFNAMSIZ 16
// typedef unsigned long long u64;
// typedef unsigned u32;
// typedef unsigned short u16;
// typedef unsigned char u8;
// #include "linux/xeth.h"
import "C"

const (
	SizeofJumboFrame   = C.XETH_SIZEOF_JUMBO_FRAME
	SizeofSbHdr        = C.sizeof_struct_xeth_sb_hdr
	SizeofSbSetStat    = C.sizeof_struct_xeth_sb_set_stat
	SbOpSetNetStat     = C.XETH_SBOP_SET_NET_STAT
	SbOpSetEthtoolStat = C.XETH_SBOP_SET_ETHTOOL_STAT
)

type SbHdr C.struct_xeth_sb_hdr
type SbSetStat C.struct_xeth_sb_set_stat
