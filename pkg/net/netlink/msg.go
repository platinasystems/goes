// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netlink

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/syscall/align"
)

type AttrTypes interface {
	~byte | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64
}

type MsgTypes interface {
	NlMsghdr | NlMsgerr | IfAddrmsg | IfInfomsg |
		RtAttr | RtGenmsg | RtMsg | RtNexthop |
		RtnlLinkStats[uint32] | RtnlLinkStats[uint64] |
		RtnlLinkIfmap
}

// Concatenate message with attribute header and type value padded to 4-byte
// alignment.
func CatAttr[T AttrTypes](msg []byte, ifla uint16, v T) []byte {
	rta, msg := Expand[RtAttr](msg)
	rta.Len = SizeofRtAttr + uint16(unsafe.Sizeof(v))
	rta.Type = ifla
	p, msg := Expand[T](msg)
	*p = v
	return msg
}

// Concatenate message with attribute header and given bytes padded to 4-byte
// alignment.
func CatBytesAttr(msg []byte, ifla uint16, b []byte) []byte {
	rta, msg := Expand[RtAttr](msg)
	rta.Len = SizeofRtAttr + uint16(len(b))
	rta.Type = ifla
	i := len(msg)
	msg = msg[:len(msg)+align.RTA.Roundup(len(b))]
	copy(msg[i:], b)
	return msg
}

// Concatenate message with attribute header and null terminate string padded
// to 4-byte alignment.
func CatStringAttr[S ~string](msg []byte, ifla uint16, s S) []byte {
	rta, msg := Expand[RtAttr](msg)
	rta.Len = SizeofRtAttr + uint16(len(s)) + 1
	rta.Type = ifla
	i := len(msg)
	msg = msg[:len(msg)+align.RTA.Roundup(len(s)+1)]
	copy(msg[i:], []byte(s))
	msg[i+len(s)] = 0
	return msg
}

// Expand data by type padded to 4-byte alignment.
func Expand[T AttrTypes | MsgTypes](data []byte) (p *T, x []byte) {
	i := len(data)
	size := align.RTA.Roundup(int(unsafe.Sizeof(*p)))
	if i+size > cap(data) {
		x = make([]byte, i+size, align.Page.Roundup(i+size))
		copy(x, data)
	} else {
		x = data[:i+size]
	}
	p = Pointer[T](x[i:])
	return
}

var (
	ExpandNlMsghdr  = Expand[NlMsghdr]
	ExpandIfAddrmsg = Expand[IfAddrmsg]
	ExpandIfInfomsg = Expand[IfInfomsg]
	ExpandRtGenmsg  = Expand[RtGenmsg]
)

// Return type at beginning of data along with the 4-byte aligned remainder.
func Extract[T AttrTypes | MsgTypes](data []byte) (p *T, r []byte) {
	p = Pointer[T](data)
	size := align.RTA.Roundup(int(unsafe.Sizeof(*p)))
	r = data[size:]
	return
}

var (
	ExtractNlMsghdr  = Extract[NlMsghdr]
	ExtractIfAddrmsg = Extract[IfAddrmsg]
	ExtractIfInfomsg = Extract[IfInfomsg]
)

func ExtractError(data []byte) error {
	msgerr, _ := Extract[NlMsgerr](data)
	if msgerr.Error != 0 {
		return Errno(-msgerr.Error)
	}
	return nil
}

func ExtractRtAttr(data []byte) (kind uint16, value, remainder []byte) {
	if len(data) < SizeofRtAttr {
		return
	}
	rta := Pointer[RtAttr](data)
	n := int(rta.Len)
	if n < SizeofRtAttr || n > len(data) {
		return
	}
	kind = rta.Type
	value = data[SizeofRtAttr:n]
	if n = align.RTA.Roundup(n); n < len(data) {
		remainder = data[n:]
	}
	return
}

// Return type at beginning of data.
func Pointer[T AttrTypes | MsgTypes](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

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
