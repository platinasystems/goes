// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package m

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/hw/pci"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/cmic"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
	"github.com/platinasystems/go/vnet/ethernet"
)

// Common data shared among all types satisfying Switch interface.
type SwitchCommon struct {
	Vnet *vnet.Vnet

	// Index into Driver.Switches
	index uint

	PciDev *pci.Device

	cmic.Cmic

	SwitchConfig

	Ports      []Porter
	PortBlocks []PortBlocker
	Phys       []Phyer
}

func (c *SwitchCommon) GetSwitchCommon() *SwitchCommon    { return c }
func (c *SwitchCommon) GetPorts() []Porter                { return c.Ports }
func (c *SwitchCommon) GetPortBlocks() []PortBlocker      { return c.PortBlocks }
func (c *SwitchCommon) GetPhys() []Phyer                  { return c.Phys }
func (c *SwitchCommon) GetPhyReferenceClockInHz() float64 { return c.PhyReferenceClockInHz }
func (c *SwitchCommon) Name() (s string) {
	s = ""
	if p := GetPlatform(c.Vnet); len(p.Switches) > 1 {
		s = c.PciDev.Addr.String()
	}
	return
}

type SwitchConfig struct {
	// Bus function to match.
	PciFunction uint8

	PhyReferenceClockInHz float64
	CoreFrequencyInHz     float64

	Phys               []PhyConfig
	Ports              []PortConfig
	MMUPipeByPortBlock []uint8
}

type Switch interface {
	GetSwitchCommon() *SwitchCommon
	Init()
	PortInit(v *vnet.Vnet)
	Interrupt()
	PhyIDForPort(b sbus.Block, s SubPortIndex) (PhyID, PhyBusID)
	GetPhyReferenceClockInHz() float64
	GetPorts() []Porter
	GetPortBlocks() []PortBlocker
	GetPhys() []Phyer
	Name() string
	String() string
}

type Platform struct {
	vnet.Package
	// Platform ethernet address allocation block (from EEPROM).
	AddressBlock     ethernet.AddressBlock
	Switches         []Switch
	switchByDeviceID map[pci.VendorDeviceID]Switch
}

var packageIndex uint

func (p *Platform) InitPlatform(v *vnet.Vnet) {
	packageIndex = v.AddPackage("fe1", p)
	p.DependsOn("tuntap")
	p.DependedOnBy("pci-discovery")
	cliInit(v)
	return
}

func GetPlatform(v *vnet.Vnet) *Platform { return v.GetPackage(packageIndex).(*Platform) }

// Each block can have 1 2 or 4 ports.
type PortBlockIndex uint32

// Subport can be either 0 1 2 or 3
type SubPortIndex uint32

type LaneMask uint32

func (lm LaneMask) foreach(f func(lane LaneMask), isMask bool) {
	m := elib.Word(lm)
	var i int
	for m != 0 {
		m, i = elib.NextSet(m)
		lm := LaneMask(i)
		if isMask {
			lm = 1 << lm
		}
		f(lm)
	}
}

func (lm LaneMask) Foreach(f func(lane LaneMask))     { lm.foreach(f, false) }
func (lm LaneMask) ForeachMask(f func(lane LaneMask)) { lm.foreach(f, true) }

func (lm LaneMask) FirstLane() uint { return elib.FirstSet(elib.Word(lm)).MinLog2() }
func (lm LaneMask) NLanes() uint    { return elib.NSetBits(elib.Word(lm)) }

type PhyInterface uint32

const (
	PhyInterfaceInvalid PhyInterface = iota
	// KR[124] backplane interface.
	PhyInterfaceKR
	PhyInterfaceCR
	// Serial interface to optics: SGMII, SFI/XFI, MLD
	PhyInterfaceOptics
)

type PortCommon struct {
	Switch
	PortBlock PortBlocker
	Phy       Phyer
	LaneMask
	PhyInterface
	FrontPanelIndex uint
	SubPortIndex
	SpeedBitsPerSec float64
	Enabled         bool
	Autoneg         bool
	IsManagement    bool
}

type PortConfig struct {
	PortBlockIndex  uint
	SubPortIndex    uint
	IsManagement    bool
	SpeedBitsPerSec float64
	PhyInterface
	// fixme derive
	LaneMask
}

func (c *PortCommon) GetPortCommon() *PortCommon    { return c }
func (c *PortCommon) GetSwitch() Switch             { return c.Switch }
func (c *PortCommon) GetLaneMask() LaneMask         { return c.LaneMask }
func (c *PortCommon) GetPhyInterface() PhyInterface { return c.PhyInterface }
func (c *PortCommon) GetPortBlock() PortBlocker     { return c.PortBlock }
func (c *PortCommon) GetSubPortIndex() uint         { return uint(c.SubPortIndex) }
func (c *PortCommon) GetPhy() Phyer                 { return c.Phy }

func GetSubPortIndex(p Porter) uint { return p.GetPortCommon().GetSubPortIndex() }
func IsProvisioned(p Porter) bool   { return p.GetPortCommon().LaneMask != 0 }

type Porter interface {
	GetPortCommon() *PortCommon
	GetSwitch() Switch
	GetLaneMask() LaneMask
	GetPhyInterface() PhyInterface
	GetPhy() Phyer
	GetPortBlock() PortBlocker
	GetPortName() string
}

type PhyID uint8
type PhyBusID uint8

type PortLoopbackType int

const (
	PortLoopbackNone      PortLoopbackType = iota
	PortLoopbackMac                        // loopback at Mac
	PortLoopbackPhyLocal                   // local loopback at Phy
	PortLoopbackPhyRemote                  // remote loopback at Phy
)

type Phyer interface {
	GetPhyCommon() *PhyCommon
	Init()
	ValidateSpeed(p Porter, bitsPerSecond float64, isHiGig bool) bool
	SetSpeed(p Porter, bitsPerSecond float64, isHiGig bool)
	SetLoopback(p Porter, t PortLoopbackType)
	SetEnable(p Porter, enable bool)
	SetAutoneg(p Porter, enable bool)
}

type PhyConfig struct {
	Index           uint8
	FrontPanelIndex uint8 // Index on front panel of switch
	IsManagement    bool

	// Chips has 4 PHYSICAL serdes pins per core for Rx and 4 for Tx.
	// Signals are labelled FC[31:0]_[RT]D[NP][3:0] for Rx/Tx
	// Board layout may map these pins to any of 4 LOGICAL lanes in optics modules/backplane.
	Rx_logical_lane_by_phys_lane []uint8
	Tx_logical_lane_by_phys_lane []uint8

	Rx_invert_lane_polarity []bool
	Tx_invert_lane_polarity []bool
}

type PhyCommon struct{}

func (c *PhyCommon) GetPhyCommon() *PhyCommon { return c }

type PortBlocker interface {
	Name() string
	Enable(enable bool)
	GetPhy() Phyer
	GetPipe() uint
	GetPortBlockIndex() PortBlockIndex
	GetPortIndex(p Porter) uint
	SetPortEnable(p Porter, enable bool)
	SetPortLoopback(p Porter, enable bool)
}
