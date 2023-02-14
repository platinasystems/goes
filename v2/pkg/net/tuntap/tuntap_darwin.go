// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

//go:generate sh -c "go tool cgo -godefs -- if_utun_darwin.go > zif_utun_darwin.go"

const (
	HasPI  = true
	CanTAP = false
)

var namsiz = uintptr(IFNAMSIZ)
var namsizp = uintptr(unsafe.Pointer(&namsiz))

func New(cfg *Configuration) (*os.File, error) {
	if cfg.IsTap {
		return nil, fmt.Errorf("can't tap")
	}
	if cfg.Persist {
		return nil, fmt.Errorf("can't persist")
	}
	if cfg.Owner != Unset {
		return nil, fmt.Errorf("can't change owner")
	}
	if cfg.Group != Unset {
		return nil, fmt.Errorf("can't change group")
	}

	fdi, err := syscall.
		Socket(PF_SYSTEM, syscall.SOCK_DGRAM, SYSPROTO_CONTROL)
	if err != nil {
		return nil, fmt.Errorf("socket: %v", err)
	}
	fdp := uintptr(fdi)
	fdc := func() error { return syscall.Close(fdi) }

	ci := newCtlInfo()

	err = ioctl(fdp, CTLIOCGINFO, uintptr(unsafe.Pointer(ci)))
	if err != nil {
		return nil, egress.Marked(err, fdc)
	}

	sac := &SockaddrCtl{
		Sc_len:     SizeofSockaddrCtl,
		Sc_family:  AF_SYSTEM,
		Ss_sysaddr: AF_SYS_CONTROL,
		Sc_id:      ci.Id,
		Sc_unit:    uint32(cfg.Unit) + 1,
	}
	_, _, errno := syscall.RawSyscall(syscall.SYS_CONNECT, fdp,
		uintptr(unsafe.Pointer(sac)), SizeofSockaddrCtl)
	if errno != 0 {
		err = os.NewSyscallError("connect", errno)
		return nil, egress.Marked(err, fdc)
	}

	name := make([]byte, IFNAMSIZ, IFNAMSIZ)

	_, _, errno = syscall.Syscall6(syscall.SYS_GETSOCKOPT, fdp,
		SYSPROTO_CONTROL, UTUN_OPT_IFNAME,
		uintptr(unsafe.Pointer(&name[0])), namsizp, 0)
	if errno != 0 {
		err = os.NewSyscallError("ifname", errno)
		return nil, egress.Marked(err, fdc)
	}

	for i, b := range name {
		if b == 0 {
			name = name[:i]
			break
		}
	}

	if err = syscall.SetNonblock(fdi, true); err != nil {
		return nil, egress.Marked(err, fdc)
	}

	return os.NewFile(fdp, string(name)), nil
}

func newCtlInfo() *CtlInfo {
	ci := new(CtlInfo)
	for i, b := range []byte(UTUN_CONTROL_NAME) {
		ci.Name[i] = int8(b)
	}
	return ci
}
