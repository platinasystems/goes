// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

//go:generate sh -c "go tool cgo -godefs -- if_tun_linux.go > zif_tun_linux.go"

const (
	HasPI  = false
	CanTAP = true
	DevTun = "/dev/net/tun"
)

func New(cfg *Configuration) (*os.File, error) {
	ifr := new(Ifreq)
	ifrp := uintptr(unsafe.Pointer(ifr))

	prefix := "tun"
	if cfg.IsTap {
		prefix = "tap"
		ifr.Flags().Put(IFF_TAP | IFF_NO_PI)
	} else if !cfg.Link.IsAutogen() {
		return nil, fmt.Errorf("can't set tun link address")
	} else {
		ifr.Flags().Put(IFF_TUN | IFF_NO_PI)
	}
	copy(ifr.Ifrn[:], []byte(fmt.Sprintf("%s%d", prefix, cfg.Unit)))

	fdi, err := syscall.Open(DevTun, os.O_RDWR, 0)
	if err != nil {
		err = fmt.Errorf("%s: %w", DevTun, err)
		return nil, err
	}
	fdp := uintptr(fdi)
	fdc := func() error { return syscall.Close(fdi) }

	err = ioctl(fdp, syscall.TUNSETIFF, ifrp)
	if err != nil {
		return nil, egress.Marked(err, fdc)
	}

	bzero(ifr.Ifrn[:])
	bzero(ifr.Ifru[:])

	err = ioctl(fdp, syscall.TUNGETIFF, ifrp)
	if err != nil {
		return nil, egress.Marked(err, fdc)
	}

	ifname := gstring(ifr.Ifrn[:])

	if cfg.Owner != Unset {
		err = ioctl(fdp, syscall.TUNSETOWNER, uintptr(cfg.Owner))
		if err != nil {
			return nil, egress.Marked(err, fdc)
		}
	}

	if cfg.Group != Unset {
		err = ioctl(fdp, syscall.TUNSETGROUP, uintptr(cfg.Group))
		if err != nil {
			return nil, egress.Marked(err, fdc)
		}
	}

	if cfg.Persist {
		err = ioctl(fdp, syscall.TUNSETPERSIST, uintptr(1))
		if err != nil {
			return nil, egress.Marked(err, fdc)
		}
	}

	if !cfg.Link.IsAutogen() {
		if err = setlink(ifname, cfg.Link); err != nil {
			return nil, egress.Marked(err, fdc)
		}
	}

	if err = syscall.SetNonblock(fdi, true); err != nil {
		return nil, egress.Marked(err, fdc)
	}

	return os.NewFile(fdp, ifname), nil
}

func setlink(ifname string, link Link) error {
	socki, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		return fmt.Errorf("socket: %w", err)
	}
	sockp := uintptr(socki)
	defer syscall.Close(socki)

	ifr := new(Ifreq)
	ifrp := uintptr(unsafe.Pointer(ifr))
	copy(ifr.Ifrn[:], []byte(ifname))
	err = ioctl(sockp, syscall.SIOCGIFINDEX, ifrp)
	if err != nil {
		return fmt.Errorf("get ifindex: %w", err)
	}
	_ = ifr.Index().Value()
	err = ioctl(sockp, syscall.SIOCGIFHWADDR, ifrp)
	if err != nil {
		return fmt.Errorf("get hwaddr: %w", err)
	}
	_ = ifr.AF().Value()
	ha := make(net.HardwareAddr, 6)
	copy(ha, ifr.HA())
	link.HardwareAddr[0] &= 0xfe // clear multicast
	copy(ifr.HA(), link.HardwareAddr)
	err = ioctl(sockp, syscall.SIOCSIFHWADDR, ifrp)
	if err != nil {
		err = fmt.Errorf("hwaddr(%v->%v): %w", ha, link, err)
	}
	return err
}

func (ifr *Ifreq) Flags() *host.Uint16 {
	return (*host.Uint16)(ifr.Ifru[:2])
}

func (ifr *Ifreq) Index() *host.Uint32 {
	return (*host.Uint32)(ifr.Ifru[:4])
}

func (ifr *Ifreq) AF() *big.Uint16 {
	return (*big.Uint16)(ifr.Ifru[:2])
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
