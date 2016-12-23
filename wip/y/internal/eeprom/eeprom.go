package eeprom

import (
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
)

// EEPROM TLVs.
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
)

type fields struct {
	ONIEData            [8]byte
	ONIEDataVersion     byte
	ProductName         string
	PlatformName        string
	VendorName          string
	Manufacturer        string
	Vendor              string
	LabelRevision       string
	PartNumber          string
	SerialNumber        string
	DeviceVersion       string
	ManufactureDate     string
	CountryCode         string
	DiagVersion         string
	ServiceTag          string
	VendorExtension     string
	ONIEVersion         string
	BaseEthernetAddress ethernet.Address
	NEthernetAddress    uint
	CRC32               uint
}

// i2c bus id, i2c bus address, size of device
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

	// always write the 'address' to location 0
	err = bus.Do(i2c.Write, 0, i2c.ByteData, data)

	// now read the data from the eeprom at location
	err = bus.Do(rw, regOffset, size, data)
	return
}

func (d *Device) getByte(i uint) byte {
	var data i2c.SMBusData
	data[0] = uint8(i)
	if err := d.i2cDo(i2c.Read, uint8(0), i2c.Byte, &data); err != nil {
		panic(fmt.Errorf("i2c.Read: %s", err))
	}
	return byte(data[0])
}

func (d *Device) getUint16(i uint) uint {
	b0 := uint(d.getByte(i + 0))
	b1 := uint(d.getByte(i + 1))
	return b0<<8 | b1
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
	for i = uint(0); i < uint(len(f.ONIEData)); i++ {
		f.ONIEData[i] = d.getByte(i)
	}
	f.ONIEDataVersion = d.getByte(i)
	dataLen := d.getUint16(i + 1)
	i += 3
	for j := uint(0); j < dataLen; j++ {
		d.rawData = append(d.rawData, d.getByte(i+j))
	}

	i = 0
	for i < dataLen {
		// Parse tlv
		t, l := d.rawData[i], uint(d.rawData[i+1])
		v := d.rawData[i+2 : i+2+l]
		i += 2 + l
		switch t {
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
			f.DeviceVersion = string(v)
		case label_revision:
			f.LabelRevision = string(v)
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
			f.VendorExtension = string(v)
		case crc:
			f.CRC32 = uint(v[0])<<24 | uint(v[1])<<16 | uint(v[2])<<8 | uint(v[3])
		default:
			panic(fmt.Errorf("unknown tlv in eeprom: %x %x", t, v))
		}
	}
	return
}
