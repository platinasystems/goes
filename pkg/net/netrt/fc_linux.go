// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import "syscall"

var NetstatFlagCodes = []NetstatFlagCode{
	{syscall.RTF_UP, 'U'},
	{syscall.RTF_GATEWAY, 'G'},
	{syscall.RTF_HOST, 'H'},
	{syscall.RTF_REINSTATE, 'r'},
	{syscall.RTF_DYNAMIC, 'D'},
	{syscall.RTF_MODIFIED, 'M'},
	{syscall.RTF_MSS, 'x'},
	{syscall.RTF_MTU, 'F'},
	{syscall.RTF_IRTT, 'i'},
	{syscall.RTF_REJECT, 'R'},
	{syscall.RTF_STATIC, 'S'},
	{0x800, ' '},
	{syscall.RTF_NOFORWARD, 'B'},
	{syscall.RTF_THROW, 't'},
	{syscall.RTF_NOPMTUDISC, 'F'},
	{0x8000, ' '},
	{syscall.RTF_DEFAULT, 'd'},
	{syscall.RTF_ALLONLINK, 'a'},
	{0x40000, ' '},
	{0x80000, ' '},
	{syscall.RTF_LINKRT, 'L'},
	{syscall.RTF_NONEXTHOP, 'h'},
	{0x400000, ' '},
	{0x800000, ' '},
	{syscall.RTF_CACHE, 'c'},
	{syscall.RTF_FLOW, 'f'},
	{syscall.RTF_POLICY, 'p'},
	{syscall.RTF_NAT, 'n'},
	{syscall.RTF_BROADCAST, 'b'},
	{syscall.RTF_MULTICAST, 'm'},
	{syscall.RTF_INTERFACE, 'I'},
	{syscall.RTF_LOCAL, 'l'},
}
