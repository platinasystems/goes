// Copyright © 2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package flash_eraseall

import (
	"errors"
	"fmt"
	"github.com/platinasystems/goes/lang"
	"os"
	"syscall"
	"unsafe"
)

type Command struct{}

func (Command) String() string { return "flash_eraseall" }

func (Command) Usage() string {
	return "flash_eraseall [MTD device]"
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "erase a MTD device",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
DESCRIPTION
	The flash_eraseall command erases the specified MTD device.`,
	}
}

type MTDinfo struct {
	typ       byte
	flags     uint32
	size      uint32
	erasesize uint32
	writesize uint32
	oobsize   uint32
	unused    uint64
}

type EraseInfo struct {
	start  uint32
	length uint32
}

const (
	MEMGETINFO = 0x80204d01
	MEMERASE   = 0x40084d02
)

func (c Command) Main(args ...string) (err error) {
	if len(args) != 1 {
		return errors.New(c.Usage())
	}

	m, err := os.OpenFile(args[0], os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("Unable to open %s: %w", args[0], err)
	}
	mi := &MTDinfo{}
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, m.Fd(),
		uintptr(MEMGETINFO), uintptr(unsafe.Pointer(mi)))
	if e != 0 {
		return fmt.Errorf("Error getting info on %s: %w", args[0], err)
	}
	ei := &EraseInfo{0, mi.erasesize}
	for ei.start = 0; ei.start < mi.size; ei.start += mi.erasesize {
		fmt.Println("Erasing Block...", ei.start, ei.length)
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, m.Fd(),
			uintptr(MEMERASE), uintptr(unsafe.Pointer(ei)))
		if e != 0 {
			return fmt.Errorf("Erase error block %d: %s", ei.start,
				e)
		}
	}

	return nil
}
