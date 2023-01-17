// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

package netif

/*
#include <asm-generic/ioctl.h>
#include <linux/if.h>
#include <linux/ipv6.h>
#include <string.h>
*/
import "C"

const IFNAMSIZ = C.IFNAMSIZ

type Ifreq C.struct_ifreq

const SizeofIfreq = C.sizeof_struct_ifreq

type In6Addr C.struct_in6_addr

const SizeofIn6Addr = C.sizeof_struct_in6_addr

type In6Ifreq C.struct_in6_ifreq

const SizeofIn6Ifreq = C.sizeof_struct_in6_ifreq
