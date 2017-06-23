// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build vfio

package pci

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/elib/iomux"

	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"sync"
	"syscall"
	"unsafe"
)

type vfio_group struct {
	// Group number.
	number uint

	// /dev/vfio/GROUP_NUMBER
	fd int

	status vfio_group_status

	devices []*vfio_pci_device
}

type vfio_pci_device struct {
	Device

	m     *vfio_main
	group *vfio_group

	info      vfio_device_info
	irq_infos []vfio_irq_info

	// device fd from VFIO_GROUP_GET_DEVICE_FD
	iomux.File
}

type vfio_main struct {
	api_version int

	// /dev/vfio/vfio
	container_fd int

	iommu_info vfio_iommu_type1_info
	dma_map    vfio_iommu_type1_dma_map

	// Groups indexed by iommu group number.
	groups_by_number map[uint]*vfio_group

	devices []*vfio_pci_device

	// Chunks are 2^log2LinesPerChunk cache lines long.
	// Kernel gives us memory in "Chunks" which are physically contiguous.
	log2LinesPerChunk, log2BytesPerChunk uint8

	once sync.Once
}

var default_vfio_main = &vfio_main{}

func vfio_ioctl(fd, call int, arg uintptr) (r uintptr, err error) {
	r, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(call), arg)
	if e != 0 {
		err = os.NewSyscallError("ioctl", e)
	}
	return
}

func (m *vfio_main) ioctl(call int, arg uintptr) (uintptr, error) {
	return vfio_ioctl(m.container_fd, call, arg)
}
func (x *vfio_group) ioctl(call int, arg uintptr) (uintptr, error) {
	return vfio_ioctl(x.fd, call, arg)
}
func (x *vfio_pci_device) ioctl(call int, arg uintptr) (uintptr, error) {
	return vfio_ioctl(x.File.Fd, call, arg)
}

func (m *vfio_main) init(dma_heap_bytes uint) (err error) {
	m.container_fd, err = syscall.Open("/dev/vfio/vfio", syscall.O_RDWR, 0)
	if err != nil {
		return
	}
	defer func() {
		if err != nil && m.container_fd != 0 {
			syscall.Close(m.container_fd)
		}
	}()

	{
		var v uintptr
		if v, err = m.ioctl(vfio_get_api_version, 0); err != nil {
			return
		}
		m.api_version = int(v)

		if v, err = m.ioctl(vfio_check_extension, vfio_type1_iommu); v == 0 || err != nil {
			if err == nil && v == 0 {
				err = errors.New("vfio type 1 iommu not supported by kernel")
			}
			return
		}
	}

	// Enable the IOMMU model we want.
	if _, err = m.ioctl(vfio_set_iommu, vfio_type1_iommu); err != nil {
		return
	}

	// Fetch iommu info.  Supported page sizes.
	m.iommu_info.set_size(unsafe.Sizeof(m.iommu_info))
	if _, err = m.ioctl(vfio_iommu_get_info, uintptr(unsafe.Pointer(&m.iommu_info))); err != nil {
		return
	}

	{
		addr, data, e := elib.RawMmap(0, uintptr(dma_heap_bytes),
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS,
			0, 0)
		if e != nil {
			err = e
			return
		}
		m.dma_map = vfio_iommu_type1_dma_map{
			vaddr: uint64(addr),
			size:  uint64(dma_heap_bytes),
		}
		m.dma_map.set(unsafe.Sizeof(m.dma_map), vfio_dma_map_flag_read|vfio_dma_map_flag_write)
		if _, err = m.ioctl(vfio_iommu_map_dma, uintptr(unsafe.Pointer(&m.dma_map))); err != nil {
			return
		}

		hw.DmaInit(data)
	}

	return err
}

func sysfsWrite(path, format string, args ...interface{}) error {
	fn := "/sys/bus/pci/drivers/vfio-pci/" + path
	f, err := os.OpenFile(fn, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, format, args...)
	return err
}

func (d *vfio_pci_device) new_id() (err error) {
	err = sysfsWrite("new_id", "%04x %04x", int(d.VendorID()), int(d.DeviceID()))
	if err != nil {
		return
	}
	return
}

func (d *vfio_pci_device) remove_id() (err error) {
	err = sysfsWrite("remove_id", "%04x %04x", int(d.VendorID()), int(d.DeviceID()))
	return
}

func NewDevice() Devicer {
	d := &vfio_pci_device{m: default_vfio_main}
	d.Device.Devicer = d
	return d
}

