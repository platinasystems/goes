// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package platina_eeprom

import (
	"bytes"
	"fmt"

	"github.com/platinasystems/go/internal/optional/eeprom"
)

const (
	ChassisTypeType      = eeprom.Type(0x50)
	BoardTypeType        = eeprom.Type(0x51)
	SubTypeType          = eeprom.Type(0x52)
	PcbaNumberType       = eeprom.Type(0x53)
	PcbaSerialNumberType = eeprom.Type(0x54)

	Tor1CpuPcbaSerialNumberType  = eeprom.Type(0x10)
	Tor1FanPcbaSerialNumberType  = eeprom.Type(0x11)
	Tor1MainPcbaSerialNumberType = eeprom.Type(0x12)
)

var TypesByName = map[string]eeprom.Type{
	"ChassisType":      ChassisTypeType,
	"BoardType":        BoardTypeType,
	"SubType":          SubTypeType,
	"PcbaNumber":       PcbaNumberType,
	"PcbaSerialNumber": PcbaSerialNumberType,

	"Tor1CpuPcbaSerialNumber":  Tor1CpuPcbaSerialNumberType,
	"Tor1FanPcbaSerialNumber":  Tor1FanPcbaSerialNumberType,
	"Tor1MainPcbaSerialNumber": Tor1MainPcbaSerialNumberType,
}

type XtlvMap map[eeprom.Type]*bytes.Buffer

func (m XtlvMap) Bytes() []byte {
	buf := new(bytes.Buffer)
	if b, found := m[eeprom.VendorExtensionType]; found {
		buf.Write(b.Bytes())
		return buf.Bytes()
	}
	for _, t := range []eeprom.Type{
		BoardTypeType,
		ChassisTypeType,
		SubTypeType,
	} {
		if b, found := m[t]; found {
			buf.WriteByte(byte(t))
			buf.WriteByte(byte(b.Len()))
			buf.Write(b.Bytes())
		}
	}
	for _, t := range []eeprom.Type{
		Tor1CpuPcbaSerialNumberType,
		Tor1FanPcbaSerialNumberType,
		Tor1MainPcbaSerialNumberType,
	} {
		if b, found := m[t]; found {
			buf.WriteByte(byte(PcbaSerialNumberType))
			buf.WriteByte(byte(b.Len()))
			buf.Write(b.Bytes())
		}
	}
	return buf.Bytes()
}

func (m XtlvMap) Del(name string) {
	t, found := TypesByName[name]
	if found {
		delete(m, t)
	}
}

func (m XtlvMap) Set(name, s string) error {
	t, found := TypesByName[name]
	if !found {
		return fmt.Errorf("%s: unknown", name)
	}
	b := new(bytes.Buffer)
	b.WriteString(s)
	m[t] = b
	return nil
}

func (m XtlvMap) String() string {
	if b, found := m[eeprom.VendorExtensionType]; found {
		return fmt.Sprintf("eeprom.VendorExtension: %q\n", b.String())
	}
	buf := new(bytes.Buffer)
	for _, t := range []eeprom.Type{
		BoardTypeType,
		ChassisTypeType,
		SubTypeType,
	} {
		b, found := m[t]
		if !found {
			continue
		}
		x := b.Bytes()[0]
		s := map[eeprom.Type]map[byte]string{
			BoardTypeType: map[byte]string{
				0x00: "ToR",
				0x01: "Broadwell 2-Core",
				0x02: "Broadwell 4-Core",
				0x03: "Broadwell 8-Core",
				0x04: "MC",
				0x05: "LC 32x100",
				0x06: "MCB",
				0x07: "Fan Controller",
			},
			ChassisTypeType: map[byte]string{
				0x00: "ToR",
				0x01: "4-slot",
				0x02: "8-slot",
				0x03: "16-slot",
				0xff: "n/a",
			},
			SubTypeType: map[byte]string{
				0x00: "beta",
				0x01: "production",
				0xff: "alpha",
			},
		}[t][x]
		if len(s) == 0 {
			s = fmt.Sprint("#0x", x)
		}
		fmt.Fprintf(buf, "eeprom.%s: %s\n", t, s)
	}
	for _, t := range []eeprom.Type{
		Tor1CpuPcbaSerialNumberType,
		Tor1FanPcbaSerialNumberType,
		Tor1MainPcbaSerialNumberType,
	} {
		b, found := m[t]
		if !found || b.Len() == 0 {
			continue
		}
		fmt.Fprintf(buf, "eeprom.%s: %s\n", t, b)
	}
	return buf.String()
}

func (m XtlvMap) Write(buf []byte) (n int, err error) {
	for len(buf) > 2 {
		t := eeprom.Type(buf[0])
		switch t {
		case BoardTypeType, ChassisTypeType, SubTypeType,
			PcbaSerialNumberType:
		default:
			v := new(bytes.Buffer)
			v.Write(buf)
			m[eeprom.VendorExtensionType] = v
			n = len(buf)
			return
		}
		l := int(buf[1])
		buf = buf[2:]
		if l == 0 {
			continue
		}
		if l > len(buf) {
			l = len(buf)
		}
		n += 2 + l
		if t == PcbaSerialNumberType {
			switch {
			case string(buf[:3]) == "cpu":
				t = Tor1CpuPcbaSerialNumberType
			case string(buf[:3]) == "fan":
				t = Tor1FanPcbaSerialNumberType
			case string(buf[:4]) == "main":
				t = Tor1MainPcbaSerialNumberType
			}
		}
		v, found := m[t]
		if !found {
			v = new(bytes.Buffer)
			m[t] = v
		} else {
			v.Reset()
		}
		v.Write(buf[:l])
		buf = buf[l:]
	}
	return
}
