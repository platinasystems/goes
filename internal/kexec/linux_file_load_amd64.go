// package kexec contains the kexec-related syscalls
// +build linux,amd64

package kexec

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/log"
)

func fileLoadSyscall(k *os.File, i *os.File, cmdline string, flags uintptr) (err error) {
	c, err := syscall.BytePtrFromString(cmdline)
	if err != nil {
		return err
	}
	log.Printf("kexec command line", "%s", cmdline)
	_, _, e := syscall.Syscall6(320,
		k.Fd(), i.Fd(), uintptr(len(cmdline)+1),
		uintptr(unsafe.Pointer(c)), flags, uintptr(0))

	err = nil
	if e != 0 {
		err = e
	}
	return err
}
