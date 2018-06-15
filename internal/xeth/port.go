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
	PORT_TP = iota
	PORT_AUI
	PORT_MII
	PORT_FIBRE
	PORT_BNC
	PORT_DA
)

const (
	PORT_NONE  = 0xef
	PORT_OTHER = 0xff
)

type Port uint8

func (port Port) String() string {
	var ports = []string{
		"tp",
		"aui",
		"mii",
		"fibre",
		"bnc",
		"da",
	}
	i := int(port)
	if i < len(ports) {
		return ports[i]
	} else if i == PORT_NONE {
		return "none"
	} else if i == PORT_OTHER {
		return "other"
	}
	return fmt.Sprint("@", i)
}
