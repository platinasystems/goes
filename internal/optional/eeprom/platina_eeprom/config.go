// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package platina_eeprom

import (
	"time"

	"github.com/platinasystems/go/internal/optional/eeprom"
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
	eeprom.Vendor.Extension.NamesByType = map[eeprom.Type]string{
		ChassisTypeType:      "ChassisType",
		BoardTypeType:        "BoardType",
		SubTypeType:          "SubType",
		PcbaNumberType:       "PcbaNumber",
		PcbaSerialNumberType: "PcbaSerialNumber",

		Tor1CpuPcbaSerialNumberType:  "Tor1CpuPcbaSerialNumber",
		Tor1FanPcbaSerialNumberType:  "Tor1FanPcbaSerialNumber",
		Tor1MainPcbaSerialNumberType: "Tor1MainPcbaSerialNumber",
	}
	eeprom.Vendor.Extension.New = func() interface{} {
		return make(XtlvMap)
	}
	eeprom.Vendor.ReadBytes = ReadBytes
	eeprom.Vendor.Write = Write
}
