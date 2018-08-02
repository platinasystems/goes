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
	"reflect"
	"unsafe"
)

const (
	FIB_EVENT_ENTRY_REPLACE = iota
	FIB_EVENT_ENTRY_APPEND
	FIB_EVENT_ENTRY_ADD
	FIB_EVENT_ENTRY_DEL
	FIB_EVENT_RULE_ADD
	FIB_EVENT_RULE_DEL
	FIB_EVENT_NH_ADD
	FIB_EVENT_NH_DEL
)

type FibEntryEvent int
type FibRuleEvent int
type FibNHEvent int

func (event FibEntryEvent) String() string {
	var events = []string{
		FIB_EVENT_ENTRY_REPLACE: "replace",
		FIB_EVENT_ENTRY_APPEND:  "append",
		FIB_EVENT_ENTRY_ADD:     "add",
		FIB_EVENT_ENTRY_DEL:     "del",
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

func (event FibRuleEvent) String() string {
	var events = []string{
		FIB_EVENT_RULE_ADD: "add",
		FIB_EVENT_RULE_DEL: "del",
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

func (event FibNHEvent) String() string {
	var events = []string{
		FIB_EVENT_NH_ADD: "add",
		FIB_EVENT_NH_DEL: "del",
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

func (fe *MsgFibentry) NextHop(i int) *NextHop {
	ptr := unsafe.Pointer(fe)
	nhptr := uintptr(ptr) + uintptr(SizeofMsgFibentry)
	return (*NextHop)(unsafe.Pointer(nhptr + uintptr(SizeofNextHop*i)))
}

func (fe *MsgFibentry) NextHops() []NextHop {
	ptr := unsafe.Pointer(fe)
	nhs := int(fe.Nhs)
	return *(*[]NextHop)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(ptr) + uintptr(SizeofMsgFibentry),
		Len:  nhs,
		Cap:  nhs,
	}))
}

func (fe *MsgFibentry) Prefix() *net.IPNet {
	ipBuf := make([]byte, 4)
	maskBuf := make([]byte, 4)
	*(*uint32)(unsafe.Pointer(&ipBuf[0])) = fe.Address
	*(*uint32)(unsafe.Pointer(&maskBuf[0])) = fe.Mask
	return &net.IPNet{net.IP(ipBuf), net.IPMask(maskBuf)}
}

func (fe *MsgFibentry) String() string {
	kind := Kind(fe.Kind)
	event := FibEntryEvent(fe.Event)
	prefix := fe.Prefix()
	return fmt.Sprintln(kind, event, Rtn(fe.Type), prefix,
		"netns", Netns(fe.Net),
		"table", RtTable(fe.Id),
		fe.NextHops())
}

func (nh *NextHop) IP() net.IP {
	buf := make([]byte, 4)
	*(*uint32)(unsafe.Pointer(&buf[0])) = nh.Gw
	return net.IP(buf)
}

func (nh NextHop) String() string {
	return fmt.Sprint(nh.IP(), "@", InterfaceByIndex(nh.Ifindex).Name)
}
