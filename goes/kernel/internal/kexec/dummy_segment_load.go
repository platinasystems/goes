// package kexec contains the kexec-related syscalls
// +build !linux

package kexec

import (
	"syscall"
)

func SegmentLoad(entry uint64, segments *[]KexecSegment, flags uintptr) (err error) {
	err = syscall.ENOSYS
	return
}
