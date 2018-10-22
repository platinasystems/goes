// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package eeprom provides the ability to read data from an EEPROM device,
// connected to an i2c bus, conforming to a TLV format.
//
// The goes 'machine' must set the i2c bus and address before calling the
// GetInfo() at initialization time to collect data and store it into the
// Fields structure. Once collected and stored, the fields can be referenced
// by the goes code.

package eeprom

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"hash/crc32"

	"github.com/platinasystems/i2c"
)

// EEPROM TLVs offsets
const (
	product_name          = 0x21
	part_number           = 0x22
	serial_number         = 0x23
	base_ethernet_address = 0x24
	manufacture_date      = 0x25
	device_version        = 0x26
	label_revision        = 0x27
	platform_name         = 0x28
	onie_version          = 0x29
	n_ethernet_address    = 0x2a
	manufacturer          = 0x2b
	country_code          = 0x2c
	vendor                = 0x2d
	diag_version          = 0x2e
	service_tag           = 0x2f
	vendor_extension      = 0xfd
	crc                   = 0xfe
	//platina vendor extension fields
	chassis_type       = 0x50
	board_type         = 0x51
	sub_type           = 0x52
	pcba_number        = 0x53
	pcba_serial_number = 0x54
)

var ONIEId = "TlvInfo" + string(0x00)
var ONIEVer uint8 = 0x01
var lengthOffset uint = 9

// EEPROM TLV field types
type fields struct {
	ONIEData        [8]byte
	ONIEDataVersion byte
	ProductName     string
	PlatformName    string
	Manufacturer    string
	Vendor          string
	PartNumber      string
	SerialNumber    string
	DeviceVersion   byte
	ManufactureDate string
	CountryCode     string
	DiagVersion     string
	ServiceTag      string
	VendorExtension string
	ONIEVersion     string
	// FIXME BaseEthernetAddress ethernet.Address
	BaseEthernetAddress      [6]byte
	NEthernetAddress         uint
	CRC32                    uint
	ChassisType              byte
	BoardType                byte
	SubType                  byte
	PcbaPartNumber           string
	Tor1MainPcbaSerialNumber string
	Tor1CpuPcbaSerialNumber  string
	Tor1FanPcbaSerialNumber  string
}

// i2c bus id, i2c bus address, Fields of content, and raw data
type Device struct {
	BusIndex   int
	BusAddress int
	Fields     fields
	rawData    []byte
}

func (d *Device) i2cDo(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(d.BusIndex)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(d.BusAddress)
	if err != nil {
		return
	}

	// read the data from the eeprom..
	err = bus.Do(rw, regOffset, size, data)
	return
}

func (d *Device) getByte(i uint) byte {
	var data i2c.SMBusData
	data[0] = uint8(i & 0x00ff)
	var err error

	//write two byte address
	if err = d.i2cDo(i2c.Write, uint8(i>>8), i2c.ByteData, &data); err != nil {
		panic(err)
	}
	//EEPROM has a 5ms minimum write delay, wait 10ms
	time.Sleep(10 * time.Millisecond)

	//read byte
	if err = d.i2cDo(i2c.Read, uint8(0), i2c.Byte, &data); err != nil {
		panic(err)
	}
	return byte(data[0])

}

func (d *Device) setByte(a uint16, v uint8) error {

	var data i2c.SMBusData
	data[0] = uint8(a & 0x00ff)
	data[1] = uint8(v)

	//write two byte address followed by 1 byte data
	err := d.i2cDo(i2c.Write, uint8(a>>8), i2c.WordData, &data)

	//EEPROM has a 5ms minimum write delay, wait 10ms
	time.Sleep(10 * time.Millisecond)
	return err
}

func (d *Device) getUint16(i uint) uint {
	b0 := uint(d.getByte(i + 0))
	b1 := uint(d.getByte(i + 1))
	return ((b0 << 8) | b1)
}

func (d *Device) GetInfo() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	d.getInfo()
	return
}

