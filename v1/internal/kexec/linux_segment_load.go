// package kexec contains the kexec-related syscalls
// +build linux

package kexec

import (
	"fmt"
	"syscall"
	"unsafe"
)

type KexecSegment struct {
	Buf   *byte
	Bufsz uint
	Mem   uintptr
	Memsz uint
}

func (s KexecSegment) String() string {
	return fmt.Sprintf("Buf@%xx Bufsize=%x Mem=%x Memsz=%x", s.Buf, s.Bufsz,
		s.Mem, s.Memsz)
}

func SliceAddSegment(s []KexecSegment, b *[]byte, a uintptr) []KexecSegment {
	var ks KexecSegment
	ks.Buf = &(*b)[0]
	bufsz := uint(len(*b))
	ks.Bufsz = bufsz
	memsz := (bufsz + 4095) &^ 4095
	ks.Mem = a
	ks.Memsz = memsz
	s = append(s, ks)

	return s
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
