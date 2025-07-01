// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nettun

import (
	"context"
	"os"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"golang.org/x/sys/unix"
)

//go:generate sh -c "go tool cgo -godefs -- if_utun_darwin.go > zif_utun_darwin.go"

const (
	HasPI          = true
	CanTAP         = false
	CanPersist     = false
	CanChangeOwner = false
	CanChangeGroup = false
)

var namsiz = uintptr(IFNAMSIZ)

func New(
	unit int,
	isTAP bool,
	persist bool,
	owner, group int,
	ha netif.HardwareAddr,
) (*os.File, error) {
	if isTAP {
		return nil, xerrors.Unsupported("tap")
	} else if persist {
		return nil, xerrors.Unsupported("persist")
	} else if owner != Unset {
		return nil, xerrors.Unsupported("change-owner")
	} else if group != Unset {
		return nil, xerrors.Unsupported("change-group")
	}

	fd, err := unix.Socket(PF_SYSTEM, unix.SOCK_DGRAM, SYSPROTO_CONTROL)
	if err != nil {
		return nil, xerrors.Mark(err)
	}
	defer func() {
		if err != nil {
			unix.Close(fd)
		}
	}()

	ci := new(CtlInfo)
	for i, b := range []byte(UTUN_CONTROL_NAME) {
		ci.Name[i] = int8(b)
	}

	err = ioctl(uintptr(fd), CTLIOCGINFO, uintptr(unsafe.Pointer(ci)))
	if err != nil {
		return nil, err
	}

	if unit < 0 {
		ctx := context.Background()
		unit, _ = netif.NextUnit(ctx, "utun")
	}
	sac := &SockaddrCtl{
		Sc_len:     SizeofSockaddrCtl,
		Sc_family:  AF_SYSTEM,
		Ss_sysaddr: AF_SYS_CONTROL,
		Sc_id:      ci.Id,
		Sc_unit:    uint32(unit) + 1,
	}
	_, _, errno := unix.RawSyscall(unix.SYS_CONNECT, uintptr(fd),
		uintptr(unsafe.Pointer(sac)), SizeofSockaddrCtl)
	if errno != 0 {
		err = os.NewSyscallError("sysctl", errno)
		return nil, xerrors.Mark(err)
	}

	name := make([]byte, IFNAMSIZ, IFNAMSIZ)
	_, _, errno = unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(fd),
		SYSPROTO_CONTROL, UTUN_OPT_IFNAME,
		uintptr(unsafe.Pointer(&name[0])),
		uintptr(unsafe.Pointer(&namsiz)),
		0)
	if errno != 0 {
		err = os.NewSyscallError("ifname", errno)
		return nil, err
	}

	for i, b := range name {
		if b == 0 {
			name = name[:i]
			break
		}
	}

	if err = unix.SetNonblock(fd, true); err != nil {
		return nil, err
	}

	return os.NewFile(uintptr(fd), string(name)), nil
}
