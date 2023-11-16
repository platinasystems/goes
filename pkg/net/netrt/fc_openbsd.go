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
	{syscall.RTF_MASK, 'n'},
	{syscall.RTF_CLONING, 'C'},
	{syscall.RTF_XRESOLVE, 'X'},
	{syscall.RTF_LLINFO, 'L'},
	{syscall.RTF_STATIC, 'S'},
	{syscall.RTF_BLACKHOLE, 'B'},
	{syscall.RTF_PERMANENT_ARP, 'A'},
	{syscall.RTF_PROTO3, '3'},
	{syscall.RTF_ANNOUNCE, 'a'},
	{syscall.RTF_PROTO2, '2'},
	{syscall.RTF_PROTO1, '1'},
	{syscall.RTF_USETRAILERS, '1'},
	{syscall.RTF_CLONED, 'W'},
	{syscall.RTF_SOURCE, 's'},
	{syscall.RTF_MPATH, 'P'},
	{0x80000, '_'},
	{syscall.RTF_MPLS, 'm'},
	{syscall.RTF_TUNNEL, 't'},
}
