// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package platina_eeprom

import (
	"time"

	"github.com/platinasystems/go/goes/cmd/eeprom"
)

var config struct {
	bus struct {
		index   int
		address int
		delay   time.Duration
	}
	minMacs int
	oui     [3]byte
}

type BusIndex int
type BusAddress int
type BusDelay time.Duration
type MinMacs int
type OUI [3]byte

func Config(args ...interface{}) {
	for _, arg := range args {
		switch t := arg.(type) {
		case BusIndex:
			config.bus.index = int(t)
		case BusAddress:
			config.bus.address = int(t)
		case BusDelay:
			config.bus.delay = time.Duration(t)
		case MinMacs:
			config.minMacs = int(t)
		case OUI:
			copy(config.oui[:], t[:])
		}
	}
	eeprom.Types = append(eeprom.Types,
		ChassisTypeType,
		BoardTypeType,
		SubTypeType,
		PcbaNumberType,
		PcbaSerialNumberType,

		Tor1CpuPcbaSerialNumberType,
		Tor1FanPcbaSerialNumberType,
		Tor1MainPcbaSerialNumberType,
	)
	eeprom.Vendor.New = func() eeprom.VendorExtension {
		return make(XtlvMap)
	}
	eeprom.Vendor.ReadBytes = ReadBytes
	eeprom.Vendor.Write = Write
}
