// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build uio_pci_generic

package pci

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/iomux"

	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"syscall"
)

type uio_pci_generic_main struct {
	busCommon

	// Chunks are 2^log2LinesPerChunk cache lines long.
	// Kernel gives us memory in "Chunks" which are physically contiguous.
	log2LinesPerChunk, log2BytesPerChunk uint8

	once sync.Once
}

type uio_pci_generic_device struct {
	Device

	// /dev/uioN
	iomux.File

	index uint32

	uio_minor uint32
}

func sysfsWrite(path, format string, args ...interface{}) error {
	fn := "/sys/bus/pci/drivers/uio_pci_generic/" + path
	f, err := os.OpenFile(fn, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, format, args...)
	return err
}

func (d *uio_pci_generic_device) bind() (err error) {
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
		if _, err = fmt.Sscanf(fi.Name(), "uio%d", &d.uio_minor); err == nil {
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

func (d *uio_pci_generic_device) unbind() (err error) {
	err = sysfsWrite("unbind", "%s", &d.Addr)
	if err != nil {
		return
	}
	err = sysfsWrite("remove_id", "%04x %04x", int(d.VendorID()), int(d.DeviceID()))
	return
}

var DefaultBus = &uio_pci_generic_main{}

func (d *uio_pci_generic_device) GetDevice() *Device  { return &d.Device }
func (m *uio_pci_generic_main) NewDevice() BusDevice  { return &uio_pci_generic_device{} }
func (m *uio_pci_generic_main) Validate() (err error) { return }

func (m *uio_pci_generic_main) heap_init(minor uint32, log2_dma_heap_bytes uint) (err error) {
	t := &hw.PageTable
	n_dma_heap_bytes := uintptr(1) << log2_dma_heap_bytes
	var data []byte
	t.Data, data, err = elib.MmapSlice(0, n_dma_heap_bytes,
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_ANONYMOUS|syscall.MAP_HUGETLB|syscall.MAP_LOCKED,
		^uintptr(0), 0)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			elib.Munmap(data)
		}
	}()
	hw.DmaInit(data)

	var f *os.File
	f, err = os.OpenFile("/proc/self/pagemap", os.O_RDONLY, 0)
	if err != nil {
		return
	}
	defer f.Close()

	const (
		log2_page_size      = 12
		log2_huge_page_size = log2_page_size + 9
		page_size           = 1 << log2_page_size
		huge_page_size      = 1 << log2_huge_page_size
	)

	t.Log2BytesPerPage = huge_page_size
	t.Pages = make([]uintptr, n_dma_heap_bytes/huge_page_size)

	for i := range t.Pages {
		var (
			v uint64
			b [8]byte
		)
		a := t.Data + uintptr(i)*huge_page_size
		pfn := int64(a) / page_size
		if _, err = f.Seek(pfn*8, os.SEEK_SET); err != nil {
			return
		}
		if _, err = f.Read(b[:]); err != nil {
			return
		}
		v = binary.LittleEndian.Uint64(b[:])

		// Bits 0-54 are the physical page number.
		phys_address := (v & (1<<54 - 1)) * page_size
		t.Pages[i] = uintptr(phys_address)
	}
	return
}

func (d *uio_pci_generic_device) Open() (err error) {
	err = d.bind()
	if err != nil {
		return
	}

	uioPath := fmt.Sprintf("/dev/uio%d", d.uio_minor)
	d.File.Fd, err = syscall.Open(uioPath, syscall.O_RDONLY, 0)
	if err != nil {
		panic(fmt.Errorf("open %s: %s", uioPath, err))
	}

	// Initialize DMA heap once device is open.
	m := DefaultBus
	m.once.Do(func() {
		err = m.heap_init(d.uio_minor, 28)
	})
	if err != nil {
		panic(err)
	}

	// Listen for interrupts.
	iomux.Add(d)

	return
}
func (d *uio_pci_generic_device) Close() (err error) {
	return d.unbind()
}

var errShouldNeverHappen = errors.New("should never happen")

func (d *uio_pci_generic_device) ErrorReady() error    { return errShouldNeverHappen }
func (d *uio_pci_generic_device) WriteReady() error    { return errShouldNeverHappen }
func (d *uio_pci_generic_device) WriteAvailable() bool { return false }
func (d *uio_pci_generic_device) String() string       { return "pci " + d.Device.String() }

// UIO file is ready when interrupt occurs.
func (d *uio_pci_generic_device) ReadReady() (err error) {
	var b [4]byte
	if _, err = syscall.Read(d.File.Fd, b[:]); err != nil {
		return
	}
	d.DriverDevice.Interrupt()
	return
}
