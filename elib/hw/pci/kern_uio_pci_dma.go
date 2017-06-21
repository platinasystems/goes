// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build uio_pci_dma

package pci

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/iomux"

	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"syscall"
	"unsafe"
)

type uioPciDmaMain struct {
	// /dev/uio-dma
	uio_dma_fd int

	// Chunks are 2^log2LinesPerChunk cache lines long.
	// Kernel gives us memory in "Chunks" which are physically contiguous.
	log2LinesPerChunk, log2BytesPerChunk uint8

	once sync.Once
}

const (
	uio_dma_cache_default = iota
	uio_dma_cache_disable
	uio_dma_cache_writecombine
)

const (
	uio_dma_bidirectional = iota
	uio_dma_todevice
	uio_dma_fromdevice
)

const (
	uio_dma_alloc = 0x400455c8
	uio_dma_free  = 0x400455c9
	uio_dma_map   = 0x400455ca
	uio_dma_unmap = 0x400455cb
)

type uio_dma_alloc_req struct {
	dma_mask    uint64
	memnode     uint16
	cache       uint16
	flags       uint32
	chunk_count uint32
	chunk_size  uint32
	mmap_offset uint64
}

type uio_dma_free_req struct {
	mmap_offset uint64
}

type uio_dma_map_req struct {
	mmap_offset uint64
	flags       uint32
	devid       uint32
	direction   uint32
	chunk_count uint32
	chunk_size  uint32
	dma_addr    [256]uint64
}

type uio_dma_unmap_req struct {
	mmap_offset uint64
	devid       uint32
	flags       uint32
	direction   uint32
}

func (h *uioPciDmaMain) alloc_and_map(mmap_offset, size uint64, uioMinorDevice uint32) (m uio_dma_map_req, err error) {
	r := uio_dma_alloc_req{}
	r.dma_mask = 0xffffffff

	r.chunk_size = uint32(size)
	r.mmap_offset = mmap_offset
	r.cache = uio_dma_cache_writecombine
	for {
		r.chunk_count = uint32(size) / r.chunk_size
		_, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(h.uio_dma_fd), uintptr(uio_dma_alloc), uintptr(unsafe.Pointer(&r)))
		if e == 0 {
			break
		}
		if r.chunk_size == 4<<10 {
			err = fmt.Errorf("ioctl UIO_DMA_ALLOC fails: %s", e)
			return
		}
		r.chunk_size /= 2
	}

	m.direction = uio_dma_bidirectional
	m.chunk_size = r.chunk_size
	m.chunk_count = r.chunk_count
	m.mmap_offset = r.mmap_offset
	m.devid = uioMinorDevice
	_, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(h.uio_dma_fd), uintptr(uio_dma_map), uintptr(unsafe.Pointer(&m)))
	if e != 0 {
		fr := uio_dma_free_req{mmap_offset: r.mmap_offset}
		_, _, f := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(h.uio_dma_fd), uintptr(uio_dma_free), uintptr(unsafe.Pointer(&fr)))
		err = fmt.Errorf("uio-dma-map: %s", e)
		if f != 0 {
			err = fmt.Errorf("%s, uio-dma-free: %s", err, f)
		}
		return
	}

	return
}

func (h *uioPciDmaMain) mmap(addr, length, prot, flags, fd, offset uintptr) (a uintptr, b []byte, err error) {
	r, _, e := syscall.RawSyscall6(syscall.SYS_MMAP, addr, length, prot, flags, fd, offset)
	if e != 0 {
		err = fmt.Errorf("uio-dma mmap: %s", e)
		return
	}
	slice := reflect.SliceHeader{Data: r, Len: int(length), Cap: int(length)}
	a = r
	b = *(*[]byte)(unsafe.Pointer(&slice))
	return
}

