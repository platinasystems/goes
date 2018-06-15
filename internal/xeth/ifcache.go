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

import "net"

var ifcache map[int32]*net.Interface

func InterfaceByIndex(ifindex int32) *net.Interface {
	if ifcache == nil {
		ifcache = make(map[int32]*net.Interface)
	}
	if itf, found := ifcache[ifindex]; found {
		return itf
	}
	itf, err := net.InterfaceByIndex(int(ifindex))
	if err != nil {
		itf = &net.Interface{
			Index: int(ifindex),
			Name:  "unknown",
		}
	}
	ifcache[ifindex] = itf
	return itf
}
