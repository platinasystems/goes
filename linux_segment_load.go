// package kexec contains the kexec-related syscalls
// +build linux

package kexec

import (
	"syscall"
	"unsafe"
)

type KexecSegment struct {
	Buf   *byte
	Bufsz uint
	Mem   uintptr
	Memsz uint
}

func SegmentLoad(entry uint64, segments *[]KexecSegment, flags uintptr) (err error) {
	_, _, e := syscall.Syscall6(syscall.SYS_KEXEC_LOAD, uintptr(entry),
		uintptr(len(*segments)),
		uintptr(unsafe.Pointer(&((*segments)[0].Buf))),
		uintptr(flags), uintptr(0), uintptr(0))
	err = nil
	if e != 0 {
		err = e
	}
	return
}
