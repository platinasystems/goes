// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "syscall"

const (
	RTF_UP         = syscall.RTF_UP         // 0x00000001
	RTF_GATEWAY    = syscall.RTF_GATEWAY    // 0x00000002
	RTF_HOST       = syscall.RTF_HOST       // 0x00000004
	RTF_REINSTATE  = syscall.RTF_REINSTATE  // 0x00000008
	RTF_DYNAMIC    = syscall.RTF_DYNAMIC    // 0x00000010
	RTF_MODIFIED   = syscall.RTF_MODIFIED   // 0x00000020
	RTF_MSS        = syscall.RTF_MSS        // 0x00000040
	RTF_MTU        = syscall.RTF_MTU        // 0x00000040
	RTF_WINDOW     = syscall.RTF_WINDOW     // 0x00000080
	RTF_IRTT       = syscall.RTF_IRTT       // 0x00000100
	RTF_REJECT     = syscall.RTF_REJECT     // 0x00000200
	RTF_STATIC     = syscall.RTF_STATIC     // 0x00000400
	RTF_XRESOLVE   = syscall.RTF_XRESOLVE   // 0x00000800
	RTF_NOFORWARD  = syscall.RTF_NOFORWARD  // 0x00001000
	RTF_THROW      = syscall.RTF_THROW      // 0x00002000
	RTF_NOPMTUDISC = syscall.RTF_NOPMTUDISC // 0x00004000
	RTF_0x8000     = 0x8000                 // 0x00008000
	RTF_DEFAULT    = syscall.RTF_DEFAULT    // 0x00010000
	RTF_ALLONLINK  = syscall.RTF_ALLONLINK  // 0x00020000
	RTF_ADDRCONF   = syscall.RTF_ADDRCONF   // 0x00040000
	RTF_0x80000    = 0x80000                // 0x00080000
	RTF_LINKRT     = syscall.RTF_LINKRT     // 0x00100000
	RTF_NONEXTHOP  = syscall.RTF_NONEXTHOP  // 0x00200000
	RTF_0x400000   = 0x400000               // 0x00400000
	RTF_0x800000   = 0x800000               // 0x00800000
	RTF_CACHE      = syscall.RTF_CACHE      // 0x01000000
	RTF_FLOW       = syscall.RTF_FLOW       // 0x02000000
	RTF_POLICY     = syscall.RTF_POLICY     // 0x04000000
	RTF_NAT        = syscall.RTF_NAT        // 0x08000000
	RTF_BROADCAST  = syscall.RTF_BROADCAST  // 0x10000000
	RTF_MULTICAST  = syscall.RTF_MULTICAST  // 0x20000000
	RTF_INTERFACE  = syscall.RTF_INTERFACE  // 0x40000000
	RTF_LOCAL      = syscall.RTF_LOCAL      // 0x80000000
)

const (
	// unsupported
	RTF_ANNOUNCE  = 0
	RTF_BLACKHOLE = 0
	RTF_PROTO1    = 0
	RTF_PROTO2    = 0
	RTF_STICKY    = 0
)
