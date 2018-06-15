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
	"bytes"
	"fmt"
)

type Iff uint32

const (
	IFF_UP Iff = 1 << iota
	IFF_BROADCAST
	IFF_DEBUG
	IFF_LOOPBACK
	IFF_POINTOPOINT
	IFF_NOTRAILERS
	IFF_RUNNING
	IFF_NOARP
	IFF_PROMISC
	IFF_ALLMULTI
	IFF_MASTER
	IFF_SLAVE
	IFF_MULTICAST
	IFF_PORTSEL
	IFF_AUTOMEDIA
	IFF_DYNAMIC
	IFF_LOWER_UP
	IFF_DORMANT
	IFF_ECHO
)

func (iff Iff) String() string {
	var iffs = []string{
		"up",
		"broadcast",
		"debug",
		"loopback",
		"pointopoint",
		"notrailers",
		"running",
		"noarp",
		"promisc",
		"allmulti",
		"master",
		"slave",
		"multicast",
		"portsel",
		"automedia",
		"dynamic",
		"lower_up",
		"dormant",
		"echo",
	}
	var sep string
	if iff == 0 {
		return "none"
	}
	buf := new(bytes.Buffer)
	for i, s := range iffs {
		bit := Iff(1) << uint(i)
		if iff&bit == bit {
			fmt.Fprint(buf, sep, s)
			sep = ", "
		}
	}
	return buf.String()
}
