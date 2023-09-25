// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package platina_eeprom

import "fmt"

const (
	ChassisTypeType      = Type(0x50)
	BoardTypeType        = Type(0x51)
	SubTypeType          = Type(0x52)
	PcbaNumberType       = Type(0x53)
	PcbaSerialNumberType = Type(0x54)

	Tor1CpuPcbaSerialNumberType  = Type(0x10)
	Tor1FanPcbaSerialNumberType  = Type(0x11)
	Tor1MainPcbaSerialNumberType = Type(0x12)

	VendorExtensionType = Type(0x00)
)

var Types = []Type{
	ChassisTypeType,
	BoardTypeType,
	SubTypeType,
	PcbaNumberType,
	PcbaSerialNumberType,

	Tor1CpuPcbaSerialNumberType,
	Tor1FanPcbaSerialNumberType,
	Tor1MainPcbaSerialNumberType,

	VendorExtensionType,
}

var typesByName = map[string]Type{
	"ChassisType":      ChassisTypeType,
	"BoardType":        BoardTypeType,
	"SubType":          SubTypeType,
	"PcbaNumber":       PcbaNumberType,
	"PcbaSerialNumber": PcbaSerialNumberType,

	"Tor1CpuPcbaSerialNumber":  Tor1CpuPcbaSerialNumberType,
	"Tor1FanPcbaSerialNumber":  Tor1FanPcbaSerialNumberType,
	"Tor1MainPcbaSerialNumber": Tor1MainPcbaSerialNumberType,

	"VendorExtension": VendorExtensionType,
}

type Type uint8

func (t Type) Byte() byte {
	return byte(t)
}

func (t Type) String() string {
	s := map[Type]string{
		ChassisTypeType:      "ChassisType",
		BoardTypeType:        "BoardType",
		SubTypeType:          "SubType",
		PcbaNumberType:       "PcbaNumber",
		PcbaSerialNumberType: "PcbaSerialNumber",

		Tor1CpuPcbaSerialNumberType:  "Tor1CpuPcbaSerialNumber",
		Tor1FanPcbaSerialNumberType:  "Tor1FanPcbaSerialNumber",
		Tor1MainPcbaSerialNumberType: "Tor1MainPcbaSerialNumber",

		VendorExtensionType: "VendorExtension",
	}[t]
	if len(s) == 0 {
		s = fmt.Sprintf("%#x", t)
	}
	return s
}
