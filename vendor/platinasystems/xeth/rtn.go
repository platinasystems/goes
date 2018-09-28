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

const (
	RTN_UNSPEC = iota
	// Gateway or direct route
	RTN_UNICAST
	// Accept locally
	RTN_LOCAL
	// Accept locally as broadcast, send as broadcast
	RTN_BROADCAST
	// Accept locally as broadcast, but send as unicast
	RTN_ANYCAST
	// Multicast route
	RTN_MULTICAST
	// Drop
	RTN_BLACKHOLE
	// Destination is unreachable
	RTN_UNREACHABLE
	// Administratively prohibited
	RTN_PROHIBIT
	// Not in this table
	RTN_THROW
	// Translate this address
	RTN_NAT
	// Use external resolver
	RTN_XRESOLVE
	__RTN_MAX
)

const RTN_MAX = __RTN_MAX - 1

type Rtn uint8

func (rtn Rtn) String() string {
	var rtns = []string{
		"unspec",
		"unicast",
		"local",
		"broadcast",
		"anycast",
		"multicast",
		"blackhole",
		"unreachable",
		"prohibit",
		"throw",
		"nat",
		"xresolve",
	}
	i := int(rtn)
	if i < len(rtns) {
		return rtns[i]
	}
	return fmt.Sprint("@", i)
}
