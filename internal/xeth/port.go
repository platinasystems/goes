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

type Port uint8

const (
	PORT_TP    Port = 0x00
	PORT_AUI   Port = 0x01
	PORT_MII   Port = 0x02
	PORT_FIBRE Port = 0x03
	PORT_BNC   Port = 0x04
	PORT_DA    Port = 0x05
	PORT_NONE  Port = 0xef
	PORT_OTHER Port = 0xff
)

func (port Port) String() string {
	s, found := map[Port]string{
		PORT_TP:    "tp",
		PORT_AUI:   "aui",
		PORT_MII:   "mii",
		PORT_FIBRE: "fibre",
		PORT_BNC:   "bnc",
		PORT_DA:    "da",
		PORT_NONE:  "none",
		PORT_OTHER: "other",
	}[port]
	if !found {
		s = "invalid"
	}
	return s
}
