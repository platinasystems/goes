// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"golang.org/x/sys/unix"
)

//go:generate sh -c "go tool cgo -godefs -- if_tun_linux.go > zif_tun_linux.go"

const (
	HasPI          = false
	CanTAP         = true
	CanPersist     = false
	CanChangeOwner = false
	CanChangeGroup = false
	DevNetTun      = "/dev/net/tun"
)

func New(
	unit uint,
	isTAP bool,
	persist bool,
	owner, group int,
	ha netif.HardwareAddr,
) (*os.File, error) {
	ifr := new(Ifreq)
	ifrp := uintptr(unsafe.Pointer(ifr))

	prefix := "tun"
	if isTAP {
		prefix = "tap"
		ifr.PutFlags(IFF_TAP | IFF_NO_PI)
	} else {
		ifr.PutFlags(IFF_TUN | IFF_NO_PI)
	}
	copy(ifr.Ifrn[:], []byte(fmt.Sprintf("%s%d", prefix, unit)))

	if _, err := os.Stat(DevNetTun); err != nil {
		const (
			major = 10
			minor = 200
			dev   = (major << 8) | (minor & 0xff) |
				((minor & 0xfff00) << 12)
		)
		if _, err = os.Stat("/dev/net"); err != nil {
			return nil, egress.Mark(err)
		}
		err = unix.Mknod(DevNetTun, unix.S_IFCHR, dev)
		if err != nil {
			return nil, egress.Mark(err)
		}
	}

	fd, err := unix.Open(DevNetTun, os.O_RDWR, 0)
	if err != nil {
		return nil, egress.Mark(err)
	}
	defer func() {
		if err != nil {
			unix.Close(fd)
		}
	}()
	defer egress.Recovery(&err)

	if err = ioctl(uintptr(fd), unix.TUNSETIFF, ifrp); err != nil {
		return nil, egress.Mark(err)
	}

	bzero(ifr.Ifrn[:])
	bzero(ifr.Ifru[:])

	if err = ioctl(uintptr(fd), unix.TUNGETIFF, ifrp); err != nil {
		return nil, egress.Mark(err)
	}

	ifname := gstring(ifr.Ifrn[:])

	if owner != Unset {
		err = ioctl(uintptr(fd), unix.TUNSETOWNER, uintptr(owner))
		if err != nil {
			return nil, egress.Mark(err)
		}
	}

	if group != Unset {
		err = ioctl(uintptr(fd), unix.TUNSETGROUP, uintptr(group))
		if err != nil {
			return nil, egress.Mark(err)
		}
	}

	if persist {
		err = ioctl(uintptr(fd), unix.TUNSETPERSIST, uintptr(1))
		if err != nil {
			return nil, egress.Mark(err)
		}
	}

	if isTAP {
		if err = setmac(ifname, ha); err != nil {
			return nil, egress.Mark(err)
		}
	}

	if err = unix.SetNonblock(fd, true); err != nil {
		return nil, egress.Mark(err)
	}

	return os.NewFile(uintptr(fd), ifname), nil
}

func setmac(ifname string, ha netif.HardwareAddr) error {
	//FIXME w/ netlink
	socki, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, 0)
	if err != nil {
		return fmt.Errorf("socket: %w", err)
	}
	sockp := uintptr(socki)
	defer unix.Close(socki)

	ifr := new(Ifreq)
	copy(ifr.Ifrn[:], []byte(ifname))
	err = ioctl(sockp, unix.SIOCGIFINDEX, uintptr(unsafe.Pointer(ifr)))
	if err != nil {
		return fmt.Errorf("get ifindex: %w", err)
	}
	_ = ifr.Index()
	err = ioctl(sockp, unix.SIOCGIFHWADDR, uintptr(unsafe.Pointer(ifr)))
	if err != nil {
		return fmt.Errorf("get hwaddr: %w", err)
	}
	_ = ifr.AF()
	old := netif.NewHardwareAddr()
	copy(old, ifr.HA())
	copy(ifr.HA(), ha)
	err = ioctl(sockp, unix.SIOCSIFHWADDR, uintptr(unsafe.Pointer(ifr)))
	if err != nil {
		err = fmt.Errorf("hwaddr(%v->%v): %w", old, ha, err)
	}
	return err
}

func (ifr *Ifreq) Flags() uint16 {
	return binary.NativeEndian.Uint16(ifr.Ifru[:2])
}

func (ifr *Ifreq) PutFlags(flags uint16) {
	binary.NativeEndian.PutUint16(ifr.Ifru[:2], flags)
}

func (ifr *Ifreq) Index() uint32 {
	return binary.NativeEndian.Uint32(ifr.Ifru[:4])
}

func (ifr *Ifreq) AF() uint16 {
	return binary.BigEndian.Uint16(ifr.Ifru[:2])
}

func (ifr *Ifreq) HA() net.HardwareAddr {
	return net.HardwareAddr(ifr.Ifru[2 : 2+6])
}

func bzero(a []byte) {
	for i := range a {
		a[i] = 0
	}
}

func gstring(a []byte) string {
	for i, b := range a {
		if b == 0 {
			return string(a[:i])
		}
	}
	return string(a)
}
