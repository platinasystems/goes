// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "syscall"

var NetstatFlagCodes = []NetstatFlagCode{
	{syscall.RTF_UP, 'U'},
	{syscall.RTF_GATEWAY, 'G'},
	{syscall.RTF_HOST, 'H'},
	{syscall.RTF_REJECT, 'R'},
	{syscall.RTF_DYNAMIC, 'D'},
	{syscall.RTF_MODIFIED, 'M'},
	{syscall.RTF_DONE, '.'},
	{0x80, 'd'}, // RTF_DECLONE
	{syscall.RTF_CLONING, 'C'},
	{syscall.RTF_XRESOLVE, 'X'},
	{syscall.RTF_LLINFO, 'L'},
	{syscall.RTF_STATIC, 'S'},
	{syscall.RTF_BLACKHOLE, 'B'},
	{0x2000, '_'},
	{syscall.RTF_PROTO2, '2'},
	{syscall.RTF_PROTO1, '1'},
	{syscall.RTF_PRCLONING, 'c'},
	{syscall.RTF_WASCLONED, 'W'},
	{syscall.RTF_PROTO3, '3'},
	{0x80000, 'F'}, // RTF_FIXEDMTU
	{syscall.RTF_PINNED, 'P'},
	{syscall.RTF_LOCAL, 'l'},
	{syscall.RTF_BROADCAST, 'b'},
	{syscall.RTF_MULTICAST, 'm'},
	{syscall.RTF_IFSCOPE, 'I'},
	{syscall.RTF_CONDEMNED, '*'},
	{syscall.RTF_IFREF, 'i'},
	{syscall.RTF_PROXY, 'Y'},
	{syscall.RTF_ROUTER, 'r'},
}
