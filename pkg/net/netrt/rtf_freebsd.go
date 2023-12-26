// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "syscall"

const (
	RTF_UP            = syscall.RTF_UP            // 0x00000001
	RTF_GATEWAY       = syscall.RTF_GATEWAY       // 0x00000002
	RTF_HOST          = syscall.RTF_HOST          // 0x00000004
	RTF_REJECT        = syscall.RTF_REJECT        // 0x00000008
	RTF_DYNAMIC       = syscall.RTF_DYNAMIC       // 0x00000010
	RTF_MODIFIED      = syscall.RTF_MODIFIED      // 0x00000020
	RTF_DONE          = syscall.RTF_DONE          // 0x00000040
	RTF_MASK          = 0x80                      // 0x00000080
	RTF_CLONING       = 0x100                     // 0x00000100
	RTF_XRESOLVE      = syscall.RTF_XRESOLVE      // 0x00000200
	RTF_LLINFO        = syscall.RTF_LLINFO        // 0x00000400
	RTF_STATIC        = syscall.RTF_STATIC        // 0x00000800
	RTF_BLACKHOLE     = syscall.RTF_BLACKHOLE     // 0x00001000
	RTF_ANNOUNCE      = 0x2000                    // 0x00002000
	RTF_PROTO2        = syscall.RTF_PROTO2        // 0x00004000
	RTF_PROTO1        = syscall.RTF_PROTO1        // 0x00008000
	RTF_PRCLONING     = syscall.RTF_PRCLONING     // 0x00010000
	RTF_WASCLONED     = 0x20000                   // 0x00020000
	RTF_PROTO3        = syscall.RTF_PROTO3        // 0x00040000
	RTF_FIXEDMTU      = 0x80000                   // 0x00080000
	RTF_PINNED        = syscall.RTF_PINNED        // 0x00100000
	RTF_LOCAL         = syscall.RTF_LOCAL         // 0x00200000
	RTF_BROADCAST     = syscall.RTF_BROADCAST     // 0x00400000
	RTF_MULTICAST     = syscall.RTF_MULTICAST     // 0x00800000
	RTF_IFSCOPE       = 0x1000000                 // 0x01000000
	RTF_CONDEMNED     = 0x2000000                 // 0x02000000
	RTF_IFREF         = 0x04000000                // 0x04000000
	RTF_PROXY         = 0x08000000                // 0x08000000
	RTF_STICKY        = syscall.RTF_STICKY        // 0x10000000
	RTF_0x20000000    = 0x20000000                // 0x20000000
	RTF_RNH_LOCKED    = syscall.RTF_RNH_LOCKED    // 0x40000000
	RTF_GWFLAG_COMPAT = syscall.RTF_GWFLAG_COMPAT // 0x80000000
)
