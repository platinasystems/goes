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
	IFA_ADD = NETDEV_UP
	IFA_DEL = NETDEV_DOWN
)

type IfaEvent int

func (event IfaEvent) String() string {
	var events = []string{
		IFA_ADD: "add",
		IFA_DEL: "del",
	}
	var s string
	i := int(event)
	if i < len(events) {
		s = events[i]
	}
	if len(s) == 0 {
		s = fmt.Sprint("@", i)
	}
	return s
}

func (ifa *MsgIfa) IsAdd() bool { return ifa.Event == IFA_ADD }
func (ifa *MsgIfa) IsDel() bool { return ifa.Event == IFA_DEL }

func (ifa *MsgIfa) Prefix() *net.IPNet {
	ipBuf := make([]byte, 4)
	maskBuf := make([]byte, 4)
	*(*uint32)(unsafe.Pointer(&ipBuf[0])) = ifa.Address
	*(*uint32)(unsafe.Pointer(&maskBuf[0])) = ifa.Mask
	return &net.IPNet{net.IP(ipBuf), net.IPMask(maskBuf)}
}

func (ifa *MsgIfa) String() string {
	kind := Kind(ifa.Kind)
	ifname := fmt.Sprint((*Ifname)(&ifa.Ifname), "[", ifa.Ifindex, "]")
	event := IfaEvent(ifa.Event)
	prefix := ifa.Prefix()
	return fmt.Sprintln(kind, ifname, event, prefix)
}
