// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netstat

const rtfmark = "" +
	"U" + // 0x00000001 RTF_UP
	"G" + // 0x0000002 RTF_GATEWAY
	"H" + // 0x0000004 RTF_HOST
	"R" + // 0x0000008 RTF_REJECT
	"D" + // 0x0000010 RTF_DYNAMIC
	"M" + // 0x0000020 RTF_MODIFIED
	"." + // 0x0000040 RTF_DONE
	"m" + // 0x0000080 RTM_MASK
	"C" + // 0x0000100 RTF_CLONING
	"M" + // 0x0000200 RTF_MULTICAST
	"L" + // 0x0000400 RTF_LLINFO
	"S" + // 0x0000800 RTF_STATIC
	"B" + // 0x0001000 RTF_BLACKHOLE
	"3" + // 0x0002000 RTF_PROTO3
	"2" + // 0x0004000 RTF_PROTO2
	"1" + // 0x0008000 RTF_PROTO1
	"W" + // 0x0010000 RTF_CLONED
	"c" + // 0x0020000 RTF_CACHED
	"p" + // 0x0040000 RTF_MPATH
	"?" + // 0x0080000 RTF_0x80000
	"m" + // 0x0100000 RTF_MPLS
	"l" + // 0x0200000 RTF_LOCAL
	"b" + // 0x0400000 RTF_BROADCAST
	"k" + // 0x0800000 RTF_CONNECTED
	"b" + // 0x1000000 RTF_BFD
	""
