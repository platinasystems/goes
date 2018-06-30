// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fe1

import (
	"github.com/platinasystems/go/internal/xeth"
	"github.com/platinasystems/go/vnet/devices/optics/sfp"
	"github.com/platinasystems/go/vnet/ethernet"
)

// from go/main/goes-platina-mk1/vnetd.go:vnetdInit()
// see go/vnet/vnet.go:PortEntry
type PortProvision struct {
	Name         string
	Lanes        uint
	Speed        string
	Count        uint
	Portindex    int16
	Subportindex int8
	PuntIndex    uint8
	Vid          ethernet.VlanTag
}

type PortProvisionConfig struct {
	Ports []PortProvision
}

// later may add stg here
type BridgeProvision struct {
	PuntIndex        uint8
	Addr             [xeth.ETH_ALEN]uint8
	TaggedPortVids   []ethernet.VlanTag
	UntaggedPortVids []ethernet.VlanTag
}

// mapped by vid
type BridgeProvisionConfig struct {
	Bridges map[ethernet.VlanTag]*BridgeProvision
}

type PlatformConfig struct {
	KernelIxgbe   bool
	KernelIxgbevf bool
	// Reset switch via gpio hard reset pin.
	DisableGpioSwitchReset bool
	// Reset switch via cpu soft reset.
	EnableCpuSwitchReset bool
	// Enable using PCI MSI interrupt for fe1 switch.
	EnableMsiInterrupt     bool
	UseCpuForPuntAndInject bool

	PortConfig   PortProvisionConfig
	BridgeConfig BridgeProvisionConfig
}

type SwitchPort struct {
	Switch, Port uint8
}

// Platform configuration for FE1 based systems.
type Platform struct {
	Version             uint
	BaseEthernetAddress ethernet.Address
	NEthernetAddress    uint
	Init                func()
	QsfpModules         map[SwitchPort]*sfp.QsfpModule
	PlatformConfig
}

// Inject ports for fe1: allow both cpu (pcie) & 10g ixge ports for injects.
const (
	SingleTaggedInjectNextCpu = iota
	SingleTaggedInjectNextFirstIxge
)
