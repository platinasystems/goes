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

type vfio_device_region_info struct {
	vfio_region_info
	sparse_areas []vfio_region_sparse_mmap_area
	cap_type     vfio_region_info_cap_type
}

type vfio_pci_device struct {
	Device

	m     *vfio_main
	group *vfio_group

	info         vfio_device_info
	region_infos []vfio_device_region_info

	irq_infos []vfio_irq_info

	// device fd from VFIO_GROUP_GET_DEVICE_FD
	device_fd int

	interrupt_event_fd int
	iomux.File
}

type vfio_main struct {
	busCommon

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

	container_init_once, dma_init_once sync.Once
}

func vfio_ioctl(fd int, kind vfio_ioctl_kind, arg uintptr) (r uintptr, err error) {
	r, _, e := syscall.RawSyscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(kind), arg)
	if e != 0 {
		err = os.NewSyscallError("ioctl "+kind.String(), e)
	}
	return
}

func (m *vfio_main) ioctl(kind vfio_ioctl_kind, arg uintptr) (uintptr, error) {
	return vfio_ioctl(m.container_fd, kind, arg)
}
func (x *vfio_group) ioctl(kind vfio_ioctl_kind, arg uintptr) (uintptr, error) {
	return vfio_ioctl(x.fd, kind, arg)
}
func (x *vfio_pci_device) ioctl(kind vfio_ioctl_kind, arg uintptr) (uintptr, error) {
	return vfio_ioctl(x.device_fd, kind, arg)
}

func (m *vfio_main) container_init() (err error) {
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

		if v, err = m.ioctl(vfio_check_extension, vfio_type1v2_iommu); v == 0 || err != nil {
			if err == nil && v == 0 {
				err = errors.New("vfio type 1 version 2 iommu not supported by kernel")
			}
			return
		}
	}

	return
}