func (h *uioPciDmaMain) heapInit(uioMinorDevice uint32, maxSize uint64) (err error) {
	h.uio_dma_fd, err = syscall.Open("/dev/uio-dma", syscall.O_RDWR, 0)
	if err != nil {
		return
	}
	defer func() {
		if err != nil && h.uio_dma_fd != 0 {
			syscall.Close(h.uio_dma_fd)
		}
	}()

	mmap_offset := uint64(0)
	r, err := h.alloc_and_map(mmap_offset, maxSize, uioMinorDevice)
	if err != nil {
		return err
	}

	t := &hw.PageTable
	var data []byte
	t.Data, data, err = elib.RawMmap(0, uintptr(maxSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED,
		uintptr(h.uio_dma_fd), uintptr(mmap_offset))
	if err != nil {
		return err
	}
	hw.DmaInit(data)

	t.Log2BytesPerPage = uint(elib.MinLog2(elib.Word(r.chunk_size)))
	t.Pages = make([]uintptr, r.chunk_count)
	for i := range t.Pages {
		t.Pages[i] = uintptr(r.dma_addr[i])
	}

	return err
}

var uioPciDma = &uioPciDmaMain{}

type uioPciDevice struct {
	Device

	// /dev/uioN
	iomux.File

	index uint32

	uioMinorDevice uint32
}

func sysfsWrite(path, format string, args ...interface{}) error {
	fn := "/sys/bus/pci/drivers/uio_pci_dma/" + path
	f, err := os.OpenFile(fn, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, format, args...)
	return err
}

func (d *uioPciDevice) bind() (err error) {
	err = sysfsWrite("new_id", "%04x %04x", int(d.VendorID()), int(d.DeviceID()))
	if err != nil {
		return
	}

	err = sysfsWrite("bind", "%s", &d.Addr)
	if err != nil {
		return
	}

	var fis []os.FileInfo
	fis, err = ioutil.ReadDir(d.SysfsPath("uio"))
	if err != nil {
		return
	}

	ok := false
	for _, fi := range fis {
		if _, err = fmt.Sscanf(fi.Name(), "uio%d", &d.uioMinorDevice); err == nil {
			ok = true
			break
		}
	}
	if !ok {
		err = fmt.Errorf("failed to get minor number for uio device")
		return
	}

	return
}

func (d *uioPciDevice) unbind() (err error) {
	err = sysfsWrite("unbind", "%s", &d.Addr)
	if err != nil {
		return
	}
	err = sysfsWrite("remove_id", "%04x %04x", int(d.VendorID()), int(d.DeviceID()))
	return
}

func NewDevice() Devicer {
	d := &uioPciDevice{}
	d.Device.Devicer = d
	return d
}

func (d *uioPciDevice) GetDevice() *Device { return &d.Device }

func (d *uioPciDevice) Open() (err error) {
	err = d.bind()
	if err != nil {
		return
	}

	uioPath := fmt.Sprintf("/dev/uio%d", d.uioMinorDevice)
	d.File.Fd, err = syscall.Open(uioPath, syscall.O_RDONLY, 0)
	if err != nil {
		panic(fmt.Errorf("open %s: %s", uioPath, err))
	}

	// Initialize DMA heap once device is open.
	m := uioPciDma
	m.once.Do(func() {
		err = m.heapInit(d.uioMinorDevice, 64<<20)
	})
	if err != nil {
		panic(err)
	}

	// Listen for interrupts.
	iomux.Add(d)

	return
}
func (d *uioPciDevice) Close() (err error) {
	return
	// FIXME: Enabling unbind causes kernel crashes.  Not sure why.
	// return d.unbind()
}

var errShouldNeverHappen = errors.New("should never happen")

func (d *uioPciDevice) ErrorReady() error    { return errShouldNeverHappen }
func (d *uioPciDevice) WriteReady() error    { return errShouldNeverHappen }
func (d *uioPciDevice) WriteAvailable() bool { return false }
func (d *uioPciDevice) String() string       { return "pci " + d.Device.String() }

// UIO file is ready when interrupt occurs.
func (d *uioPciDevice) ReadReady() (err error) {
	d.DriverDevice.Interrupt()
	return
}
