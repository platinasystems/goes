// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package net_tool

import "golang.org/x/sys/unix"

const RTF_ANNOUNCE = unix.RTF_PROTO2

func ndpFlags(f uint) (s string) {
	f &= unix.RTF_GATEWAY | RTF_ANNOUNCE
	switch {
	case f == unix.RTF_GATEWAY:
		s = "R"
	case f == RTF_ANNOUNCE:
		s = "p"
	case f == unix.RTF_GATEWAY|RTF_ANNOUNCE:
		s = "Rp"
	}
	return
}

const (
	ND6_LLINFO_NOSTATE = iota - 2
	ND6_LLINFO_WAITDELETE
	ND6_LLINFO_INCOMPLETE
	ND6_LLINFO_REACHABLE
	ND6_LLINFO_STALE
	ND6_LLINFO_DELAY
	ND6_LLINFO_PROBE
)

func ndpState(st int) string {
	s, ok := map[int]string{
		ND6_LLINFO_NOSTATE:    "N",
		ND6_LLINFO_WAITDELETE: "W",
		ND6_LLINFO_INCOMPLETE: "I",
		ND6_LLINFO_REACHABLE:  "R",
		ND6_LLINFO_STALE:      "S",
		ND6_LLINFO_DELAY:      "D",
		ND6_LLINFO_PROBE:      "P",
	}[st]
	if !ok {
		s = "?"
	}
	return s
}
