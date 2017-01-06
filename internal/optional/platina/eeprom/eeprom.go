// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package eeprom

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/platinasystems/go/vnet/ethernet"
)

const (
	OnieDataSz = 8
	LenOffset  = OnieDataSz + 1
	HeaderSz   = LenOffset + 2

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

	ChassisTypeType      = Type(0x50)
	BoardTypeType        = Type(0x51)
	SubTypeType          = Type(0x52)
	PcbaNumberType       = Type(0x53)
	PcbaSerialNumberType = Type(0x54)

	Tor1CpuPcbaSerialNumberType  = Type(0x10)
	Tor1FanPcbaSerialNumberType  = Type(0x11)
	Tor1MainPcbaSerialNumberType = Type(0x12)
)

var Bus struct {
	Index   int
	Address int
	Delay   time.Duration
}

type Eeprom struct {
	Onie struct {
		Data    OnieData
		Version Hex8
	}
	Tlv TlvMap
}

type TlvMap tlvMap
type XtlvMap tlvMap
type tlvMap map[Type]ByteStringer

type ByteStringer interface {
	Bytes() []byte
	String() string
}

type BoardType uint8
type ChassisType uint8
type Dec16 uint16
type EthernetAddress ethernet.Address
type Hex8 uint8
type Hex32 uint32
type OnieData [OnieDataSz]byte
type SubType uint8
type String string
type Type uint8
type Unknown []byte

func NewEeprom(buf []byte) *Eeprom {
	p := new(Eeprom)
	i := copy(p.Onie.Data[:], buf[:OnieDataSz])
	p.Onie.Version = Hex8(buf[i])
	i++
	i += 2 // skip over length
	p.Tlv = make(TlvMap)
	p.Tlv.parse(buf[i:])
	return p
}

func (p *Eeprom) Bytes() []byte {
	tlvbuf := new(bytes.Buffer)
	for t, v := range p.Tlv {
		tlvbuf.WriteByte(byte(t))
		b := v.Bytes()
		tlvbuf.WriteByte(byte(len(b)))
		tlvbuf.Write(b)
	}
	buf := new(bytes.Buffer)
	buf.Write(p.Onie.Data[:])
	buf.WriteByte(byte(p.Onie.Version))
	binary.Write(buf, binary.BigEndian, uint16(tlvbuf.Len()))
	buf.Write(tlvbuf.Bytes())
	return buf.Bytes()
}

func (p *Eeprom) String() string {
	buf := new(bytes.Buffer)
	fmt.Fprintln(buf, "eeprom.Onie.Data:", p.Onie.Data)
	fmt.Fprintln(buf, "eeprom.Onie.Version:", p.Onie.Version)
	buf.WriteString(p.Tlv.String())
	return buf.String()
}

func (p *Eeprom) Equal(np *Eeprom) error {
	if p.Onie.Data != np.Onie.Data {
		return fmt.Errorf("Onie.Data: [% x] vs. [% x]",
			p.Onie.Data, np.Onie.Data)
	}
	if p.Onie.Version != np.Onie.Version {
		return fmt.Errorf("Onie.Version: %x vs. %x",
			p.Onie.Version, np.Onie.Version)
	}
	return p.Tlv.Equal(np.Tlv)
}

