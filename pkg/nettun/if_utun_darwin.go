// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

package nettun

/*
#include <string.h>
#include <sys/ioccom.h>
#include <sys/socket.h>
#include <sys/kern_control.h>
#include <sys/sys_domain.h>
#include <net/if.h>
#include <net/if_utun.h>
*/
import "C"

const (
	IFNAMSIZ         = C.IFNAMSIZ
	AF_SYS_CONTROL   = C.AF_SYS_CONTROL
	AF_SYSTEM        = C.AF_SYSTEM
	PF_SYSTEM        = C.AF_SYSTEM
	SYSPROTO_CONTROL = C.SYSPROTO_CONTROL
	CTLIOCGINFO      = C.CTLIOCGINFO
)

type CtlInfo C.struct_ctl_info

const SizeofCtlInfo = C.sizeof_struct_ctl_info

type SockaddrCtl C.struct_sockaddr_ctl

const SizeofSockaddrCtl = C.sizeof_struct_sockaddr_ctl

// Name registered by the utun kernel control
const UTUN_CONTROL_NAME = C.UTUN_CONTROL_NAME

// Socket option names to manage utun
const (
	UTUN_OPT_FLAGS  = C.UTUN_OPT_FLAGS
	UTUN_OPT_IFNAME = C.UTUN_OPT_IFNAME
	// get|set (type int)
	UTUN_OPT_EXT_IFDATA_STATS = C.UTUN_OPT_EXT_IFDATA_STATS
	// set to increment stat counters (type struct utun_stats_param)
	UTUN_OPT_INC_IFDATA_STATS_IN = C.UTUN_OPT_INC_IFDATA_STATS_IN
	// set to increment stat counters (type struct utun_stats_param)
	UTUN_OPT_INC_IFDATA_STATS_OUT = C.UTUN_OPT_INC_IFDATA_STATS_OUT

	// set the delegate interface (char[])
	UTUN_OPT_SET_DELEGATE_INTERFACE = C.UTUN_OPT_SET_DELEGATE_INTERFACE
	// the number of packets that can be waiting to be read
	// from the control socket at a time
	UTUN_OPT_MAX_PENDING_PACKETS = C.UTUN_OPT_MAX_PENDING_PACKETS
	UTUN_OPT_ENABLE_CHANNEL      = C.UTUN_OPT_ENABLE_CHANNEL
	UTUN_OPT_GET_CHANNEL_UUID    = C.UTUN_OPT_GET_CHANNEL_UUID
	UTUN_OPT_ENABLE_FLOWSWITCH   = C.UTUN_OPT_ENABLE_FLOWSWITCH

	// Must be set before connecting
	UTUN_OPT_ENABLE_NETIF = C.UTUN_OPT_ENABLE_NETIF
	// Must be set before connecting
	UTUN_OPT_SLOT_SIZE = C.UTUN_OPT_SLOT_SIZE
	// Must be set before connecting
	UTUN_OPT_NETIF_RING_SIZE = C.UTUN_OPT_NETIF_RING_SIZE
	// Must be set before connecting
	UTUN_OPT_TX_FSW_RING_SIZE = C.UTUN_OPT_TX_FSW_RING_SIZE
	// Must be set before connecting
	UTUN_OPT_RX_FSW_RING_SIZE = C.UTUN_OPT_RX_FSW_RING_SIZE
	// Must be set before connecting
	UTUN_OPT_KPIPE_TX_RING_SIZE = C.UTUN_OPT_KPIPE_TX_RING_SIZE
	// Must be set before connecting
	UTUN_OPT_KPIPE_RX_RING_SIZE = C.UTUN_OPT_KPIPE_RX_RING_SIZE
	// Must be set before connecting
	UTUN_OPT_ATTACH_FLOWSWITCH = C.UTUN_OPT_ATTACH_FLOWSWITCH
)

// Flags for by UTUN_OPT_FLAGS
const (
	UTUN_FLAGS_NO_OUTPUT        = C.UTUN_FLAGS_NO_OUTPUT
	UTUN_FLAGS_NO_INPUT         = C.UTUN_FLAGS_NO_INPUT
	UTUN_FLAGS_ENABLE_PROC_UUID = C.UTUN_FLAGS_ENABLE_PROC_UUID
)

type UtunStatsParam C.struct_utun_stats_param

const SizeofUtunStatsParam = C.sizeof_struct_utun_stats_param
