// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || ios || freebsd || openbsd || netbsd

package xnet

import (
	"net"
	"net/netip"
	"unsafe"

	"golang.org/x/sys/unix"
)

const SAMin = 2

var (
	SAExpandIn    = SAExpand[SAIn]
	SAExtractIn   = SAExtract[SAIn]
	SAPointerSAIn = SAPointer[SAIn]
)

var (
	SAExpandIn6    = SAExpand[SAIn6]
	SAExtractIn6   = SAExtract[SAIn6]
	SAPointerSAIn6 = SAPointer[SAIn6]
)

var (
	SAExpandDataLink    = SAExpand[SADataLink]
	SAExtractDataLink   = SAExtract[SADataLink]
	SAPointerSADataLink = SAPointer[SADataLink]
)

// sockaddr_in
type SAIn struct{ unix.RawSockaddrInet4 }

// sockaddr_in6
type SAIn6 struct{ unix.RawSockaddrInet6 }

// Replaced sockaddr_dl [unix.RawSockaddrDatalink] b/c it's
//
//	Data   [12]int8
//
// is insufficient for name, address and possible selector.
type SADataLink struct {
	Len    uint8
	Family uint8
	Index  uint16
	Type   uint8
	Nlen   uint8
	Alen   uint8
	Slen   uint8
}

func SALen(data []byte) int     { return int(data[0]) }
func SAFamily(data []byte) uint { return uint(data[1]) }

func (sa *SAIn) Write(addr []byte) (int, error) {
	sa.Len = uint8(Sizeof(sa))
	sa.Family = AF_INET
	return copy(sa.Addr[:], addr), nil
}

func (sa *SAIn6) Write(addr []byte) (int, error) {
	sa.Len = uint8(Sizeof(sa))
	sa.Family = AF_INET6
	return copy(sa.Addr[:], addr), nil
}

func SAIP4(data []byte) (ip netip.Addr) {
	if sa := SAPointer[SAIn](data); len(data) > -Sizeof(sa) {
		ip = netip.AddrFrom4(sa.Addr)
	}
	return
}

func SAIP6(data []byte) (ip netip.Addr) {
	if sa := SAPointer[SAIn6](data); len(data) > -Sizeof(sa) {
		ip = netip.AddrFrom16(sa.Addr)
	}
	return
}

func SAIP(data []byte) netip.Addr {
	switch SAFamily(data) {
	case AF_INET:
		return SAIP4(data)
	case AF_INET6:
		return SAIP6(data)
	}
	return netip.Addr{}
}

func (dl *SADataLink) NAS(body []byte) (
	name string,
	address net.HardwareAddr,
	selector []byte,
) {
	nlen, alen, slen := int(dl.Nlen), int(dl.Alen), int(dl.Slen)
	if nlen == 0xff {
		nlen = 0
	}
	if alen == 0xff {
		alen = 0
	}
	if slen == 0xff {
		slen = 0
	}
	if len(body) < nlen+alen+slen {
		panic("insufficient data")
	}
	if nlen > 0 && nlen < len(body) {
		b := make([]byte, nlen)
		copy(b, body[:nlen])
		for i, c := range b {
			if c == 0 {
				b = b[:i]
				break
			}
		}
		name = string(b)
		body = body[nlen:]
	}
	if alen > 0 && alen < len(body) {
		address = make(net.HardwareAddr, alen)
		copy(address, body[:alen])
		body = body[alen:]
	}
	if slen > 0 && slen < len(body) {
		selector = make([]byte, slen)
		copy(selector, body[:slen])
	}
	return
}

func SAExpand[T SAIn | SAIn6 | SADataLink](data []byte) (t *T, x []byte) {
	i := len(data)
	size := SysctlAlign(Sizeof(t))
	if i+size > cap(data) {
		x = make([]byte, i+size, PageAlign(i+size))
		copy(x, data)
	} else {
		x = data[:i+size]
	}
	t = SAPointer[T](x[i:])
	return
}

func SAAppend(msg []byte, addr netip.Addr) []byte {
	if addr.Is4() {
		sain, x := SAExpandIn(msg)
		sain.Write(addr.AsSlice())
		msg = x
	} else if addr.Is6() {
		sain6, x := SAExpandIn6(msg)
		sain6.Write(addr.AsSlice())
		msg = x
	}
	return msg
}

func SAAppendDataLink[A ~[]byte, S ~[]byte](
	msg []byte,
	i uint16,
	t uint8,
	name string,
	address A,
	selector S,
) []byte {
	n := len(msg)
	sadl, msg := SAExpandDataLink(msg)
	nbytes := []byte(name)
	abytes := []byte(address)
	sbytes := []byte(selector)
	nlen := len(nbytes)
	alen := len(abytes)
	slen := len(sbytes)
	if nlen > 0 {
		msg = append(msg, nbytes...)
	}
	if alen > 0 {
		msg = append(msg, abytes...)
	}
	if slen > 0 {
		msg = append(msg, sbytes...)
	}
	size := SysctlAlign(len(msg))
	if size > cap(msg) {
		x := make([]byte, size)
		copy(x, msg)
		msg = x
	} else {
		msg = msg[:size]
	}
	sadl.Len = uint8(len(msg[n:]))
	sadl.Family = AF_LINK
	sadl.Index = i
	sadl.Type = t
	sadl.Nlen = uint8(nlen)
	sadl.Alen = uint8(alen)
	sadl.Slen = uint8(slen)
	return msg
}

func SAExtract[T SAIn | SAIn6 | SADataLink](data []byte) (
	t *T, body, rem []byte,
) {
	if n, i := len(data), Sizeof(t); n >= i {
		t = SAPointer[T](data)
		body = data[i:]
		i = SysctlAlign(int(data[0]))
		if n <= i {
			i = n
		}
		rem = data[i:]
	}
	return
}

func SAPointer[T SAIn | SAIn6 | SADataLink](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

func Sizeof[T SAIn | SAIn6 | SADataLink](p *T) int {
	return int(unsafe.Sizeof(*p))
}
