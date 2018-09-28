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
	"syscall"
)

type AF uint8

func (af AF) String() string {
	s, found := map[AF]string{
		syscall.AF_ALG:        "alg",
		syscall.AF_APPLETALK:  "appletalk",
		syscall.AF_ASH:        "ash",
		syscall.AF_ATMPVC:     "atmpvc",
		syscall.AF_ATMSVC:     "atmsvc",
		syscall.AF_AX25:       "ax25",
		syscall.AF_BLUETOOTH:  "bluetooth",
		syscall.AF_BRIDGE:     "bridge",
		syscall.AF_CAIF:       "caif",
		syscall.AF_CAN:        "can",
		syscall.AF_DECnet:     "decnet",
		syscall.AF_ECONET:     "econet",
		syscall.AF_FILE:       "file",
		syscall.AF_IEEE802154: "ieee802154",
		syscall.AF_INET:       "inet",
		syscall.AF_INET6:      "inet6",
		syscall.AF_IPX:        "ipx",
		syscall.AF_IRDA:       "irda",
		syscall.AF_ISDN:       "isdn",
		syscall.AF_IUCV:       "iucv",
		syscall.AF_KEY:        "key",
		syscall.AF_LLC:        "llc",
		syscall.AF_MAX:        "max",
		syscall.AF_NETBEUI:    "netbeui",
		syscall.AF_NETLINK:    "netlink",
		syscall.AF_NETROM:     "netrom",
		syscall.AF_PACKET:     "packet",
		syscall.AF_PHONET:     "phonet",
		syscall.AF_PPPOX:      "pppox",
		syscall.AF_RDS:        "rds",
		syscall.AF_ROSE:       "rose",
		syscall.AF_RXRPC:      "rxrpc",
		syscall.AF_SECURITY:   "security",
		syscall.AF_SNA:        "sna",
		syscall.AF_TIPC:       "tipc",
		syscall.AF_UNSPEC:     "unspec",
		syscall.AF_WANPIPE:    "wanpipe",
		syscall.AF_X25:        "x25",
	}[af]
	if !found {
		s = fmt.Sprint(uint8(af))
	}
	return s
}
