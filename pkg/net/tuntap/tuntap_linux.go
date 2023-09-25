// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
)

//go:generate sh -c "go tool cgo -godefs -- if_tun_linux.go > zif_tun_linux.go"

const (
	HasPI          = false
	CanTAP         = true
	CanPersist     = false
	CanChangeOwner = false
	CanChangeGroup = false
	DevTun         = "/dev/net/tun"
)

func New(
	unit uint,
	isTAP bool,
	persist bool,
	owner, group int,
	ha netif.HardwareAddr,
) (t *os.File, err error) {
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

	fd, err := syscall.Open(DevTun, os.O_RDWR, 0)
	if err != nil {
		err = fmt.Errorf("%s: %w", DevTun, err)
		return
	}
	defer func() {
		if err != nil {
			syscall.Close(fd)
		}
	}()
	defer egress.Recovery(&err)

	if err = ioctl(uintptr(fd), syscall.TUNSETIFF, ifrp); err != nil {
		panic(err)
	}

	bzero(ifr.Ifrn[:])
	bzero(ifr.Ifru[:])

	if err = ioctl(uintptr(fd), syscall.TUNGETIFF, ifrp); err != nil {
		panic(err)
	}

	ifname := gstring(ifr.Ifrn[:])

	if owner != Unset {
		err = ioctl(uintptr(fd), syscall.TUNSETOWNER, uintptr(owner))
		if err != nil {
			panic(err)
		}
	}

	if group != Unset {
		err = ioctl(uintptr(fd), syscall.TUNSETGROUP, uintptr(group))
		if err != nil {
			panic(err)
		}
	}

	if persist {
		err = ioctl(uintptr(fd), syscall.TUNSETPERSIST, uintptr(1))
		if err != nil {
			panic(err)
		}
	}

	if isTAP {
		if err = setmac(ifname, ha); err != nil {
			panic(err)
		}
	}

	if err = syscall.SetNonblock(fd, true); err != nil {
		panic(err)
	}

	t = os.NewFile(uintptr(fd), ifname)
	return
}

func setmac(ifname string, ha netif.HardwareAddr) error {
	//FIXME w/ netlink
	socki, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return fmt.Errorf("socket: %w", err)
	}
	sockp := uintptr(socki)
	defer syscall.Close(socki)

	ifr := new(Ifreq)
	copy(ifr.Ifrn[:], []byte(ifname))
	err = ioctl(sockp, syscall.SIOCGIFINDEX, uintptr(unsafe.Pointer(ifr)))
	if err != nil {
		return fmt.Errorf("get ifindex: %w", err)
	}
	_ = ifr.Index()
	err = ioctl(sockp, syscall.SIOCGIFHWADDR, uintptr(unsafe.Pointer(ifr)))
	if err != nil {
		return fmt.Errorf("get hwaddr: %w", err)
	}
	_ = ifr.AF()
	old := netif.NewHardwareAddr()
	copy(old, ifr.HA())
	copy(ifr.HA(), ha)
	err = ioctl(sockp, syscall.SIOCSIFHWADDR, uintptr(unsafe.Pointer(ifr)))
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
