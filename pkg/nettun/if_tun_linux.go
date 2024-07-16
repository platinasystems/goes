// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

package nettun

/*
#include <asm-generic/ioctl.h>
#include <linux/if.h>
#include <linux/if_tun.h>
#include <string.h>
*/
import "C"

const IFNAMSIZ = C.IFNAMSIZ

const (
	// ioctls
	TUNSETNOCSUM    = C.TUNSETNOCSUM
	TUNSETDEBUG     = C.TUNSETDEBUG
	TUNSETIFF       = C.TUNSETIFF
	TUNSETPERSIST   = C.TUNSETPERSIST
	TUNSETOWNER     = C.TUNSETOWNER
	TUNSETLINK      = C.TUNSETLINK
	TUNSETGROUP     = C.TUNSETGROUP
	TUNGETFEATURES  = C.TUNGETFEATURES
	TUNSETOFFLOAD   = C.TUNSETOFFLOAD
	TUNSETTXFILTER  = C.TUNSETTXFILTER
	TUNGETIFF       = C.TUNGETIFF
	TUNGETSNDBUF    = C.TUNGETSNDBUF
	TUNSETSNDBUF    = C.TUNSETSNDBUF
	TUNATTACHFILTER = C.TUNATTACHFILTER
	TUNDETACHFILTER = C.TUNDETACHFILTER
	TUNGETVNETHDRSZ = C.TUNGETVNETHDRSZ
	TUNSETVNETHDRSZ = C.TUNSETVNETHDRSZ
	TUNSETQUEUE     = C.TUNSETQUEUE
	TUNSETIFINDEX   = C.TUNSETIFINDEX
	TUNGETFILTER    = C.TUNGETFILTER
	TUNSETVNETLE    = C.TUNSETVNETLE
	TUNGETVNETLE    = C.TUNGETVNETLE
	// The TUNSETVNETBE and TUNGETVNETBE ioctls are for cross-endian
	// support on little-endian hosts. Not all kernel configurations
	// support them, but all configurations that support SET also support
	// GET.
	TUNSETVNETBE       = C.TUNSETVNETBE
	TUNGETVNETBE       = C.TUNGETVNETBE
	TUNSETSTEERINGEBPF = C.TUNSETSTEERINGEBPF
	TUNSETFILTEREBPF   = C.TUNSETFILTEREBPF
	TUNSETCARRIER      = C.TUNSETCARRIER
	TUNGETDEVNETNS     = C.TUNGETDEVNETNS
)

const (
	// TUNSETIFF ifr flags
	IFF_TAP        = C.IFF_TAP
	IFF_TUN        = C.IFF_TUN
	IFF_NAPI       = C.IFF_NAPI
	IFF_NAPI_FRAGS = C.IFF_NAPI_FRAGS
	IFF_NO_PI      = C.IFF_NO_PI
	// This flag has no real effect
	IFF_ONE_QUEUE    = C.IFF_ONE_QUEUE
	IFF_VNET_HDR     = C.IFF_VNET_HDR
	IFF_TUN_EXCL     = C.IFF_TUN_EXCL
	IFF_MULTI_QUEUE  = C.IFF_MULTI_QUEUE
	IFF_ATTACH_QUEUE = C.IFF_ATTACH_QUEUE
	IFF_DETACH_QUEUE = C.IFF_DETACH_QUEUE
	// read-only flag/
	IFF_PERSIST  = C.IFF_PERSIST
	IFF_NOFILTER = C.IFF_NOFILTER
)

const (
	// Socket options
	TUN_TX_TIMESTAMP = C.TUN_TX_TIMESTAMP
)

const (
	// Features for GSO (TUNSETOFFLOAD).
	TUN_F_CSUM    = C.TUN_F_CSUM    // You can hand me unchecksummed packets.
	TUN_F_TSO4    = C.TUN_F_TSO4    // I can handle TSO for IPv4 packets
	TUN_F_TSO6    = C.TUN_F_TSO6    // I can handle TSO for IPv6 packets
	TUN_F_TSO_ECN = C.TUN_F_TSO_ECN // I can handle TSO with ECN bits.
	TUN_F_UFO     = C.TUN_F_UFO     // I can handle UFO packets
)

const (
	// Protocol info prepended to the packets (when IFF_NO_PI is not set)
	TUN_PKT_STRIP = C.TUN_PKT_STRIP
)

type TunPI C.struct_tun_pi
type TunFilter C.struct_tun_filter
type Ifreq C.struct_ifreq
type Sockaddr C.struct_sockaddr
