// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "syscall"

const (
	RTF_UP            = syscall.RTF_UP            // 0x0000001
	RTF_GATEWAY       = syscall.RTF_GATEWAY       // 0x0000002
	RTF_HOST          = syscall.RTF_HOST          // 0x0000004
	RTF_REJECT        = syscall.RTF_REJECT        // 0x0000008
	RTF_DYNAMIC       = syscall.RTF_DYNAMIC       // 0x0000010
	RTF_MODIFIED      = syscall.RTF_MODIFIED      // 0x0000020
	RTF_DONE          = syscall.RTF_DONE          // 0x0000040
	RTM_MASK          = 0x80                      // 0x0000080
	RTF_CLONING       = syscall.RTF_CLONING       // 0x0000100
	RTF_MULTICAST     = syscall.RTF_MULTICAST     // 0x0000200
	RTF_LLINFO        = syscall.RTF_LLINFO        // 0x0000400
	RTF_STATIC        = syscall.RTF_STATIC        // 0x0000800
	RTF_BLACKHOLE     = syscall.RTF_BLACKHOLE     // 0x0001000
	RTF_PERMANENT_ARP = syscall.RTF_PERMANENT_ARP // 0x0002000
	RTF_PROTO3        = syscall.RTF_PROTO3        // 0x0002000
	RTF_ANNOUNCE      = syscall.RTF_ANNOUNCE      // 0x0004000
	RTF_PROTO2        = syscall.RTF_PROTO2        // 0x0004000
	RTF_USETRAILERS   = syscall.RTF_USETRAILERS   // 0x0008000
	RTF_PROTO1        = syscall.RTF_PROTO1        // 0x0008000
	RTF_CLONED        = syscall.RTF_CLONED        // 0x0010000
	RTF_CACHED        = syscall.RTF_CACHED        // 0x0020000
	RTF_MPATH         = syscall.RTF_MPATH         // 0x0040000
	RTF_0x80000       = 0x80000                   // 0x0080000
	RTF_MPLS          = syscall.RTF_MPLS          // 0x0100000
	RTF_LOCAL         = syscall.RTF_LOCAL         // 0x0200000
	RTF_BROADCAST     = syscall.RTF_BROADCAST     // 0x0400000
	RTF_CONNECTED     = syscall.RTF_CONNECTED     // 0x0800000
	RTF_BFD           = syscall.RTF_BFD           // 0x1000000
	RTF_PINNED        = 0
	RTF_STICKY        = 0
	RTF_XRESOLVE      = 0
)
