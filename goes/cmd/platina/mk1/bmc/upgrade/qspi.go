package upgrade

import (
	"fmt"
	"syscall"
	"unsafe"
)

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
	MEMGETINFO = 0x80204d01 //from linux: mtd-abi.h
	MEMERASE   = 0x40084d02
	MEMLOCK    = 0x40084d05
	MEMUNLOCK  = 0x40084d06
	MEMERASE64 = 0x40104d14
	MTDdevice  = "/dev/mtd0"
)

var mi = &MTDinfo{0, 0, 0, 0, 0, 0, 0}
var ei = &EraseInfo{0, 0}
var fd int = 0

func openQSPI() (err error) {
	path := MTDdevice
	defer syscall.Close(fd)
	fd, err = syscall.Open(path, syscall.O_RDWR, 0)
	if err != nil {
		err = fmt.Errorf("open %s: %s", path, err)
		return err
	}
	_, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(MEMGETINFO), uintptr(unsafe.Pointer(mi)))
	if e != 0 {
		err = fmt.Errorf("Open %s: %s", path, e)
		return err
	}
	return nil
}

func closeQSPI() {
	syscall.Close(fd)
}

func readQSPI(offset int, length int) ([]byte, error) {
	path := MTDdevice
	buf := make([]byte, length)
	defer syscall.Close(fd)
	_, err := syscall.Seek(fd, int64(offset), 0)
	if err != nil {
		err = fmt.Errorf("Seek error: %s: %s", path, err)
		return buf, err
	}
	n, err := syscall.Read(fd, buf)
	if err != nil {
		err = fmt.Errorf("Read error %s: %s", path, err)
		return buf, err
	}
	fmt.Println(n, string(buf))
	return buf, nil
}

func writeQSPI(offset int, length int, b []byte) error {
	path := MTDdevice
	defer syscall.Close(fd)
	_, err := syscall.Seek(fd, int64(offset), 0)
	if err != nil {
		err = fmt.Errorf("Seek error: %s: %s", path, err)
		return err
	}
	buf := make([]byte, length)
	buf = b
	n, err := syscall.Write(fd, buf)
	if err != nil {
		err = fmt.Errorf("Write error %s: %s", path, err)
		return err
	}
	fmt.Println(n, string(buf))
	//TODO: add verify
	return nil
}

func eraseQSPI(offset int, length int) error {
	defer syscall.Close(fd)
	ei.length = mi.erasesize
	for i := 0; i < 1; i += 1 { //ei.length
		//ioctl(fd, MEMUNLOCK, &ei)
		//log.Print("Erasing Block %x\n", ei.start)
		//ioctl(fd, MEMERASE, &ei)
	}
	//TODO: add verify
	return nil
}
