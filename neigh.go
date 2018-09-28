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

func (msg *MsgNeighUpdate) CloneHardwareAddr() net.HardwareAddr {
	b := make([]byte, len(msg.Lladdr[:]))
	copy(b, msg.Lladdr[:])
	return net.HardwareAddr(b)
}

func (msg *MsgNeighUpdate) CloneIP() net.IP {
	b := make([]byte, msg.Len)
	copy(b, msg.Dst[:len(b)])
	return net.IP(b)
}
