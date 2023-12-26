// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "syscall"

const (
	RTF_UP        = syscall.RTF_UP        // 0x00001
	RTF_GATEWAY   = syscall.RTF_GATEWAY   // 0x00002
	RTF_HOST      = syscall.RTF_HOST      // 0x00004
	RTF_REJECT    = syscall.RTF_REJECT    // 0x00008
	RTF_DYNAMIC   = syscall.RTF_DYNAMIC   // 0x00010
	RTF_MODIFIED  = syscall.RTF_MODIFIED  // 0x00020
	RTF_DONE      = syscall.RTF_DONE      // 0x00040
	RTF_MASK      = syscall.RTF_MASK      // 0x00080
	RTF_CLONING   = syscall.RTF_CLONING   // 0x00100
	RTF_XRESOLVE  = syscall.RTF_XRESOLVE  // 0x00200
	RTF_LLINFO    = syscall.RTF_LLINFO    // 0x00400
	RTF_STATIC    = syscall.RTF_STATIC    // 0x00800
	RTF_BLACKHOLE = syscall.RTF_BLACKHOLE // 0x01000
	RTF_CLONED    = syscall.RTF_CLONED    // 0x02000
	RTF_PROTO2    = syscall.RTF_PROTO2    // 0x04000
	RTF_PROTO1    = syscall.RTF_PROTO1    // 0x08000
	RTF_SRC       = syscall.RTF_SRC       // 0x10000
	RTF_ANNOUNCE  = syscall.RTF_ANNOUNCE  // 0x20000
	RTF_PINNED    = 0
	RTF_STICKY    = 0
)