// Parse eeprom tlv's from given buf.
func (m TlvMap) parse(buf []byte) {
	for i, l := 0, 0; i < len(buf); i += 1 + 1 + l {
		t := Type(buf[i])
		l = int(buf[i+1])
		if l == 0 || i+l > len(buf) {
			continue
		}
		v := buf[i+1+1 : i+1+1+l]
		switch t {
		case ProductNameType:
			m[ProductNameType] = String(v)
		case PartNumberType:
			m[PartNumberType] = String(v)
		case SerialNumberType:
			m[SerialNumberType] = String(v)
		case BaseEthernetAddressType:
			ea := new(EthernetAddress)
			copy(ea[:], v)
			m[BaseEthernetAddressType] = ea
		case ManufactureDateType:
			m[ManufactureDateType] = String(v)
		case DeviceVersionType:
			m[DeviceVersionType] = Hex8(v[0])
		case LabelRevisionType:
			m[LabelRevisionType] = Hex8(v[0])
		case PlatformNameType:
			m[PlatformNameType] = String(v)
		case OnieVersionType:
			m[OnieVersionType] = String(v)
		case NEthernetAddressType:
			m[NEthernetAddressType] =
				Dec16(binary.BigEndian.Uint16(v))
		case ManufacturerType:
			m[ManufacturerType] = String(v)
		case CountryCodeType:
			m[CountryCodeType] = String(v)
		case VendorType:
			m[VendorType] = String(v)
		case DiagVersionType:
			m[DiagVersionType] = String(v)
		case ServiceTagType:
			m[ServiceTagType] = String(v)
		case VendorExtensionType:
			dv, found := m[DeviceVersionType].(Hex8)
			if found && (dv != 0x00) && (dv != 0xff) {
				xtlv := make(XtlvMap)
				xtlv.parse(v)
				m[VendorExtensionType] = xtlv
			} else {
				m[VendorExtensionType] = Unknown(v)
			}
		case CrcType:
			m[CrcType] = Hex32(binary.BigEndian.Uint32(v))
		default:
			m[t] = Unknown(v)
		}
	}
}

func (m TlvMap) Equal(nm TlvMap) error {
	for t, v := range m {
		if x, found := v.(XtlvMap); found {
			nx, found := nm[t].(XtlvMap)
			if !found {
				return fmt.Errorf("%s: NOTFOUND", t.String())
			}
			if err := x.Equal(nx); err != nil {
				return err
			}
			continue
		}
		nv, found := nm[t]
		if !found {
			return fmt.Errorf("%s: not found", t)
		}
		b := v.Bytes()
		nb := nv.Bytes()
		if !bytes.Equal(nb, b) {
			return fmt.Errorf("%s: [% x] vs. [% x]",
				t.String(), nb, b)
		}
	}
	return nil
}

// Parse eeprom's vendor tlv's from given buf.
func (m XtlvMap) parse(buf []byte) {
	for i, l := 0, 0; i < len(buf); i += 1 + 1 + l {
		t := Type(buf[i])
		l := int(buf[i+1])
		if l == 0 || i+l > len(buf) {
			continue
		}
		v := buf[i+1+1 : i+1+1+l]
		switch t {
		case ChassisTypeType:
			m[ChassisTypeType] = ChassisType(v[0])
		case BoardTypeType:
			m[BoardTypeType] = BoardType(v[0])
		case SubTypeType:
			m[SubTypeType] = SubType(v[0])
		case PcbaNumberType:
			m[PcbaNumberType] = String(v)
		case PcbaSerialNumberType:
			if string(v[0:3]) == "cpu" {
				m[Tor1CpuPcbaSerialNumberType] = String(v)
			} else if string(v[0:3]) == "fan" {
				m[Tor1FanPcbaSerialNumberType] = String(v)
			} else if string(v[0:4]) == "main" {
				m[Tor1MainPcbaSerialNumberType] = String(v)
			}
		default:
			m[t] = Unknown(v)
		}
	}
}

func (m XtlvMap) Equal(nm XtlvMap) error {
	for t, v := range m {
		nv, found := nm[t]
		if !found {
			return fmt.Errorf("%s: not found", t)
		}
		b := v.Bytes()
		nb := nv.Bytes()
		if !bytes.Equal(nb, b) {
			return fmt.Errorf("%s: [% x] vs.[% x]",
				t.String(), nb, b)
		}
	}
	return nil
}

