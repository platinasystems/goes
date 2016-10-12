// package fit is for parsing flattened image tree binaries
package fit

import (
	"syscall"
	"unsafe"
	"fmt"
)

type kexecSegment struct {
	buf   *byte
	bufsz uint
	mem   uintptr
	memsz uint
}

func rebootSyscall(cmd int) (err error) {
	_, _, e := syscall.Syscall6(syscall.SYS_REBOOT,
		syscall.LINUX_REBOOT_MAGIC1,
		syscall.LINUX_REBOOT_MAGIC2, uintptr(cmd), 0, 0, 0)
	err = nil
	if e != 0 {
		err = e
	}
	return
}

func KexecRebootSyscall() (err error) {
	return rebootSyscall(syscall.LINUX_REBOOT_CMD_KEXEC)
}

func (f *Fit) kexecLoadSyscall(entry uint64, segments *[]kexecSegment, flags uintptr) (err error) {
	if (f.Debug) {
		fmt.Printf("Segment count %d\n", len(*segments))
	}
	_, _, e := syscall.Syscall6(syscall.SYS_KEXEC_LOAD, uintptr(entry),
		uintptr(len(*segments)),
		uintptr(unsafe.Pointer(&((*segments)[0].buf))),
		uintptr(flags), uintptr(0), uintptr(0))
	err = nil
	if e != 0 {
		err = e
	}
	return
}

func (f *Fit) KexecLoadConfig(conf *Config, offset uintptr) (err error) {
	var segments []kexecSegment

	segments = make([]kexecSegment, len(conf.ImageList),
		len(conf.ImageList))

	for i, image := range conf.ImageList {
		segments[i].buf = &image.Data[0]
		bufsz := uint(len(image.Data))
		segments[i].bufsz = bufsz
		memsz := (bufsz + 4095) &^ 4095
		segments[i].mem = uintptr(image.LoadAddr) + offset
		segments[i].memsz = memsz
	}
	if (f.Debug) {
		fmt.Printf("Segments: %+v\n", segments)
	}
	err = f.kexecLoadSyscall(conf.BaseAddr, &segments, uintptr(0))

	return
}
