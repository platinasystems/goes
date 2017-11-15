// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fe1

import (
	"github.com/platinasystems/go/vnet/devices/optics/sfp"
	"github.com/platinasystems/go/vnet/ethernet"
)

type PortProvision struct {
	Name  string
	Lanes uint
	Speed string
	Count uint
}

type PortProvisionConfig struct {
	Ports []PortProvision
}

type PlatformConfig struct {
	SriovMode bool
	// Reset switch via gpio hard reset pin.
	DisableGpioSwitchReset bool
	// Reset switch via cpu soft reset.
	EnableCpuSwitchReset bool
	// Enable using PCI MSI interrupt for fe1 switch.
	EnableMsiInterrupt     bool
	UseCpuForPuntAndInject bool
	//Port Provisioning
	PortConfig PortProvisionConfig
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
