// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fe1

import (
	"github.com/platinasystems/go/vnet/ethernet"
)

type PlatformConfig struct {
	SriovMode bool
	// Reset switch via gpio hard reset pin.
	DisableGpioSwitchReset bool
	// Reset switch via cpu soft reset.
	EnableCpuSwitchReset bool
	// Enable using PCI MSI interrupt for fe1 switch.
	EnableMsiInterrupt bool
}

// Platform configuration for FE1 based systems.
type Platform struct {
	Version             uint
	BaseEthernetAddress ethernet.Address
	NEthernetAddress    uint
	Init                func()
	PlatformConfig
}