func (d *Device) getInfo() {
	f := &d.Fields
	var i uint

	// ONIE data..
	for i = uint(0); i < uint(len(f.ONIEData)); i++ {
		f.ONIEData[i] = d.getByte(i)
	}
	f.ONIEDataVersion = d.getByte(i)
	dataLen := d.getUint16(i + 1)
	i += 3

	// now, the fields stuff into the rawData
	for j := uint(0); j < dataLen; j++ {
		d.rawData = append(d.rawData, d.getByte(i+j))
	}

	i = 0
	for i < dataLen {
		// Parse tlv (tlv offset, then tlv data length)
		tlv, tlen := d.rawData[i], uint(d.rawData[i+1])
		v := d.rawData[i+2 : i+2+tlen]
		i += 2 + tlen
		switch tlv {
		case product_name:
			f.ProductName = string(v)
		case part_number:
			f.PartNumber = string(v)
		case serial_number:
			f.SerialNumber = string(v)
		case base_ethernet_address:
			copy(f.BaseEthernetAddress[:], v)
		case n_ethernet_address:
			f.NEthernetAddress = uint(v[0])<<8 | uint(v[1])
		case manufacture_date:
			f.ManufactureDate = string(v)
		case device_version:
			f.DeviceVersion = v[0]
		case label_revision:
			// ignore
		case platform_name:
			f.PlatformName = string(v)
		case onie_version:
			f.ONIEVersion = string(v)
		case manufacturer:
			f.Manufacturer = string(v)
		case country_code:
			f.CountryCode = string(v)
		case vendor:
			f.Vendor = string(v)
		case diag_version:
			f.DiagVersion = string(v)
		case service_tag:
			f.ServiceTag = string(v)
		case vendor_extension:
			if (f.DeviceVersion != 0x00) && (f.DeviceVersion != 0xff) {
				for j := uint(4); j < uint(len(v)); {
					etlv, etlen := v[j], uint(v[j+1])
					ev := v[j+2 : j+2+etlen]
					switch etlv {
					case chassis_type:
						f.ChassisType = ev[0]
					case board_type:
						f.BoardType = ev[0]
					case sub_type:
						f.SubType = ev[0]
					case pcba_number:
						f.PcbaPartNumber = string(ev)
					case pcba_serial_number:
						if string(ev[0:3]) == "cpu" {
							f.Tor1CpuPcbaSerialNumber = string(ev)
						} else if string(ev[0:3]) == "fan" {
							f.Tor1FanPcbaSerialNumber = string(ev)
						} else if string(ev[0:4]) == "main" {
							f.Tor1MainPcbaSerialNumber = string(ev)
						}
					default:
					}
					j += 2 + etlen
				}
			}
			f.VendorExtension = string(v)
		case crc:
			f.CRC32 = uint(v[0])<<24 | uint(v[1])<<16 | uint(v[2])<<8 | uint(v[3])
		default:
			fmt.Fprint(os.Stderr, "unknown eeprom tlv: ", tlv, " value: ", v)
		}
	}
	return
}

func (d *Device) DumpProm() (bool, []byte) {
	f := &d.Fields
	var i uint8
	var rawData []byte

	//if onie ID is not valid, return
	for i = 0; i < uint8(len(f.ONIEData)); i++ {
		f.ONIEData[i] = d.getByte(uint(i))
		if f.ONIEData[i] != ONIEId[i] {
			return false, d.rawData
		}
	}

	//read and return entire ONIE prom including ID, ver, length fields
	dataLen := d.getUint16(lengthOffset)
	for j := uint(0); j < dataLen+11; j++ {
		rawData = append(rawData, d.getByte(j))
	}
	return true, rawData

}

func (d *Device) CalcCrc() string {
	f := &d.Fields
	var i uint8

	//if onie ID is not valid, return
	for i = 0; i < uint8(len(f.ONIEData)); i++ {
		f.ONIEData[i] = d.getByte(uint(i))
		if f.ONIEData[i] != ONIEId[i] {
			return "Invalid: EEPROM not in ONIE format"
		}
	}

	//read ONIE prom
	_, rawData := d.DumpProm()

	//calculate crc32 up old crc value, write new crc value
	l := uint16(len(rawData))
	checksum := crc32.ChecksumIEEE(rawData[0 : l-4])
	//checksum := crc32.ChecksumIEEE(rawData)
	d.setByte(l-4, uint8(checksum>>24))
	d.setByte(l-3, uint8(checksum>>16))
	d.setByte(l-2, uint8(checksum>>8))
	d.setByte(l-1, uint8(checksum&0xff))

	return "crc update complete"
}

