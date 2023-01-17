// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

package netif

/*
#include <string.h>
#include <sys/ioccom.h>
#include <sys/socket.h>
#include <sys/kern_control.h>
#include <sys/sys_domain.h>
#include <net/if.h>
#include <netinet6/in6_var.h>
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

type Ifreq C.struct_ifreq

const SizeofIfreq = C.sizeof_struct_ifreq

type CtlInfo C.struct_ctl_info

const SizeofCtlInfo = C.sizeof_struct_ctl_info

type SockaddrCtl C.struct_sockaddr_ctl

const SizeofSockaddrCtl = C.sizeof_struct_sockaddr_ctl

type In6Ifreq C.struct_in6_ifreq

const SizeofIn6Ifreq = C.sizeof_struct_in6_ifreq

const (
	SIOCGIFADDR_IN6    = C.SIOCGIFADDR_IN6
	SIOCSIFADDR_IN6    = C.SIOCSIFADDR_IN6
	SIOCDIFADDR_IN6    = C.SIOCDIFADDR_IN6
	SIOCGIFNETMASK_IN6 = C.SIOCGIFNETMASK_IN6
	SIOCSIFNETMASK_IN6 = C.SIOCSIFNETMASK_IN6
	SIOCGIFDSTADDR_IN6 = C.SIOCGIFDSTADDR_IN6
	SIOCSIFDSTADDR_IN6 = C.SIOCSIFDSTADDR_IN6
	SIOCGIFAFLAG_IN6   = C.SIOCGIFAFLAG_IN6
	SIOCGIFSTAT_IN6    = C.SIOCGIFSTAT_IN6
	SIOCGIFSTAT_ICMP6  = C.SIOCGIFSTAT_ICMP6
	SIOCGSCOPE6        = C.SIOCGSCOPE6
	SIOCSSCOPE6        = C.SIOCSSCOPE6
)
