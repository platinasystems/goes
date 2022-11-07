// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"encoding/binary"
	"net"
	"os"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/os/host"
)

type Configuration struct {
	Unit    uint
	IsTap   bool
	Persist bool
	Owner   int
	Group   int
	Link
}

const (
	Unset = -1

	IndexOfFlags = 0
	SizeOfFlags  = 2
	EndOfFlags   = IndexOfFlags + SizeOfFlags

	IndexOfProto = EndOfFlags
	SizeOfProto  = 2
	EndOfProto   = IndexOfProto + SizeOfProto

	SizeOfInfo = EndOfProto
	EndOfInfo  = EndOfProto

	IndexOfData = EndOfInfo

	IndexOfEthDst = IndexOfData
	EndOfEthDst   = IndexOfEthDst + 6

	IndexOfEthSrc = EndOfEthDst
	EndOfEthSrc   = IndexOfEthSrc + 6

	MinEthLen = EndOfEthProto

	IndexOfEthProto = EndOfEthSrc
	EndOfEthProto   = IndexOfEthProto + 2

	IndexOfIpv4Src = IndexOfData + 12
	EndOfIpv4Src   = IndexOfIpv4Src + net.IPv4len

	IndexOfIpv4Dst = EndOfIpv4Src
	EndOfIpv4Dst   = IndexOfIpv4Dst + net.IPv4len

	IndexOfIpv6Src = EndOfInfo + 8
	EndOfIpv6Src   = IndexOfIpv6Src + net.IPv6len

	IndexOfIpv6Dst = EndOfIpv6Src
	EndOfIpv6Dst   = IndexOfIpv6Dst + net.IPv6len

	MinPktLen = EndOfIpv4Dst
)

func Info(b []byte) (flags, proto uint16) {
	flags = host.ByteOrder.Uint16(b[IndexOfFlags:EndOfFlags])
	proto = binary.BigEndian.Uint16(b[IndexOfProto:EndOfProto])
	return flags, proto
}

// Zero flags and copy Eth frame proto to pkt info.
func SetTapInfo(b []byte) {
	b[0] = 0
	b[1] = 0
	copy(b[IndexOfProto:EndOfProto], b[IndexOfEthProto:EndOfEthProto])
}

// Set 4 byte prefix (2 byte flags, 2 byte proto) per IP version (b[4] >> 4).
func SetTunInfo(b []byte) {
	var proto uint16
	switch TunVer(b) {
	case 4:
		proto = syscall.AF_INET
	case 6:
		proto = syscall.AF_INET6
	}
	b[0] = 0
	b[1] = 0
	binary.BigEndian.PutUint16(b[IndexOfProto:EndOfProto], proto)
}

func TapAddrs(b []byte) (src, dst net.HardwareAddr) {
	src = net.HardwareAddr(b[IndexOfEthSrc:EndOfEthSrc])
	dst = net.HardwareAddr(b[IndexOfEthDst:EndOfEthDst])
	return
}

func TunAddrs(b []byte) (src, dst net.IP) {
	switch TunVer(b) {
	case 4:
		src = net.IP(b[IndexOfIpv4Src:EndOfIpv4Src])
		dst = net.IP(b[IndexOfIpv4Dst:EndOfIpv4Dst])
	case 6:
		src = net.IP(b[IndexOfIpv6Src:EndOfIpv6Src])
		dst = net.IP(b[IndexOfIpv6Dst:EndOfIpv6Dst])
	}
	return
}

func TunVer(b []byte) uint8 {
	return uint8(b[IndexOfData] >> 4)
}

func ioctl(fd uintptr, req uintptr, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}
