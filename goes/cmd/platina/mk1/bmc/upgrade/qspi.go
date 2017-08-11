// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package upgrade

import (
	"fmt"
	"io/ioutil"
	"os"
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

var img = []string{"ubo", "dtb", "env", "ker", "ini"}
var off = []uint32{0x00000, 0x80000, 0xc0000, 0x100000, 0x300000}
var siz = []uint32{0x80000, 0x40000, 0x40000, 0x200000, 0x300000}

func writeImageAll() (err error) {
	fd, err = syscall.Open(MTDdevice, syscall.O_RDWR, 0)
	if err != nil {
		err = fmt.Errorf("Open error %s: %s", MTDdevice, err)
		return err
	}
	defer syscall.Close(fd)

	if err = infoQSPI(); err != err {
		return err
	}
	for j, i := range img {
		err := writeImage("/"+Machine+"-"+i+".bin", off[j], siz[j])
		if err != nil {
			return err
		}
	}
	return nil
}

func writeImage(im string, of uint32, sz uint32) error {
	if fi, err := os.Stat(im); !os.IsNotExist(err) {
		if fi.Size() < 1000 {
			fmt.Println("skipping file...", im)
			return nil
		}
		if err = eraseQSPI(of, sz); err != nil {
			return err
		}
		b, err := ioutil.ReadFile(im)
		if err != nil {
			return err
		}
		fmt.Println("Programming...", im)
		_, err = writeQSPI(b, of)
		if err != nil {
			return err
		}
		//TODO add verify
	}
	return nil
}

var mi = &MTDinfo{0, 0, 0, 0, 0, 0, 0}
var ei = &EraseInfo{0, 0}
var fd int = 0

func infoQSPI() (err error) {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(MEMGETINFO), uintptr(unsafe.Pointer(mi)))
	if e != 0 {
		err = fmt.Errorf("Open error : %s", e)
		return err
	}
	return nil
}

func readQSPI(of uint32, sz uint32) (int, []byte, error) {
	b := make([]byte, sz)
	_, err := syscall.Seek(fd, int64(of), 0)
	if err != nil {
		err = fmt.Errorf("Seek error: %s: %s", of, err)
		return 0, b, err
	}
	n, err := syscall.Read(fd, b)
	if err != nil {
		err = fmt.Errorf("Read error %s: %s", of, err)
		return 0, b, err
	}
	fmt.Println(n, string(b))
	return n, b, nil
}

func writeQSPI(b []byte, of uint32) (int, error) {
	_, err := syscall.Seek(fd, int64(of), 0)
	if err != nil {
		err = fmt.Errorf("Seek error: %s: %s", of, err)
		return 0, err
	}
	n, err := syscall.Write(fd, b)
	if err != nil {
		err = fmt.Errorf("Write error %s: %s", of, err)
		return 0, err
	}
	return n, nil
}

func eraseQSPI(of uint32, sz uint32) error {
	ei.length = mi.erasesize
	end := of + sz
	for ei.start = of; ei.start < end; ei.start += ei.length {
		fmt.Println("Erasing Block...", ei.start)
		_, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
			uintptr(MEMERASE), uintptr(unsafe.Pointer(ei)))
		if e != 0 {
			err := fmt.Errorf("Erase error %s: %s", ei.start, e)
			return err
		}
	}
	return nil
}
