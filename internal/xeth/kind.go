/* Copyright(c) 2018 Platina Systems, Inc.
 *
 * This program is free software; you can redistribute it and/or modify it
 * under the terms and conditions of the GNU General Public License,
 * version 2, as published by the Free Software Foundation.
 *
 * This program is distributed in the hope it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program; if not, write to the Free Software Foundation, Inc.,
 * 51 Franklin St - Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * The full GNU General Public License is included in this distribution in
 * the file called "COPYING".
 *
 * Contact Information:
 * sw@platina.com
 * Platina Systems, 3180 Del La Cruz Blvd, Santa Clara, CA 95054
 */
package xeth

import (
	"fmt"
	"unsafe"
)

const (
	XETH_MSG_KIND_BREAK = iota
	XETH_MSG_KIND_LINK_STAT
	XETH_MSG_KIND_ETHTOOL_STAT
	XETH_MSG_KIND_ETHTOOL_FLAGS
	XETH_MSG_KIND_ETHTOOL_SETTINGS
	XETH_MSG_KIND_DUMP_IFINFO
	XETH_MSG_KIND_CARRIER
	XETH_MSG_KIND_SPEED
	XETH_MSG_KIND_IFINFO
	XETH_MSG_KIND_IFA
	XETH_MSG_KIND_DUMP_FIBINFO
	XETH_MSG_KIND_FIBENTRY
	XETH_MSG_KIND_IFDEL
	XETH_MSG_KIND_NEIGH_UPDATE
	XETH_MSG_KIND_IFVID
)

const XETH_MSG_KIND_NOT_MSG = 0xff

type Kind int

func KindOf(b []byte) Kind {
	var kind Kind = XETH_MSG_KIND_NOT_MSG
	msg := (*Msg)(unsafe.Pointer(&b[0]))
	if len(b) >= SizeofMsg &&
		msg.Z64 == 0 &&
		msg.Z32 == 0 &&
		msg.Z16 == 0 &&
		msg.Z8 == 0 {
		kind = Kind(msg.Kind)
	}
	return kind
}

func (kind Kind) String() string {
	var kinds = []string{
		"break",
		"link-stat",
		"ethtool-stat",
		"ethtool-flags",
		"ethtool-settings",
		"dump-ifinfo",
		"carrier",
		"speed",
		"ifinfo",
		"ifa",
		"dump-fibinfo",
		"fib-entry",
		"ifdel",
		"neigh-update",
		"ifvid",
	}
	i := int(kind)
	if kind == XETH_MSG_KIND_NOT_MSG {
		return "not-message"
	} else if i < len(kinds) {
		return kinds[i]
	}
	return fmt.Sprint("@", i)
}
