// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pcie

import (
	"fmt"
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/hw/pci"
)

const (
	Type_express_endpoont = iota
	Type_legacy_endpoont
	_
	_
	Type_root_port
	Type_upstream_port
	Type_downstream_port
	Type_pcie_to_pci_bridge
	Type_pci_to_pcie_bridge
	Type_root_complex_integrated_endpoint
	Type_root_complex_event_collector
)

var typeNames = [...]string{
	Type_express_endpoont:                 "express endpoont",
	Type_legacy_endpoont:                  "legacy endpoont",
	Type_root_port:                        "root port",
	Type_upstream_port:                    "upstream port",
	Type_downstream_port:                  "downstream port",
	Type_pcie_to_pci_bridge:               "pcie to pci bridge",
	Type_pci_to_pcie_bridge:               "pci to pcie bridge",
	Type_root_complex_integrated_endpoint: "root complex integrated endpoint",
	Type_root_complex_event_collector:     "root complex event collector",
}

type Type uint8

func (t Type) String() string { return elib.Stringer(typeNames[:], int(t)) }

type Flags struct {
	// [3:0] version (e.g. 2 for pcie gen 2, 3 for gen 3)
	Version uint8

	// [7:4] type (Type_*)
	Type Type

	// [8]
	SlotImplemented bool

	// [13:9]
	Msi uint8
}

type flags pci.U16

func (r *flags) Get(d *pci.Device) (v Flags) {
	x := (*pci.U16)(r).Get(d)
	v.Version = uint8(x & 0xf)
	v.Type = Type((x >> 4) & 0xf)
	v.SlotImplemented = x&(1<<8) != 0
	v.Msi = uint8((x >> 9) & 0x1f)
	return
}

func (f *Flags) String() string {
	return fmt.Sprintf("type %s, version %d, msi %d", f.Type.String(), f.Version, f.Msi)
}

type devCap pci.U32

type DeviceCapabilities struct {
	// In bytes: 128 to 4k.
	Log2MaxPayloadSize uint8
}

func (r *devCap) Get(d *pci.Device) (v DeviceCapabilities) {
	x := (*pci.U32)(r).Get(d)
	v.Log2MaxPayloadSize = uint8(7 + x&7)
	return
}

func (f *DeviceCapabilities) String() (s string) {
	s += fmt.Sprintf("log2 max payload %d", f.Log2MaxPayloadSize)
	return
}

type CapabilityHeader struct {
	pci.CapabilityHeader

	Flags flags

	Device struct {
		// Device capabilities:
		//   [2:0] x where max payload size = 2^(7+x)
		//   [4:3] phantom functions
		//   [5] extended tags
		//   [8:6] L0s acceptable latency
		//   [11:9] L1 acceptable latency
		//   [12] attention button present
		//   [13] attention indicator present
		//   [14] power indicator present
		//   [15] role based error reporting
		//   [25:18] slot power limit value
		//   [27:26] slot power limit scale
		//   [28] function level reset
		Capabilities devCap
		Control      pci.U16
		Status       pci.U16
	}

	Link, Slot struct {
		Capabilities pci.U32

		Control pci.U16

		Status pci.U16
	}
	Root struct {
		Control      pci.U16
		Capabilities pci.U16
		Status       pci.U32
	}
}

func GetCapabilityHeader(d *pci.Device) *CapabilityHeader {
	return (*CapabilityHeader)(d.GetCap(pci.PCIE))
}