func (d *vfio_pci_device) GetDevice() *Device { return &d.Device }

func (d *vfio_pci_device) sysfs_get_group_number() (uint, error) {
	s, err := os.Readlink("/sys/bus/pci/devices/" + d.Device.Addr.String() + "/iommu_group")
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseUint(path.Base(s), 10, 0)
	return uint(n), err
}

func (m *vfio_main) new_group(group_number uint) (g *vfio_group, err error) {
	group_path := fmt.Sprintf("/dev/vfio/%d", group_number)
	var fd int
	fd, err = syscall.Open(group_path, syscall.O_RDWR, 0)
	if err != nil {
		err = os.NewSyscallError("open /dev/vfio/GROUP", err)
		return
	}

	defer func() {
		if err != nil && fd >= 0 {
			syscall.Close(fd)
			g = nil
		}
	}()

	g = &vfio_group{number: group_number, fd: fd}
	g.status.set_size(unsafe.Sizeof(g.status))

	if _, err = g.ioctl(vfio_group_get_status, uintptr(unsafe.Pointer(&g.status.vfio_ioctl_common))); err != nil {
		return
	}
	// Group must be viable.
	if g.status.flags&vfio_group_flags_viable == 0 {
		err = fmt.Errorf("vfio group %d is not viable (not all devices are bound for vfio)", g.number)
		return
	}
	if m.groups_by_number == nil {
		m.groups_by_number = make(map[uint]*vfio_group)
	}
	m.groups_by_number[group_number] = g

	return
}

func (d *vfio_pci_device) find_group() (g *vfio_group, err error) {
	var (
		n  uint
		ok bool
	)
	g = d.group
	if g != nil {
		return
	}
	if n, err = d.sysfs_get_group_number(); err != nil {
		return
	}
	if g, ok = d.m.groups_by_number[n]; !ok {
		g, err = d.m.new_group(n)
		if err != nil {
			return
		}
	}
	d.group = g
	g.devices = append(g.devices, d)
	return
}

func (d *vfio_pci_device) Open() (err error) {
	err = d.new_id()
	if err != nil {
		return
	}

	// Make sure group exists and is viable.
	if _, err = d.find_group(); err != nil {
		return
	}

	// Initialize DMA heap once device is open.
	d.m.once.Do(func() {
		err = d.m.init(64 << 20)
	})
	if err != nil {
		panic(err)
	}

	// Set group container.
	if d.group.status.flags&vfio_group_flags_container_set == 0 {
		if _, err = vfio_ioctl(d.group.fd, vfio_group_set_container, uintptr(d.m.container_fd)); err != nil {
			return
		}
		d.group.status.flags |= vfio_group_flags_container_set
	}

	// Get device fd.
	{
		tmp := []byte(d.Device.Addr.String())
		var fd uintptr
		if fd, err = d.group.ioctl(vfio_group_get_device_fd, uintptr(unsafe.Pointer(&tmp[0]))); err != nil {
			return
		}
		d.File.Fd = int(fd)
	}

	// Fetch device info.
	d.info.set_size(unsafe.Sizeof(d.info))
	if _, err = d.ioctl(vfio_device_get_info, uintptr(unsafe.Pointer(&d.info))); err != nil {
		return
	}

	// Fetch interrupt infos for each interrupt.
	d.irq_infos = make([]vfio_irq_info, d.info.num_irqs)
	for i := range d.irq_infos {
		ii := &d.irq_infos[i]
		ii.set_size(unsafe.Sizeof(*ii))
		if _, err = d.ioctl(vfio_device_get_irq_info, uintptr(unsafe.Pointer(ii))); err != nil {
			return
		}
	}

	panic("za")

	// Listen for interrupts.
	iomux.Add(d)

	return
}
func (d *vfio_pci_device) Close() (err error) {
	d.remove_id()
	return
}

var errShouldNeverHappen = errors.New("should never happen")

func (d *vfio_pci_device) ErrorReady() error    { return errShouldNeverHappen }
func (d *vfio_pci_device) WriteReady() error    { return errShouldNeverHappen }
func (d *vfio_pci_device) WriteAvailable() bool { return false }
func (d *vfio_pci_device) String() string       { return "pci " + d.Device.String() }

// UIO file is ready when interrupt occurs.
func (d *vfio_pci_device) ReadReady() (err error) {
	d.DriverDevice.Interrupt()
	return
}
