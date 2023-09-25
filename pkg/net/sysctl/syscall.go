// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd
// +build darwin dragonfly freebsd netbsd openbsd

package sysctl

import (
	"syscall"
	_ "unsafe"
)

//go:linkname sysctl syscall.sysctl
func sysctl(mib []int32, old *byte, oldlen *uintptr, new *byte, newlen uintptr) error

// syscall aliases

type IfMsghdr = syscall.IfMsghdr
type IfData = syscall.IfData
type IfaMsghdr = syscall.IfaMsghdr
type IfmaMsghdr = syscall.IfmaMsghdr
type IfmaMsghdr2 = syscall.IfmaMsghdr2
type RtMsghdr = syscall.RtMsghdr
type RtMetrics = syscall.RtMetrics
type SockaddrIn = syscall.RawSockaddrInet4
type SockaddrIn6 = syscall.RawSockaddrInet6

// Replaced syscall.RawSockaddrDatalink b/c it appends,
//	Data   [12]int8
// That is insufficient for name, address and possible selector.
type SockaddrDatalink struct {
	Len    uint8
	Family uint8
	Index  uint16
	Type   uint8
	Nlen   uint8
	Alen   uint8
	Slen   uint8
}

const IFNAMSIZ = syscall.IFNAMSIZ

const (
	AF_UNSPEC = syscall.AF_UNSPEC
	AF_INET   = syscall.AF_INET
	AF_INET6  = syscall.AF_INET6
	AF_LINK   = syscall.AF_LINK
	AF_ROUTE  = syscall.AF_ROUTE
)

const (
	EINVAL = syscall.EINVAL
	ENOMEM = syscall.ENOMEM
)

const (
	CTL_NET     = syscall.CTL_NET
	CTL_MAXNAME = syscall.CTL_MAXNAME
)

const (
	_ = iota
	NET_RT_DUMP
	NET_RT_FLAGS
	NET_RT_IFLIST
	NET_RT_STAT
	NET_RT_TRASH
	NET_RT_IFLIST2
	NET_RT_DUMP2
)

const NET_RT_MAXID = syscall.NET_RT_MAXID

const (
	_ = iota
	RTM_ADD
	RTM_DELETE
	RTM_CHANGE
	RTM_GET
	RTM_VERSION
	RTM_REDIRECT
	RTM_MISS
	RTM_LOCK
	RTM_OLDADD
	RTM_OLDDEL
	RTM_RESOLVE
	RTM_NEWADDR
	RTM_DELADDR
	RTM_IFINFO
	RTM_NEWMADDR
	RTM_DELMADDR
	RTM_IFINFO2
	RTM_NEWMADDR2
	RTM_GET2
)

const (
	RTM_LOSING  = RTM_VERSION
	RTM_RTTUNIT = syscall.RTM_RTTUNIT
)

const (
	RTAX_DST = iota
	RTAX_GATEWAY
	RTAX_NETMASK
	RTAX_GENMASK
	RTAX_IFP
	RTAX_IFA
	RTAX_AUTHOR
	RTAX_BRD
	RTAX_MAX
)
