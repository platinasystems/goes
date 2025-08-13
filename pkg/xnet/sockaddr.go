// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	SizeofSockaddrInet4 = int(unsafe.Sizeof(unix.RawSockaddrInet4{}))
	SizeofSockaddrInet6 = int(unsafe.Sizeof(unix.RawSockaddrInet6{}))
)

type SAInBuf []byte

func (b SAInBuf) Decode() (ap netip.AddrPort, id uint32) {
	port := ByteOrder.Uint16(b[2:4])
	switch b.Family() {
	case unix.AF_INET:
		ptr := (*[4]byte)(b[4:8])
		ap = netip.AddrPortFrom(netip.AddrFrom4(*ptr), port)
	case unix.AF_INET6:
		ptr := (*[16]byte)(b[8:24])
		ap = netip.AddrPortFrom(netip.AddrFrom16(*ptr), port)
		id = binary.NativeEndian.Uint32(b[24:28])
	}
	return
}

// Returns number of bytes encoded: SizeofSockaddrInet4 or SizeofSockaddrInet6.
func (b SAInBuf) Encode(ap netip.AddrPort, id uint32) (n int) {
	addr, port := ap.Addr(), ap.Port()
	if addr.Is4In6() {
		addr = addr.Unmap()
	}
	if addr.Is4() {
		b.SetINET()
		copy(b[4:8], addr.AsSlice())
		n = SizeofSockaddrInet4
	} else {
		b.SetINET6()
		copy(b[8:24], addr.AsSlice())
		ByteOrder.PutUint32(b[24:28], id)
		n = SizeofSockaddrInet6
	}
	ByteOrder.PutUint16(b[2:4], port)
	return
}

func (b SAInBuf) String() string {
	ap, id := b.Decode()
	if id != 0 {
		return fmt.Sprintf("%v%d", ap, id)
	}
	return ap.String()
}
