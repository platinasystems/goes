// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package eeprom

import (
	"encoding/binary"
	"fmt"

	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet/ethernet"
)

const (
	ProductNameType         = Type(0x21)
	PartNumberType          = Type(0x22)
	SerialNumberType        = Type(0x23)
	BaseEthernetAddressType = Type(0x24)
	ManufactureDateType     = Type(0x25)
	DeviceVersionType       = Type(0x26)
	LabelRevisionType       = Type(0x27)
	PlatformNameType        = Type(0x28)
	OnieVersionType         = Type(0x29)
	NEthernetAddressType    = Type(0x2a)
	ManufacturerType        = Type(0x2b)
	CountryCodeType         = Type(0x2c)
	VendorType              = Type(0x2d)
	DiagVersionType         = Type(0x2e)
	ServiceTagType          = Type(0x2f)
	VendorExtensionType     = Type(0xfd)
	CrcType                 = Type(0xfe)
)

var Types = []Typer{
	ProductNameType,
	PartNumberType,
	SerialNumberType,
	BaseEthernetAddressType,
	ManufactureDateType,
	DeviceVersionType,
	LabelRevisionType,
	PlatformNameType,
	OnieVersionType,
	NEthernetAddressType,
	ManufacturerType,
	CountryCodeType,
	VendorType,
	DiagVersionType,
	ServiceTagType,
	VendorExtensionType,
	CrcType,
}

var typesByName = map[string]Type{
	"ProductName":         ProductNameType,
	"PartNumber":          PartNumberType,
	"SerialNumber":        SerialNumberType,
	"BaseEthernetAddress": BaseEthernetAddressType,
	"ManufactureDate":     ManufactureDateType,
	"DeviceVersion":       DeviceVersionType,
	"LabelRevision":       LabelRevisionType,
	"PlatformName":        PlatformNameType,
	"OnieVersion":         OnieVersionType,
	"NEthernetAddress":    NEthernetAddressType,
	"Manufacturer":        ManufacturerType,
	"CountryCode":         CountryCodeType,
	"Vendor":              VendorType,
	"DiagVersion":         DiagVersionType,
	"ServiceTag":          ServiceTagType,
	"VendorExtension":     VendorExtensionType,
	"Crc":                 CrcType,
}

type Byter interface {
	Byte() byte
}

type Byteser interface {
	Bytes() []byte
}

type Deler interface {
	Del(string)
}

type Reseter interface {
	Reset()
}

type Setter interface {
	Set(string, string) error
}

type Scanner interface {
	Scan(string) error
}

type Typer interface {
	Byter
	fmt.Stringer
}

type VendorExtension interface {
	Byteser
	Deler
	Setter
	fmt.Stringer
}

type Type uint8

func (t Type) Byte() byte {
	return byte(t)
}

func (t Type) String() string {
	s := map[Type]string{
		ProductNameType:         "ProductName",
		PartNumberType:          "PartNumber",
		SerialNumberType:        "SerialNumber",
		BaseEthernetAddressType: "BaseEthernetAddress",
		ManufactureDateType:     "ManufactureDate",
		DeviceVersionType:       "DeviceVersion",
		LabelRevisionType:       "LabelRevision",
		PlatformNameType:        "PlatformName",
		OnieVersionType:         "OnieVersion",
		NEthernetAddressType:    "NEthernetAddress",
		ManufacturerType:        "Manufacturer",
		CountryCodeType:         "CountryCode",
		VendorType:              "Vendor",
		DiagVersionType:         "DiagVersion",
		ServiceTagType:          "ServiceTag",
		VendorExtensionType:     "VendorExtension",
		CrcType:                 "Crc",
	}[t]
	if len(s) == 0 {
		s = fmt.Sprintf("%#x", t)
	}
	return s
}

type Dec16 uint16

func (p *Dec16) Bytes() []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(*p))
	return buf[:]
}

func (p *Dec16) Scan(s string) error {
	_, err := fmt.Sscanf(s, "%d", p)
	return err
}

func (p *Dec16) String() string {
	return fmt.Sprintf("%d", uint16(*p))
}

func (p *Dec16) Write(b []byte) (int, error) {
	*p = Dec16(binary.BigEndian.Uint16(b))
	return 2, nil
}

type EthernetAddress ethernet.Address

func (p *EthernetAddress) Bytes() []byte {
	return []byte(p[:])
}

func (p *EthernetAddress) Scan(s string) error {
	input := new(parse.Input)
	input.SetString(s)
	(*ethernet.Address)(p).Parse(input)
	return nil
}

func (p *EthernetAddress) String() string {
	return (*ethernet.Address)(p).String()
}

func (p *EthernetAddress) Write(b []byte) (int, error) {
	copy(p[:], b)
	return 6, nil
}

type Hex8 uint8

func (p *Hex8) Bytes() []byte {
	return []byte{byte(*p)}
}

func (p *Hex8) Scan(s string) error {
	_, err := fmt.Sscanf(s, "%x", p)
	return err
}

func (p *Hex8) String() string {
	return fmt.Sprintf("%#02x", uint8(*p))
}

func (p *Hex8) Write(b []byte) (int, error) {
	*p = Hex8(b[0])
	return 1, nil
}

type Hex32 uint32

func (p *Hex32) Bytes() []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(*p))
	return buf[:]
}

func (p *Hex32) Scan(s string) error {
	_, err := fmt.Sscanf(s, "%x", p)
	return err
}

func (p *Hex32) String() string {
	return fmt.Sprintf("%#08x", uint32(*p))
}

func (p *Hex32) Write(b []byte) (int, error) {
	*p = Hex32(binary.BigEndian.Uint32(b))
	return 4, nil
}

type OnieData [OnieDataSz]byte

func (p *OnieData) Bytes() []byte {
	return p[:]
}

func (p *OnieData) Scan(s string) error {
	copy(p[:], []byte(s))
	return nil
}

func (p *OnieData) String() string {
	return fmt.Sprintf("%q %#0x", p[:], p[:])
}

func (p *OnieData) Write(b []byte) (int, error) {
	copy(p[:], b[:OnieDataSz])
	return OnieDataSz, nil
}
