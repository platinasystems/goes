// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

package tuntap

/*
#include <linux/if.h>
#include <linux/if_tun.h>
#include <string.h>
*/
import "C"

const (
	IFNAMSIZ = C.IFNAMSIZ
	IFF_TAP  = C.IFF_TAP
	IFF_TUN  = C.IFF_TUN
)

type Ifreq C.struct_ifreq
type Sockaddr C.struct_sockaddr
