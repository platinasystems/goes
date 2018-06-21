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
	RT_TABLE_UNSPEC = 0
	// User defined values
	RT_TABLE_COMPAT  = 252
	RT_TABLE_DEFAULT = 253
	RT_TABLE_MAIN    = 254
	RT_TABLE_LOCAL   = 255
	RT_TABLE_MAX     = 0xFFFFFFFF
)

type RtTable uint32

func (rtt RtTable) String() string {
	var s string
	switch rtt {
	case RT_TABLE_UNSPEC:
		s = "unspec"
	case RT_TABLE_COMPAT:
		s = "compat"
	case RT_TABLE_DEFAULT:
		s = "default"
	case RT_TABLE_MAIN:
		s = "main"
	case RT_TABLE_LOCAL:
		s = "local"
	case RT_TABLE_MAX:
		s = "max"
	default:
		s = fmt.Sprint(uint32(rtt))
	}
	return s
}
