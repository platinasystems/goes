// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

//go:generate sh -c "go tool cgo -godefs -- godefs_linux.go > godefed_linux.go"

const TUNDEV = "/dev/net/tun"

func New(cfg *Configuration) (*os.File, error) {
	ifr := new(Ifreq)
	ifrp := uintptr(unsafe.Pointer(ifr))

	if cfg.IsTap {
		ifr.SetName(fmt.Sprintf("tap%d", cfg.Unit))
		ifr.SetFlags(IFF_TAP)
	} else if !cfg.Link.IsAutogen() {
		return nil, fmt.Errorf("can't set tun link address")
	} else {
		ifr.SetName(fmt.Sprintf("tun%d", cfg.Unit))
		ifr.SetFlags(IFF_TUN)
	}

	fdi, err := syscall.Open(TUNDEV, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	fdp := uintptr(fdi)
	fdc := egress.New[*os.File](func() { syscall.Close(fdi) })

	err = ioctl(fdp, syscall.TUNSETIFF, ifrp)
	if err != nil {
		return fdc(fmt.Errorf("set iff: %w", err))
	}

	ifr.ClearName()
	ifr.ClearUnion()
	err = ioctl(fdp, syscall.TUNGETIFF, ifrp)
	if err != nil {
		return fdc(fmt.Errorf("get iff: %w", err))
	}
	ifname := ifr.Name()

	if cfg.Owner != Unset {
		err = ioctl(fdp, syscall.TUNSETOWNER, uintptr(cfg.Owner))
		if err != nil {
			return fdc(fmt.Errorf("owner: %w", err))
		}
	}

	if cfg.Group != Unset {
		err = ioctl(fdp, syscall.TUNSETGROUP, uintptr(cfg.Group))
		if err != nil {
			return fdc(fmt.Errorf("group: %w", err))
		}
	}

	if cfg.Persist {
		err = ioctl(fdp, syscall.TUNSETPERSIST, uintptr(1))
		if err != nil {
			return fdc(fmt.Errorf("persist: %w", err))
		}
	}

	if !cfg.Link.IsAutogen() {
		if err = setlink(ifname, cfg.Link); err != nil {
			return fdc(err)
		}
	}

	if err = syscall.SetNonblock(fdi, true); err != nil {
		return fdc(fmt.Errorf("non-block: %w", err))
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
	ifr.SetName(ifname)
	err = ioctl(sockp, syscall.SIOCGIFINDEX, ifrp)
	if err != nil {
		return fmt.Errorf("get ifindex: %w", err)
	}
	ifindex := host.ByteOrder.Uint32(ifr.Ifru[:4])
	_ = ifindex
	err = ioctl(sockp, syscall.SIOCGIFHWADDR, ifrp)
	if err != nil {
		return fmt.Errorf("get hwaddr: %w", err)
	}
	af, ha := ifr.Link()
	ifr.SetLink(af, link)
	err = ioctl(sockp, syscall.SIOCSIFHWADDR, ifrp)
	if err != nil {
		err = fmt.Errorf("set hwaddr(%v->%v): %w", ha, link, err)
	}
	return err
}

func bzero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func newIfreq(name string) *Ifreq {
	ifr := new(Ifreq)
	return ifr
}

func (ifr *Ifreq) ClearName()  { bzero(ifr.Ifrn[:]) }
func (ifr *Ifreq) ClearUnion() { bzero(ifr.Ifru[:]) }

func (ifr *Ifreq) Name() string {
	name := make([]byte, IFNAMSIZ, IFNAMSIZ)
	for i, b := range ifr.Ifrn[:] {
		if b == 0 {
			name = name[:i]
			break
		}
		name[i] = b
	}
	return string(name)
}

func (ifr *Ifreq) SetName(name string) {
	for i, b := range []byte(name) {
		ifr.Ifrn[i] = b
	}
}

func (ifr *Ifreq) Flags() uint16 {
	return host.ByteOrder.Uint16(ifr.Ifru[:2])
}

func (ifr *Ifreq) SetFlags(flags uint16) {
	host.ByteOrder.PutUint16(ifr.Ifru[:2], flags)
}

func (ifr *Ifreq) Link() (af uint16, ha net.HardwareAddr) {
	af = host.ByteOrder.Uint16(ifr.Ifru[:2])
	ha = make([]byte, 6, 6)
	copy(ha, ifr.Ifru[2:2+6])
	return
}

func (ifr *Ifreq) SetLink(af uint16, link Link) {
	host.ByteOrder.PutUint16(ifr.Ifru[:2], af)
	for i, b := range link.HardwareAddr {
		if i >= 6 {
			break
		}
		if i == 0 {
			b &= 0xfe // clear multicast
		}
		ifr.Ifru[i+2] = b
	}
}
