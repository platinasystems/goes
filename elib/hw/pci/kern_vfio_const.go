// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build vfio

package pci

import (
	"github.com/platinasystems/go/elib"
)

const vfio_api_version = 0

const (
	vfio_type1_iommu = 1 + iota
	vfio_spapr_tce_iommu
	vfio_type1v2_iommu
	vfio_dma_cc_iommu
	vfio_eeh
	vfio_type1_nesting_iommu
	vfio_spapr_tce_v2_iommu
	vfio_noiommu_iommu
)

type vfio_info_cap_header struct {
	id, version uint16
	next        uint32
}

type vfio_ioctl_kind int

var vfio_ioctl_kind_strings = [...]string{
	vfio_get_api_version:        "vfio_get_api_version",
	vfio_check_extension:        "vfio_check_extension",
	vfio_set_iommu:              "vfio_set_iommu",
	vfio_group_get_status:       "vfio_group_get_status",
	vfio_group_set_container:    "vfio_group_set_container",
	vfio_group_unset_container:  "vfio_group_unset_container",
	vfio_group_get_device_fd:    "vfio_group_get_device_fd",
	vfio_device_get_info:        "vfio_device_get_info",
	vfio_device_get_region_info: "vfio_device_get_region_info",
	vfio_device_get_irq_info:    "vfio_device_get_irq_info",
	vfio_device_set_irqs:        "vfio_device_set_irqs",
	vfio_device_reset:           "vfio_device_reset",
	vfio_iommu_get_info:         "vfio_iommu_get_info",
	vfio_iommu_map_dma:          "vfio_iommu_map_dma",
	vfio_iommu_unmap_dma:        "vfio_iommu_unmap_dma",
	vfio_iommu_enable:           "vfio_iommu_enable",
	vfio_iommu_disable:          "vfio_iommu_disable",
}

func (k vfio_ioctl_kind) String() string { return elib.StringerHex(vfio_ioctl_kind_strings[:], int(k)) }

const (
	// /dev/vfio/vfio ioctls.
	vfio_get_api_version vfio_ioctl_kind = iota + 0x3b64
	vfio_check_extension
	vfio_set_iommu
	// /dev/vfio/GROUP_NUMBER ioctls.
	vfio_group_get_status
	vfio_group_set_container
	vfio_group_unset_container
	vfio_group_get_device_fd
	// device fd ioctls.
	vfio_device_get_info
	vfio_device_get_region_info
	vfio_device_get_irq_info
	vfio_device_set_irqs
	vfio_device_reset
	vfio_ioctl_first_driver
)

const (
	// /dev/vfio/vfio driver ioctls.
	vfio_iommu_get_info vfio_ioctl_kind = iota + vfio_ioctl_first_driver
	vfio_iommu_map_dma
	vfio_iommu_unmap_dma
	vfio_iommu_enable
	vfio_iommu_disable
)

const (
	vfio_device_get_pci_hot_reset_info vfio_ioctl_kind = iota + vfio_ioctl_first_driver
	vfio_device_pci_hot_reset
)

type vfio_ioctl_common struct {
	argsz, flags uint32
}

func (c *vfio_ioctl_common) set_flags(flags uint) { c.flags = uint32(flags) }
func (c *vfio_ioctl_common) set(size uintptr, flags uint) {
	c.argsz = uint32(size)
	c.set_flags(flags)
}
func (c *vfio_ioctl_common) set_size(size uintptr) {
	c.argsz = uint32(size)
	c.flags = 0
}

type vfio_group_status struct {
	vfio_ioctl_common
}

const (
	vfio_group_flags_viable = 1 << iota
	vfio_group_flags_container_set
)

type vfio_device_info struct {
	vfio_ioctl_common
	num_regions uint32 /* Max region index + 1 */
	num_irqs    uint32 /* Max IRQ index + 1 */
}

const (
	vfio_device_flags_reset = 1 << iota
	vfio_device_flags_pci
	vfio_device_flags_platform
	vfio_device_flags_amba
	vfio_device_flags_ccw
)

const (
	vfio_device_api_pci_string      = "vfio-pci"
	vfio_device_api_platform_string = "vfio-platform"
	vfio_device_api_amba_string     = "vfio-amba"
	vfio_device_api_ccw_string      = "vfio-ccw"
)

type vfio_region_info struct {
	vfio_ioctl_common
	index      uint32
	cap_offset uint32
	size       uint64
	offset     uint64
}

