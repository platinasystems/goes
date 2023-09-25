// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
)

//go:generate sh -c "go tool cgo -godefs -- if_utun_darwin.go > zif_utun_darwin.go"

const (
	HasPI          = true
	CanTAP         = false
	CanPersist     = false
	CanChangeOwner = false
	CanChangeGroup = false
)

var (
	ErrCantTAP         = errors.New(runtime.GOOS + " can't TAP")
	ErrCantPersist     = errors.New(runtime.GOOS + " can't persist")
	ErrCantChangeOwner = errors.New(runtime.GOOS + " can't change owner")
	ErrCantChangeGroup = errors.New(runtime.GOOS + " can't change group")
)

var (
	namsiz  = uintptr(IFNAMSIZ)
	namsizp = uintptr(unsafe.Pointer(&namsiz))
)

func New(
	unit uint,
	isTAP bool,
	persist bool,
	owner, group int,
	ha netif.HardwareAddr,
) (t *os.File, err error) {
	if isTAP {
		err = ErrCantTAP
	} else if persist {
		err = ErrCantPersist
	} else if owner != Unset {
		err = ErrCantChangeOwner
	} else if group != Unset {
		err = ErrCantChangeGroup
	} else if err != nil {
		return
	}

	fd, err := syscall.
		Socket(PF_SYSTEM, syscall.SOCK_DGRAM, SYSPROTO_CONTROL)
	if err != nil {
		return nil, fmt.Errorf("socket: %v", err)
	}
	defer func() {
		if err != nil {
			syscall.Close(fd)
		}
	}()
	defer egress.Recovery(&err)

	ci := newCtlInfo()
	cip := uintptr(unsafe.Pointer(ci))

	if err = ioctl(uintptr(fd), CTLIOCGINFO, cip); err != nil {
		panic(err)
	}

	sac := &SockaddrCtl{
		Sc_len:     SizeofSockaddrCtl,
		Sc_family:  AF_SYSTEM,
		Ss_sysaddr: AF_SYS_CONTROL,
		Sc_id:      ci.Id,
		Sc_unit:    uint32(unit) + 1,
	}
	sacp := uintptr(unsafe.Pointer(sac))
	_, _, errno := syscall.RawSyscall(syscall.SYS_CONNECT, uintptr(fd),
		sacp, SizeofSockaddrCtl)
	if errno != 0 {
		panic(os.NewSyscallError("connect", errno))
	}

	name := make([]byte, IFNAMSIZ, IFNAMSIZ)
	namep := uintptr(unsafe.Pointer(&name[0]))

	_, _, errno = syscall.Syscall6(syscall.SYS_GETSOCKOPT, uintptr(fd),
		SYSPROTO_CONTROL, UTUN_OPT_IFNAME, namep, namsizp, 0)
	if errno != 0 {
		panic(os.NewSyscallError("ifname", errno))
	}

	for i, b := range name {
		if b == 0 {
			name = name[:i]
			break
		}
	}

	if err = syscall.SetNonblock(fd, true); err != nil {
		panic(err)
	}

	t = os.NewFile(uintptr(fd), string(name))
	return
}

func newCtlInfo() *CtlInfo {
	ci := new(CtlInfo)
	for i, b := range []byte(UTUN_CONTROL_NAME) {
		ci.Name[i] = int8(b)
	}
	return ci
}
