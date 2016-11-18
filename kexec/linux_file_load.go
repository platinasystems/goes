// package kexec contains the kexec-related syscalls
// +build linux

package kexec

import (
	"os"
)

func FileLoad(k *os.File, i *os.File, cmdline string, flags uintptr) (err error) {
	return fileLoadSyscall(k, i, cmdline, flags)
}
