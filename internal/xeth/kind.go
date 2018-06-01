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

import "fmt"

type Kind uint8

const XETH_MSG_KIND_INVALID = Kind(0xff)
const (
	XETH_MSG_KIND_BREAK Kind = iota
	XETH_MSG_KIND_LINK_STAT
	XETH_MSG_KIND_ETHTOOL_STAT
	XETH_MSG_KIND_ETHTOOL_FLAGS
	XETH_MSG_KIND_ETHTOOL_SETTINGS
	XETH_MSG_KIND_DUMP_IFINFO
	XETH_MSG_KIND_CARRIER
	XETH_MSG_KIND_SPEED
	XETH_MSG_KIND_IFINDEX
	XETH_MSG_KIND_IFA
)

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
		"ifindex",
		"ifa",
	}
	var s string
	i := int(kind)
	if i < len(kinds) {
		s = kinds[i]
	} else {
		s = fmt.Sprint("kind[", i, "]")
	}
	return s
}
