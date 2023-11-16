// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || ios || freebsd || openbsd || netbsd

package sockaddr

import (
	"net"
	"net/netip"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/net/sysctl"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
	"github.com/platinasystems/goes/v2/pkg/syscall/align"
)

const Min = 2

func Len(data []byte) int     { return int(data[0]) }
func Family(data []byte) uint { return uint(data[1]) }

type In struct{ syscall.RawSockaddrInet4 }

func (sa *In) Write(addr []byte) (int, error) {
	sa.Len = uint8(Sizeof(sa))
	sa.Family = af.INET
	return copy(sa.Addr[:], addr), nil
}

type In6 struct{ syscall.RawSockaddrInet6 }

func (sa *In6) Write(addr []byte) (int, error) {
	sa.Len = uint8(Sizeof(sa))
	sa.Family = af.INET6
	return copy(sa.Addr[:], addr), nil
}

func IP4(data []byte) (ip netip.Addr) {
	if sa := Pointer[In](data); len(data) > -Sizeof(sa) {
		ip = netip.AddrFrom4(sa.Addr)
	}
	return
}

func IP6(data []byte) (ip netip.Addr) {
	if sa := Pointer[In6](data); len(data) > -Sizeof(sa) {
		ip = netip.AddrFrom16(sa.Addr)
	}
	return
}

func IP(data []byte) (ip netip.Addr) {
	switch Family(data) {
	case af.INET:
		ip = IP4(data)
	case af.INET6:
		ip = IP6(data)
	}
	return
}

// Replaced syscall.RawSockaddrDatalink b/c it's
//
//	Data   [12]int8
//
// is insufficient for name, address and possible selector.
type DlHdr struct {
	Len    uint8
	Family uint8
	Index  uint16
	Type   uint8
	Nlen   uint8
	Alen   uint8
	Slen   uint8
}

func (dl *DlHdr) NAS(body []byte) (
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

func Expand[T In | In6 | DlHdr](data []byte) (t *T, x []byte) {
	i := len(data)
	size := sysctl.Align(Sizeof(t))
	if i+size > cap(data) {
		x = make([]byte, i+size, sysctl.PageAlign(i+size))
		copy(x, data)
	} else {
		x = data[:i+size]
	}
	t = Pointer[T](x[i:])
	return
}

var (
	ExpandIn    = Expand[In]
	ExpandIn6   = Expand[In6]
	ExpandDlHdr = Expand[DlHdr]
)

func Append(msg []byte, addr netip.Addr) []byte {
	if addr.Is4() {
		sain, x := ExpandIn(msg)
		sain.Write(addr.AsSlice())
		msg = x
	} else if addr.Is6() {
		sain6, x := ExpandIn6(msg)
		sain6.Write(addr.AsSlice())
		msg = x
	}
	return msg
}

func AppendDl[A ~[]byte, S ~[]byte](
	msg []byte,
	i uint16,
	t uint8,
	name string,
	address A,
	selector S,
) []byte {
	n := len(msg)
	sadl, msg := ExpandDlHdr(msg)
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
	size := sysctl.Align(len(msg))
	if size > cap(msg) {
		x := make([]byte, size)
		copy(x, msg)
		msg = x
	} else {
		msg = msg[:size]
	}
	sadl.Len = uint8(len(msg[n:]))
	sadl.Family = af.LINK
	sadl.Index = i
	sadl.Type = t
	sadl.Nlen = uint8(nlen)
	sadl.Alen = uint8(alen)
	sadl.Slen = uint8(slen)
	return msg
}

func Extract[T In | In6 | DlHdr](data []byte) (t *T, body, rem []byte) {
	l := int(data[0])
	t = Pointer[T](data)
	body = data[Sizeof(t):]
	rem = data[sysctl.Align(l):]
	return
}

var (
	ExtractIn    = Extract[In]
	ExtractIn6   = Extract[In6]
	ExtractDlHdr = Extract[DlHdr]
)

func Pointer[T In | In6 | DlHdr](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

var (
	PointerIn    = Pointer[In]
	PointerIn6   = Pointer[In6]
	PointerDlHdr = Pointer[DlHdr]
)

func Sizeof[T In | In6 | DlHdr](p *T) int {
	return int(unsafe.Sizeof(*p))
}

const (
	SizeofIn    = int(unsafe.Sizeof(&In{}))
	SizeofIn6   = int(unsafe.Sizeof(&In6{}))
	SizeofDlHdr = int(unsafe.Sizeof(&DlHdr{}))
	SizeofLong  = 4
)

var Align = align.Align(SizeofLong).Roundup
