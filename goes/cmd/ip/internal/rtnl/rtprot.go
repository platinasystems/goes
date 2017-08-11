// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "syscall"

const (
	RTPROT_UNSPEC   uint8 = syscall.RTPROT_UNSPEC
	RTPROT_REDIRECT uint8 = syscall.RTPROT_REDIRECT
	RTPROT_KERNEL   uint8 = syscall.RTPROT_KERNEL
	RTPROT_BOOT     uint8 = syscall.RTPROT_BOOT
	RTPROT_STATIC   uint8 = syscall.RTPROT_STATIC
	RTPROT_GATED    uint8 = syscall.RTPROT_GATED
	RTPROT_RA       uint8 = syscall.RTPROT_RA
	RTPROT_MRT      uint8 = syscall.RTPROT_MRT
	RTPROT_ZEBRA    uint8 = syscall.RTPROT_ZEBRA
	RTPROT_BIRD     uint8 = syscall.RTPROT_BIRD
	RTPROT_DNROUTED uint8 = syscall.RTPROT_DNROUTED
	RTPROT_XORP     uint8 = syscall.RTPROT_XORP
	RTPROT_NTK      uint8 = syscall.RTPROT_NTK
	RTPROT_DHCP     uint8 = syscall.RTPROT_DHCP
	RTPROT_MROUTED  uint8 = 17
	RTPROT_BABEL    uint8 = 42
)

var RtProtByName = map[string]uint8{
	"unspec":   RTPROT_UNSPEC,
	"redirect": RTPROT_REDIRECT,
	"kernel":   RTPROT_KERNEL,
	"boot":     RTPROT_BOOT,
	"static":   RTPROT_STATIC,
	"gated":    RTPROT_GATED,
	"ra":       RTPROT_RA,
	"mrt":      RTPROT_MRT,
	"zebra":    RTPROT_ZEBRA,
	"bird":     RTPROT_BIRD,
	"dnrouted": RTPROT_DNROUTED,
	"xorp":     RTPROT_XORP,
	"ntk":      RTPROT_NTK,
	"dhcp":     RTPROT_DHCP,
	"mrouted":  RTPROT_MROUTED,
	"babel":    RTPROT_BABEL,
}

var RtProtName = map[uint8]string{
	RTPROT_UNSPEC:   "unspec",
	RTPROT_REDIRECT: "redirect",
	RTPROT_KERNEL:   "kernel",
	RTPROT_BOOT:     "boot",
	RTPROT_STATIC:   "static",
	RTPROT_GATED:    "gated",
	RTPROT_RA:       "ra",
	RTPROT_MRT:      "mrt",
	RTPROT_ZEBRA:    "zebra",
	RTPROT_BIRD:     "bird",
	RTPROT_DNROUTED: "dnrouted",
	RTPROT_XORP:     "xorp",
	RTPROT_NTK:      "ntk",
	RTPROT_DHCP:     "dhcp",
	RTPROT_MROUTED:  "mrouted",
	RTPROT_BABEL:    "babel",
}