func (d *Device) DeleteField(n string) string {
	//do not allow deleting of crc field
	if n == "fe" {
		return "Invalid: crc delete not allowed"
	}

	var found bool = false

	r, rawData := d.DumpProm()
	if !r {
		return "Invalid: EEPROM not in ONIE format"
	}

	//delete field + 2 byte header if found, shift remaining fields
	t, _ := strconv.ParseUint(n, 16, 64)
	dataLen := d.getUint16(lengthOffset)

	for i := uint(0 + 11); i < uint(len(rawData)); {
		tlv, tlen := rawData[i], uint(rawData[i+1])
		if tlv == byte(t) {
			//return "found"
			for j := i; j < (uint(len(rawData)) - 2 - tlen); j++ {
				d.setByte(uint16(j), rawData[j+2+tlen])
			}
			dataLen -= uint(tlen + 2)
			found = true
			break
		}
		i += 2 + tlen
	}
	if !found {
		return "field not found"
	}
	//update length field
	d.setByte(9, uint8(dataLen>>8))
	d.setByte(10, uint8(dataLen&0xFF))
	return "first matching field deleted"

}
func (d *Device) AddCrc() string {
	dataLen := d.getUint16(lengthOffset)
	d.setByte(uint16(lengthOffset), uint8((dataLen+6)>>8))
	d.setByte(uint16(lengthOffset+1), uint8((dataLen+6)&0xFF))
	d.setByte(uint16(dataLen+11), 0xfe)
	d.setByte(uint16(dataLen+12), 0x4)
	d.setByte(uint16(dataLen+13), 0x0)
	d.setByte(uint16(dataLen+14), 0x0)
	d.setByte(uint16(dataLen+15), 0x0)
	d.setByte(uint16(dataLen+16), 0x0)

	return "crc field added"
}

func (d *Device) CopyAll(rawData []byte) string {
	for j := uint(0); j < uint(len(rawData)); j++ {
		d.setByte(uint16(j), rawData[j])
	}
	return "copy complete"
}

func (d *Device) WriteField(n string, v []byte) string {
	f := &d.Fields
	var i uint8

	// write onie ID, onie version, length = 6, and placeholder crc32
	if strings.Contains(n, "onie") {
		for i = 0; i < uint8(len(ONIEId)); i++ {
			d.setByte(uint16(i), uint8(ONIEId[i]))
		}
		d.setByte(uint16(i), ONIEVer)
		d.setByte(uint16(i+1), 0)
		d.setByte(uint16(i+2), 6)
		d.setByte(uint16(i+3), crc)
		d.setByte(uint16(i+4), 4)
		d.setByte(uint16(i+5), 0)
		d.setByte(uint16(i+6), 0)
		d.setByte(uint16(i+7), 0)
		d.setByte(uint16(i+8), 0)
		return "onie header written"
	} else if strings.Contains(n, "length") {
		d.setByte(uint16(9), v[0])
		d.setByte(uint16(10), v[1])
		return "length written"
	} else {

		//if onie ID is not valid, return
		for i = 0; i < uint8(len(f.ONIEData)); i++ {
			f.ONIEData[i] = d.getByte(uint(i))
			if f.ONIEData[i] != ONIEId[i] {
				return "Invalid: EEPROM not in ONIE format"
			}
		}

		t, err := strconv.ParseUint(n, 16, 64)
		if err == nil {
			switch uint8(t) {
			case 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f, 0xfd:
				dataLen := d.getUint16(lengthOffset)
				o := uint16(dataLen + 11 - 6)
				d.setByte(o, uint8(t))
				d.setByte(o+1, uint8(len(v)))
				for i = 0; i < uint8(len(v)); i++ {
					d.setByte(uint16(i+uint8(o+2)), uint8(v[i]))
				}
				newLength := uint16(dataLen + uint(len(v)) + 2)
				d.setByte(uint16(lengthOffset), uint8(newLength>>8))
				d.setByte(uint16(lengthOffset+1), uint8(newLength&0xFF))
				d.setByte(newLength+11-6, crc)
				d.setByte(newLength+11-5, 0x4)
				d.setByte(newLength+11-4, 0)
				d.setByte(newLength+11-3, 0)
				d.setByte(newLength+11-2, 0)
				d.setByte(newLength+11-1, 0)
				return "write complete. run diag prom crc"
			default:
				return "Invalid: field not recognized"
			}
		}

	}

	return "Invalid or incomplete arguments"
}