const (
	vfio_region_info_flag_read = 1 << iota
	vfio_region_info_flag_write
	vfio_region_info_flag_mmap
	vfio_region_info_flag_caps
)

const (
	vfio_region_info_cap_kind_sparse_mmap = 1 + iota
	vfio_region_info_cap_kind_type
)

type vfio_region_sparse_mmap_area struct {
	offset, size uint64
}

type vfio_region_info_cap_sparse_mmap struct {
	vfio_info_cap_header
	nr_areas uint32
	_        uint32
	// areas follow
}

type vfio_region_info_cap_type struct {
	vfio_info_cap_header
	kind    uint32 /* global per bus driver */
	subtype uint32 /* type specific */
}

const (
	vfio_region_type_pci_vendor_type = 1 << 31
	vfio_region_type_pci_vendor_mask = 0xffff
)

/* 8086 Vendor sub-types */
const (
	vfio_region_subtype_intel_igd_opregion = iota + 1
	vfio_region_subtype_intel_igd_host_cfg
	vfio_region_subtype_intel_igd_lpc_cfg
)

type vfio_irq_info struct {
	vfio_ioctl_common
	index uint32 /* IRQ index */
	count uint32 /* Number of IRQs within this index */
}

const (
	vfio_irq_info_eventfd = 1 << iota
	vfio_irq_info_maskable
	vfio_irq_info_automasked
	vfio_irq_info_noresize
)

type vfio_irq_set struct {
	vfio_ioctl_common
	index uint32 // one of vfio_pci_*_irq_index
	start uint32 // first irq
	count uint32 // number of bytes of data
	// count data items follows (either 1 byte each for bool or 4 bytes each for eventfd)
}

const (
	vfio_irq_set_data_none      = 1 << iota /* Data not present */
	vfio_irq_set_data_bool                  /* Data is bool (uint8) */
	vfio_irq_set_data_eventfd               /* Data is eventfd (int32) */
	vfio_irq_set_action_mask                /* Mask interrupt */
	vfio_irq_set_action_unmask              /* Unmask interrupt */
	vfio_irq_set_action_trigger             /* Trigger interrupt */
)

const (
	vfio_irq_set_data_type_mask   = vfio_irq_set_data_none | vfio_irq_set_data_bool | vfio_irq_set_data_eventfd
	vfio_irq_set_action_type_mask = vfio_irq_set_action_mask | vfio_irq_set_action_unmask | vfio_irq_set_action_trigger
)

const (
	vfio_pci_bar0_region_index = iota
	vfio_pci_bar1_region_index
	vfio_pci_bar2_region_index
	vfio_pci_bar3_region_index
	vfio_pci_bar4_region_index
	vfio_pci_bar5_region_index
	vfio_pci_rom_region_index
	vfio_pci_config_region_index
	vfio_pci_vga_region_index
	vfio_pci_num_regions
)

const (
	vfio_pci_intx_irq_index = iota
	vfio_pci_msi_irq_index
	vfio_pci_msix_irq_index
	vfio_pci_err_irq_index
	vfio_pci_req_irq_index
	vfio_pci_num_irqs
)

const (
	vfio_ccw_config_region_index = iota
	vfio_ccw_num_regions
)

const (
	vfio_ccw_io_irq_index = iota
	vfio_ccw_num_irqs
)

type vfio_pci_dependent_device struct {
	group_id uint32
	segment  uint16
	bus      uint8
	devfn    uint8 /* Use PCI_SLOT/PCI_FUNC */
}

type vfio_pci_hot_reset_info struct {
	vfio_ioctl_common
	count uint32
	// devices []vfio_pci_dependent_device follow
}

type vfio_pci_hot_reset struct {
	vfio_ioctl_common
	count uint32
	// group_fds []int32 follow
}

type vfio_iommu_type1_info struct {
	vfio_ioctl_common
	iova_pgsizes uint64 /* Bitmap of supported page sizes */
}

const (
	vfio_iommu_info_pgsizes = 1 << iota /* supported page sizes info */
)

type vfio_iommu_type1_dma_map struct {
	vfio_ioctl_common
	vaddr uint64 /* Process virtual address */
	iova  uint64 /* IO virtual address */
	size  uint64 /* Size of mapping (bytes) */
}

const (
	vfio_dma_map_flag_read  = 1 << iota /* readable from device */
	vfio_dma_map_flag_write             /* writable from device */
)

type vfio_iommu_type1_dma_unmap struct {
	vfio_ioctl_common
	iova uint64 /* IO virtual address */
	size uint64 /* Size of mapping (bytes) */
}
