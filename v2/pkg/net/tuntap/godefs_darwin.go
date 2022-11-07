// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

package tuntap

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
	UTUN_CONTROL_NAME = C.UTUN_CONTROL_NAME
	UTUN_OPT_IFNAME   = C.UTUN_OPT_IFNAME
	IFNAMSIZ          = C.IFNAMSIZ
	AF_SYS_CONTROL    = C.AF_SYS_CONTROL
	AF_SYSTEM         = C.AF_SYSTEM
	PF_SYSTEM         = C.AF_SYSTEM
	SYSPROTO_CONTROL  = C.SYSPROTO_CONTROL
	CTLIOCGINFO       = C.CTLIOCGINFO
)

type CtlInfo C.struct_ctl_info

const SizeofCtlInfo = C.sizeof_struct_ctl_info

type SockaddrCtl C.struct_sockaddr_ctl

const SizeofSockaddrCtl = C.sizeof_struct_sockaddr_ctl
