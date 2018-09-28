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
	"net"
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

func ToMsg(buf []byte) *Msg {
	return (*Msg)(unsafe.Pointer(&buf[0]))
}

func KindOf(buf []byte) Kind {
	var kind Kind = XETH_MSG_KIND_NOT_MSG
	msg := ToMsg(buf)
	if len(buf) >= SizeofMsg &&
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

func (kind Kind) cache(buf []byte) {
	switch kind {
	case XETH_MSG_KIND_IFA:
		msg := ToMsgIfa(buf)
		Interface.cache(msg.Ifindex, msg)
	case XETH_MSG_KIND_IFINFO:
		msg := ToMsgIfinfo(buf)
		switch msg.Reason {
		case XETH_IFINFO_REASON_NEW:
			Interface.cache(msg.Ifindex, msg)
		case XETH_IFINFO_REASON_DEL:
			Interface.del(msg.Ifindex)
		case XETH_IFINFO_REASON_UP:
			Interface.cache(msg.Ifindex, net.Flags(msg.Flags))
		case XETH_IFINFO_REASON_DOWN:
			Interface.cache(msg.Ifindex, net.Flags(msg.Flags))
		case XETH_IFINFO_REASON_DUMP:
			Interface.cache(msg.Ifindex, msg)
		case XETH_IFINFO_REASON_REG:
			entry, found := Interface.index[msg.Ifindex]
			if found {
				entry.Netns = Netns(msg.Net)
			} else {
				Interface.cache(msg.Ifindex, msg)
			}
		case XETH_IFINFO_REASON_UNREG:
			entry, found := Interface.index[msg.Ifindex]
			if found {
				entry.Netns = DefaultNetns
			}
		}
	case XETH_MSG_KIND_ETHTOOL_FLAGS:
		msg := ToMsgEthtoolFlags(buf)
		Interface.cache(msg.Ifindex, msg)
	case XETH_MSG_KIND_ETHTOOL_SETTINGS:
		msg := ToMsgEthtoolSettings(buf)
		Interface.cache(msg.Ifindex, msg)
	}
}

func (kind Kind) validate(buf []byte) error {
	if kind == XETH_MSG_KIND_NOT_MSG {
		return fmt.Errorf("corrupt message")
	}
	n, found := map[Kind]int{
		XETH_MSG_KIND_ETHTOOL_FLAGS:    SizeofMsgEthtoolFlags,
		XETH_MSG_KIND_ETHTOOL_SETTINGS: SizeofMsgEthtoolSettings,
		XETH_MSG_KIND_IFA:              SizeofMsgIfa,
		XETH_MSG_KIND_IFINFO:           SizeofMsgIfinfo,
		XETH_MSG_KIND_NEIGH_UPDATE:     SizeofMsgNeighUpdate,
	}[kind]
	if found && n != len(buf) {
		return fmt.Errorf("mismatched %s", kind)
	}
	return nil
}

func ToMsgCarrier(buf []byte) *MsgCarrier {
	return (*MsgCarrier)(unsafe.Pointer(&buf[0]))
}

func ToMsgEthtoolFlags(buf []byte) *MsgEthtoolFlags {
	return (*MsgEthtoolFlags)(unsafe.Pointer(&buf[0]))
}

func ToMsgEthtoolSettings(buf []byte) *MsgEthtoolSettings {
	return (*MsgEthtoolSettings)(unsafe.Pointer(&buf[0]))
}

func ToMsgIfa(buf []byte) *MsgIfa {
	return (*MsgIfa)(unsafe.Pointer(&buf[0]))
}

func ToMsgIfinfo(buf []byte) *MsgIfinfo {
	return (*MsgIfinfo)(unsafe.Pointer(&buf[0]))
}

func ToMsgNeighUpdate(buf []byte) *MsgNeighUpdate {
	return (*MsgNeighUpdate)(unsafe.Pointer(&buf[0]))
}

func ToMsgSpeed(buf []byte) *MsgSpeed {
	return (*MsgSpeed)(unsafe.Pointer(&buf[0]))
}

func ToMsgStat(buf []byte) *MsgStat {
	return (*MsgStat)(unsafe.Pointer(&buf[0]))
}
