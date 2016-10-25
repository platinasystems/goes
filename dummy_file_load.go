// package kexec contains the kexec-related syscalls
// +build !linux, linux,!amd64

package kexec

import (
	"os"
	"syscall"
)

func FileLoad(k *os.File, i *os.File, cmdline string, flags uintptr) (err error) {
	err = syscall.ENOSYS
	return
}
