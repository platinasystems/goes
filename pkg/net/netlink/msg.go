// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

import (
	"net/netip"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/ifaddr"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/iflink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

type Messages interface {
	MsgHdr | MsgErr |
		ifaddr.Msg |
		rtnetlink.IfInfoMsg |
		rtnetlink.RtGenMsg |
		rtnetlink.RtMsg
}

type Attributes interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int8 | ~int16 | ~int32 | ~int64 |
		Attr |
		ifaddr.Attr |
		iflink.Attr |
		rtnetlink.Attr |
		iflink.IfMap |
		iflink.Stats[uint32] |
		iflink.Stats[uint64] |
		rtnetlink.RtNextHop
}

type AttrType interface {
	integer.Int | integer.Uint
}

// Concatenate message with attribute header and type value padded to 4-byte
// alignment.
func CatAttr[T AttrType, A Attributes](msg []byte, t T, v A) []byte {
	attr, msg := Expand[Attr](msg)
	integer.Assign(&attr.Len, unsafe.Sizeof(*attr))
	integer.Add(&attr.Len, unsafe.Sizeof(v))
	integer.Assign(&attr.Type, t)
	p, msg := Expand[A](msg)
	*p = v
	return msg
}

// Concatenate message with attribute header and given bytes padded to 4-byte
// alignment.
func CatBytesAttr[T AttrType](msg []byte, t T, b []byte) []byte {
	attr, msg := Expand[Attr](msg)
	integer.Assign(&attr.Len, unsafe.Sizeof(*attr))
	integer.Add(&attr.Len, len(b))
	integer.Assign(&attr.Type, t)
	i := len(msg)
	msg = grow(msg, len(b))
	copy(msg[i:], b)
	return msg
}

// Concatenate message with attribute header and null terminate string padded
// to 4-byte alignment.
func CatStringAttr[T AttrType, S ~string](msg []byte, t T, s S) []byte {
	attr, msg := Expand[Attr](msg)
	integer.Assign(&attr.Len, unsafe.Sizeof(*attr))
	integer.Add(&attr.Len, len(s)+1)
	integer.Assign(&attr.Type, t)
	i := len(msg)
	b := []byte(s)
	msg = grow(msg, len(b)+1)
	copy(msg[i:], b)
	msg[i+len(b)] = 0
	return msg
}

// Expand data by type padded to 4-byte alignment.
func Expand[T Attributes | Messages](data []byte) (p *T, x []byte) {
	i := len(data)
	x = grow(data, int(unsafe.Sizeof(*p)))
	p = Pointer[T](x[i:])
	return
}

func grow(data []byte, n int) (x []byte) {
	i := len(data)
	n = i + NLA_ALIGN(n)
	if n > cap(data) {
		x = make([]byte, n, page.Align(n))
		copy(x, data)
	} else {
		x = data[:n]
	}
	return
}

var (
	ExpandMsgHdr    = Expand[MsgHdr]
	ExpandIfAddrMsg = Expand[ifaddr.Msg]
	ExpandIfInfoMsg = Expand[rtnetlink.IfInfoMsg]
	ExpandRtGenMsg  = Expand[rtnetlink.RtGenMsg]
	ExpandRtMsg     = Expand[rtnetlink.RtMsg]
)

// Return type at beginning of data along with the aligned remainder.
func ExtractMsg[M Messages](data []byte) (p *M, r []byte) {
	p = Pointer[M](data)
	size := NLMSG_ALIGN(int(unsafe.Sizeof(*p)))
	r = data[size:]
	return
}

var (
	ExtractMsgHdr    = ExtractMsg[MsgHdr]
	ExtractMsgErr    = ExtractMsg[MsgErr]
	ExtractIfAddrMsg = ExtractMsg[ifaddr.Msg]
	ExtractIfInfoMsg = ExtractMsg[rtnetlink.IfInfoMsg]
	ExtractRtMsg     = ExtractMsg[rtnetlink.RtMsg]
)

func (m *MsgErr) Err() error {
	if m.Error != 0 {
		return syscall.Errno(-m.Error)
	}
	return nil
}

func HasAttr(data []byte) bool {
	return len(data) >= NLA_ALIGNTO
}

func ExtractAttr(data []byte) (t uint16, value, remainder []byte) {
	attr := Pointer[Attr](data)
	asz := int(unsafe.Sizeof(*attr))
	if len(data) < asz {
		return
	}
	n := int(attr.Len)
	if n < asz || n > len(data) {
		return
	}
	t = attr.Type
	value = data[asz:n]
	if n = NLA_ALIGN(n); n < len(data) {
		remainder = data[n:]
	} else {
		remainder = data[len(data):]
	}
	return
}

// Return type at beginning of data.
func Pointer[T Attributes | Messages](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

var PointerMsgHdr = Pointer[MsgHdr]

func Clone(data []byte) []byte {
	clone := make([]byte, len(data))
	copy(clone, data)
	return clone
}

// Clone data up to its first null (`\0`) if any.
func CloneString(data []byte) string {
	for i, b := range data {
		if b == 0 {
			data = data[:i]
			break
		}
	}
	return string(Clone(data))
}

func IP(family uint8, data []byte) (netip.Addr, bool) {
	switch family {
	case af.INET:
		data = data[:4]
	case af.INET6:
		data = data[:16]
	}
	return netip.AddrFromSlice(data)
}

func Via(data []byte) (netip.Addr, bool) {
	family := *(Pointer[uint16](data))
	return IP(uint8(family), data[2:])
}