func (m *vfio_main) dma_init(log2_dma_heap_bytes uint) (err error) {
	// Enable the IOMMU model we want.
	if _, err = m.ioctl(vfio_set_iommu, vfio_type1v2_iommu); err != nil {
		return
	}

	// Fetch iommu info.  Supported page sizes.
	m.iommu_info.set_size(unsafe.Sizeof(m.iommu_info))
	if _, err = m.ioctl(vfio_iommu_get_info, uintptr(unsafe.Pointer(&m.iommu_info))); err != nil {
		return
	}

	addr, data, e := elib.MmapSliceAligned(log2_dma_heap_bytes, hw.PhysmemLog2AddressAlign,
		syscall.MAP_SHARED|syscall.MAP_ANONYMOUS,
		syscall.PROT_READ|syscall.PROT_WRITE)
	if e != nil {
		err = e
		return
	}
	m.dma_map = vfio_iommu_type1_dma_map{
		vaddr: uint64(addr),
		iova:  uint64(hw.DmaPhysAddress(addr)),
		size:  uint64(1) << log2_dma_heap_bytes,
	}
	m.dma_map.set(unsafe.Sizeof(m.dma_map), vfio_dma_map_flag_read|vfio_dma_map_flag_write)
	if _, err = m.ioctl(vfio_iommu_map_dma, uintptr(unsafe.Pointer(&m.dma_map))); err != nil {
		return
	}

	hw.DmaInit(data)

	return
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

func (d *vfio_pci_device) sysfsWrite(name string) (err error) {
	err = sysfsWrite(name, "%04x %04x", int(d.VendorID()), int(d.DeviceID()))
	return
}

func (d *vfio_pci_device) new_id() error    { return d.sysfsWrite("new_id") }
func (d *vfio_pci_device) remove_id() error { return d.sysfsWrite("remove_id") }
func (d *vfio_pci_device) bind() error      { return d.sysfsWrite("bind") }
func (d *vfio_pci_device) unbind() error    { return d.sysfsWrite("unbind") }

var DefaultBus = &vfio_main{}

func (d *vfio_main) NewDevice() BusDevice { return &vfio_pci_device{m: d} }

// fixme should check that all iommu groups are fully populated.
func (d *vfio_main) Validate() (err error) { return }

func (d *vfio_pci_device) GetDevice() *Device { return &d.Device }

func (d *vfio_pci_device) sysfs_get_group_number() (uint, error) {
	s, err := os.Readlink("/sys/bus/pci/devices/" + d.Device.Addr.String() + "/iommu_group")
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseUint(path.Base(s), 10, 0)
	return uint(n), err
}

func (d *vfio_pci_device) new_group(group_number uint) (g *vfio_group, err error) {
	m := d.m
	group_path := fmt.Sprintf("/dev/vfio/%d", group_number)
	var fd int
	fd, err = syscall.Open(group_path, syscall.O_RDWR, 0)
	if err != nil {
		err = os.NewSyscallError("open "+group_path, err)
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
		g, err = d.new_group(n)
		if err != nil {
			return
		}
	}
	d.group = g
	g.devices = append(g.devices, d)
	return
}

func (d *vfio_pci_device) Open() (err error) {
	// Wrap error with device.
	defer func() {
		if err != nil {
			err = fmt.Errorf("pci %s: %s", d.Device.String(), err)
		}
	}()

	err = d.new_id()
	if err != nil {
		return
	}

	// Make sure group exists and is viable.
	if _, err = d.find_group(); err != nil {
		return
	}

	// Initialize DMA heap once device is open.
	d.m.container_init_once.Do(func() {
		err = d.m.container_init()
	})
	if err != nil {
		return
	}

	// Set group container.
	if d.group.status.flags&vfio_group_flags_container_set == 0 {
		if _, err = vfio_ioctl(d.group.fd, vfio_group_set_container, uintptr(unsafe.Pointer(&d.m.container_fd))); err != nil {
			return
		}
		d.group.status.flags |= vfio_group_flags_container_set
	}

	// Initialize DMA heap once at least one group has been added to container.
	d.m.dma_init_once.Do(func() {
		err = d.m.dma_init(28)
	})
	if err != nil {
		return
	}

	// Get device fd.
	{
		tmp := []byte(d.Device.Addr.String())
		var fd uintptr
		if fd, err = d.group.ioctl(vfio_group_get_device_fd, uintptr(unsafe.Pointer(&tmp[0]))); err != nil {
			return
		}
		d.device_fd = int(fd)
	}

	// Fetch device info.
	d.info.set_size(unsafe.Sizeof(d.info))
	if _, err = d.ioctl(vfio_device_get_info, uintptr(unsafe.Pointer(&d.info))); err != nil {
		return
	}

	// Fetch regions.
	d.region_infos = make([]vfio_device_region_info, d.info.num_regions)
	for i := range d.region_infos {
		type tmp struct {
			vfio_region_info
			caps [4 << 10]byte
		}
		x := &tmp{}
		x.set_size(unsafe.Sizeof(*x))
		x.index = uint32(i)
		if _, err = d.ioctl(vfio_device_get_region_info, uintptr(unsafe.Pointer(x))); err != nil {
			if i == vfio_pci_vga_region_index {
				// ignore vga region missing
				err = nil
			} else {
				return
			}
		}
		ri := &d.region_infos[i]
		ri.vfio_region_info = x.vfio_region_info
		if x.flags&vfio_region_info_flag_caps != 0 {
			o := x.cap_offset - uint32(unsafe.Sizeof(vfio_region_info{}))
			b := x.caps[o:]
			h, p := get_vfio_info_cap_header(b, 0)
			for h != nil {
				switch h.kind {
				case vfio_region_info_cap_kind_sparse_mmap:
					m := (*vfio_region_info_cap_sparse_mmap)(p)
					for i := uint32(0); i < m.nr_areas; i++ {
						a := m.get_area(i)
						ri.sparse_areas = append(ri.sparse_areas, *a)
					}
				case vfio_region_info_cap_kind_type:
					m := (*vfio_region_info_cap_type)(p)
					ri.cap_type = *m
				default:
					panic(fmt.Errorf("vfio region info unknown cap: %+v", h))
				}
				h, p = h.next(b)
			}
		}
	}

	// Fetch interrupt infos for each interrupt.
	d.irq_infos = make([]vfio_irq_info, d.info.num_irqs)
	for i := range d.irq_infos {
		x := &d.irq_infos[i]
		x.set_size(unsafe.Sizeof(*x))
		x.index = uint32(i)
		if _, err = d.ioctl(vfio_device_get_irq_info, uintptr(unsafe.Pointer(x))); err != nil {
			return
		}
	}

	// Set bus master.
	{
		c := (*ConfigHeader)(d.getRegs(0))
		v := c.Command.Get(&d.Device)
		v |= BusMasterEnable
		c.Command.Set(&d.Device, v)
	}

	// Get eventfd for interrupt.
	{
		r, _, e := syscall.RawSyscall(syscall.SYS_EVENTFD, 0, syscall.O_CLOEXEC|syscall.O_NONBLOCK, 0)
		if e != 0 {
			err = os.NewSyscallError("eventfd", e)
			return
		}
		d.interrupt_event_fd = int(r)
	}

	// Enable interrupt.
	{
		var ii *vfio_irq_info
		if ii = &d.irq_infos[vfio_pci_msi_irq_index]; ii.count == 0 {
			// No MSI? Choose first one.
			for i := range d.irq_infos {
				if d.irq_infos[i].count > 0 {
					ii = &d.irq_infos[i]
					break
				}
			}
		}
		if ii.count == 0 {
			panic("no irq")
		}
		type set struct {
			vfio_irq_set
			data [1]int32 // event fds
		}
		var s set
		s.set(unsafe.Sizeof(s), vfio_irq_set_data_eventfd|vfio_irq_set_action_trigger)
		s.index = ii.index
		s.start = uint32(0)
		s.count = uint32(len(s.data))
		s.data[0] = int32(d.interrupt_event_fd)
		if _, err = d.ioctl(vfio_device_set_irqs, uintptr(unsafe.Pointer(&s))); err != nil {
			return
		}
	}

	// Listen for interrupts.
	{
		d.Fd = int(d.interrupt_event_fd)
		iomux.Add(d)
	}

	// Reset device.
	if _, err = d.ioctl(vfio_device_reset, 0); err != nil {
		err = nil // ignore error
	}

	return
}

func (d *vfio_pci_device) Close() (err error) {
	if d.interrupt_event_fd > 0 {
		iomux.Del(d)
		syscall.Close(d.interrupt_event_fd)
	}
	if d.device_fd > 0 {
		syscall.Close(d.device_fd)
	}
	if err = d.unbind(); err != nil {
		return
	}
	err = d.remove_id()
	return
}

func (d *vfio_pci_device) MapResource(i uint) (res uintptr, err error) {
	r := &d.Device.Resources[i]
	if r.Index >= uint32(len(d.region_infos)) {
		err = fmt.Errorf("%s: mmap unknown resource BAR %d", d.Device.String(), r.Index)
		return
	}
	ri := &d.region_infos[r.Index]
	prot := uintptr(syscall.PROT_READ | syscall.PROT_WRITE)
	flags := uintptr(syscall.MAP_SHARED)
	if len(ri.sparse_areas) > 0 {
		var mem uintptr
		mem, r.Mem, err = elib.MmapSlice(0, uintptr(ri.size), uintptr(syscall.PROT_NONE), uintptr(syscall.MAP_SHARED|syscall.MAP_ANONYMOUS), 0, 0)
		if err == nil {
			for i := range ri.sparse_areas {
				a := &ri.sparse_areas[i]
				_, _, err = elib.MmapSlice(mem+uintptr(a.offset), uintptr(a.size), prot, flags|uintptr(syscall.MAP_FIXED),
					uintptr(d.device_fd), uintptr(ri.offset+a.offset))
				if err != nil {
					break
				}
			}
		}
	} else {
		_, r.Mem, err = elib.MmapSlice(0, uintptr(ri.size), prot, flags, uintptr(d.device_fd), uintptr(ri.offset))
	}
	if err != nil {
		err = fmt.Errorf("%s: mmap resource%d: %s", d.Device.String(), r.Index, err)
		return
	}
	res = uintptr(unsafe.Pointer(&r.Mem[0]))
	return
}

func (d *vfio_pci_device) region_rw(region, offset, vʹ, nBytes uint, isWrite bool) (v uint, err error) {
	f := os.NewFile(uintptr(d.device_fd), d.String())
	if _, err = f.Seek(int64(region)<<40|int64(offset), os.SEEK_SET); err != nil {
		return
	}
	var b [4]byte
	if isWrite {
		for i := range b {
			b[i] = byte((vʹ >> uint(8*i)) & 0xff)
		}
		_, err = f.Write(b[:nBytes])
		v = vʹ
	} else {
		_, err = f.Read(b[:nBytes])
		if err == nil {
			for i := range b {
				v |= uint(b[i]) << (8 * uint(i))
			}
		}
	}
	return
}
func (d *vfio_pci_device) ConfigRw(offset, v, nBytes uint, isWrite bool) uint {
	v, err := d.region_rw(vfio_pci_config_region_index, offset, v, nBytes, isWrite)
	if err != nil {
		panic(err)
	}
	return v
}

var errShouldNeverHappen = errors.New("should never happen")

func (d *vfio_pci_device) ErrorReady() error    { return errShouldNeverHappen }
func (d *vfio_pci_device) WriteReady() error    { return errShouldNeverHappen }
func (d *vfio_pci_device) WriteAvailable() bool { return false }
func (d *vfio_pci_device) String() string       { return "pci " + d.Device.String() }

// UIO file is ready when interrupt occurs.
func (d *vfio_pci_device) ReadReady() (err error) {
	var b [8]byte
	if _, err = syscall.Read(d.interrupt_event_fd, b[:]); err != nil {
		return
	}
	d.DriverDevice.Interrupt()
	return
}