func (m TlvMap) Bytes() []byte   { return tlvMap(m).Bytes() }
func (m TlvMap) String() string  { return tlvMap(m).String() }
func (m XtlvMap) Bytes() []byte  { return tlvMap(m).Bytes() }
func (m XtlvMap) String() string { return tlvMap(m).String() }

func (m tlvMap) Bytes() []byte {
	buf := new(bytes.Buffer)
	for t, v := range m {
		buf.WriteByte(byte(t))
		b := v.Bytes()
		buf.WriteByte(byte(len(b)))
		buf.Write(b)
	}
	return buf.Bytes()
}

func (m tlvMap) String() string {
	buf := new(bytes.Buffer)
	for t, v := range m {
		if x, found := v.(XtlvMap); found {
			buf.WriteString(x.String())
		} else {
			fmt.Fprint(buf, "eeprom.", t, ": ", v, "\n")
		}
	}
	return buf.String()
}

func (t BoardType) Bytes() []byte {
	return []byte{byte(t)}
}

func (t BoardType) String() string {
	s := map[BoardType]string{
		0x00: "ToR",
		0x01: "Broadwell 2-Core",
		0x02: "Broadwell 4-Core",
		0x03: "Broadwell 8-Core",
		0x04: "MC",
		0x05: "LC 32x100",
		0x06: "MCB",
		0x07: "Fan Controller",
	}[t]
	if len(s) == 0 {
		s = "unknown"
	}
	return s
}

func (t ChassisType) Bytes() []byte {
	return []byte{byte(t)}
}

func (t ChassisType) String() string {
	s := map[ChassisType]string{
		0x00: "ToR",
		0x01: "4-slot",
		0x02: "8-slot",
		0x03: "16-slot",
		0xff: "n/a",
	}[t]
	if len(s) == 0 {
		s = "unknown"
	}
	return s
}

func (t Dec16) Bytes() []byte {
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], uint16(t))
	return buf[:]
}

func (t Dec16) String() string {
	return fmt.Sprintf("%d", uint16(t))
}

func (t *EthernetAddress) Bytes() []byte {
	return []byte(t[:])
}

func (t *EthernetAddress) String() string {
	return (*ethernet.Address)(t).String()
}

func (t Hex8) Bytes() []byte {
	return []byte{byte(t)}
}

func (t Hex8) String() string {
	return fmt.Sprintf("%#02x", uint8(t))
}

func (t Hex32) Bytes() []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(t))
	return buf[:]
}

func (t Hex32) String() string {
	return fmt.Sprintf("%#08x", uint32(t))
}

func (t OnieData) Bytes() []byte {
	return []byte(t[:])
}

func (t OnieData) String() string {
	return fmt.Sprintf("%q [% x]", string(t[:]), t[:])
}

func (t SubType) Bytes() []byte {
	return []byte{byte(t)}
}

func (t SubType) String() string {
	s := map[SubType]string{
		0x00: "beta",
		0x01: "production",
		0xff: "alpha",
	}[t]
	if len(s) == 0 {
		s = "unknown"
	}
	return s
}

func (t String) Bytes() []byte {
	return []byte(t)
}

func (t String) String() string {
	return string(t)
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
		ChassisTypeType:         "ChassisType",
		BoardTypeType:           "BoardType",
		SubTypeType:             "SubType",
		PcbaNumberType:          "PcbaNumber",
		PcbaSerialNumberType:    "PcbaSerialNumber",

		Tor1CpuPcbaSerialNumberType:  "Tor1CpuPcbaSerialNumberType",
		Tor1FanPcbaSerialNumberType:  "Tor1FanPcbaSerialNumberType",
		Tor1MainPcbaSerialNumberType: "Tor1MainPcbaSerialNumberType",
	}[t]
	if len(s) == 0 {
		s = fmt.Sprintf("%#x", t)
	}
	return s
}

func (t Unknown) Bytes() []byte {
	return []byte(t)
}

func (t Unknown) String() string {
	return fmt.Sprintf("%q [% x]", string(t), []byte(t))
}
