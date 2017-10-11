// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sfp

import (
	"github.com/platinasystems/go/elib/hw"

	"fmt"
	"strings"
	"time"
	"unsafe"
)

type QsfpThreshold struct {
	Alarm, Warning struct{ Hi, Lo float64 }
}

type QsfpModuleConfig struct {
	TemperatureInCelsius QsfpThreshold
	SupplyVoltageInVolts QsfpThreshold
	RxPowerInWatts       QsfpThreshold
	TxBiasCurrentInAmps  QsfpThreshold
}

type Access interface {
	// Activate/deactivate module reset.
	SfpReset(active bool)
	// Enable/disable low power mode.
	SfpSetLowPowerMode(enable bool)
	SfpReadWrite(offset uint, p []uint8, isWrite bool) (ok bool)
}

type state struct {
	txDisable uint8
}

type QsfpModule struct {
	// Read in when module is inserted and taken out of reset.
	e Eeprom
	// Only compliance values from EEPROM are valid.
	eepromDataplaneSubsetValid bool
	// All values from EEPROM are valid.
	allEepromValid bool

	signalValues [QsfpNSignal]bool

	Config QsfpModuleConfig

	s state
	a Access
}

func getQsfpRegs() *qsfpRegs { return (*qsfpRegs)(hw.BasePointer) }
func getEepromRegs() *Eeprom {
	r := getQsfpRegs()
	return (*Eeprom)(unsafe.Pointer(hw.BaseAddress + uintptr(r.upperMemory[0].offset())))
}

func (r *reg8) offset() uint  { return uint(uintptr(unsafe.Pointer(r)) - hw.BaseAddress) }
func (r *reg16) offset() uint { return uint(uintptr(unsafe.Pointer(r)) - hw.BaseAddress) }

func (r *reg8) get(m *QsfpModule) reg8 {
	var b [1]uint8
	m.a.SfpReadWrite(r.offset(), b[:], false)
	return reg8(b[0])
}

func (r *reg8) set(m *QsfpModule, v uint8) bool {
	var b [1]uint8
	b[0] = v
	return m.a.SfpReadWrite(r.offset(), b[:], true)
}

func (r *reg16) get(m *QsfpModule) (v uint16) {
	var b [2]uint8
	m.a.SfpReadWrite(r.offset(), b[:], false)
	return uint16(b[0])<<8 | uint16(b[1])
}

func (r *reg16) set(m *QsfpModule, v uint16) (ok bool) {
	var b [2]uint8
	b[0] = uint8(v >> 8)
	b[1] = uint8(v)
	return m.a.SfpReadWrite(r.offset(), b[:], true)
}

func (r *regi16) get(m *QsfpModule) (v int16) { v = int16((*reg16)(r).get(m)); return }
func (r *regi16) set(m *QsfpModule, v int16)  { (*reg16)(r).set(m, uint16(v)) }

func (t *QsfpThreshold) get(m *QsfpModule, r *qsfpThreshold, unit float64) {
	t.Warning.Hi = float64(r.warning.hi.get(m)) * unit
	t.Warning.Lo = float64(r.warning.lo.get(m)) * unit
	t.Alarm.Hi = float64(r.alarm.hi.get(m)) * unit
	t.Alarm.Lo = float64(r.alarm.lo.get(m)) * unit
}

const (
	TemperatureToCelsius = 1 / 256.
	SupplyVoltageToVolts = 100e-6
	RxPowerToWatts       = 1e-7
	TxBiasCurrentToAmps  = 2e-6
)

func (m *QsfpModule) SetSignal(s QsfpSignal, new bool) (old bool) {
	old = m.signalValues[s]
	m.signalValues[s] = new
	if old != new {
		switch s {
		case QsfpModuleIsPresent:
			m.Present(new)
		case QsfpInterruptStatus:
			if new {
				m.Interrupt()
			}
		}
	}
	return
}
func (m *QsfpModule) GetSignal(s QsfpSignal) bool { return m.signalValues[s] }

func (m *QsfpModule) Init(a Access) { m.a = a }

func (m *QsfpModule) Interrupt() {
}

func (m *QsfpModule) Present(is bool) {
	r := getQsfpRegs()

	if !is {
		m.invalidateCache()
	} else {
		// Wait for module to become ready.
		start := time.Now()
		for r.status.get(m)&(1<<0) != 0 {
			if time.Since(start) >= 100*time.Millisecond {
				panic("ready timeout")
			}
		}
		// Read enough of EEPROM to keep dataplane happy.
		// Reading eeprom is slow over i2c.
		if r.upperMemoryMapPageSelect.get(m) != 0 {
			r.upperMemoryMapPageSelect.set(m, 0)
		}
		m.validateCache(false)
	}
}

func (m *QsfpModule) validateCache(everything bool) {
	r := getQsfpRegs()
	if everything && !m.allEepromValid {
		// Read whole EEPROM.
		m.allEepromValid = true
		p := (*[128]uint8)(unsafe.Pointer(&m.e))
		m.a.SfpReadWrite(r.upperMemory[0].offset(), p[:], false)
	} else if !m.eepromDataplaneSubsetValid {
		// For performance only read fields needed for data plane.
		m.eepromDataplaneSubsetValid = true
		er := getEepromRegs()
		m.e.Id = Id((*reg8)(&er.Id).get(m))
		m.e.ConnectorType = ConnectorType((*reg8)(&er.ConnectorType).get(m))
		m.e.Compatibility[0] = er.Compatibility[0].get(m)
		if Compliance(m.e.Compatibility[0])&ComplianceExtendedValid != 0 {
			m.e.Options[0] = er.Options[0].get(m)
		}
	}
}

func (m *QsfpModule) invalidateCache() {
	m.allEepromValid = false
	m.eepromDataplaneSubsetValid = false
}

func (m *QsfpModule) TxEnable(enableMask, laneMask uint) uint {
	r := getQsfpRegs()
	was := m.s.txDisable
	disableMask := byte(^enableMask)
	is := 0xf & ((was &^ byte(laneMask)) | disableMask)
	if is != was {
		r.txDisable.set(m, is)
		m.s.txDisable = is
	}
	return uint(was)
}

func (m *QsfpModule) GetId() Id                       { return m.e.Id }
func (m *QsfpModule) GetConnectorType() ConnectorType { return m.e.ConnectorType }
func (m *QsfpModule) GetCompliance() (c Compliance, x ExtendedCompliance) {
	c = Compliance(m.e.Compatibility[0])
	x = ExtendedComplianceUnspecified
	if c&ComplianceExtendedValid != 0 {
		x = ExtendedCompliance(m.e.Options[0])
	}
	return
}

func trim(r []reg8) string {
	// Strip trailing nulls.
	var b []byte
	for i := 0; i < len(r); i++ {
		if r[i] == 0 {
			break
		}
		b = append(b, byte(r[i]))
	}
	return strings.TrimSpace(string(b))
}

func (m *QsfpModule) String() string {
	m.validateCache(true)
	e := &m.e
	s := fmt.Sprintf("Id: %v", e.Id)
	s += fmt.Sprintf("\n  Vendor: %s, Part Number %s, Revision 0x%x, Serial %s, Date %s",
		trim(e.VendorName[:]), trim(e.VendorPartNumber[:]), trim(e.VendorRevision[:]),
		trim(e.VendorSerialNumber[:]), trim(e.VendorDateCode[:]))
	s += fmt.Sprintf("\n  Connector Type: %v", e.ConnectorType)

	c, x := m.GetCompliance()
	s += fmt.Sprintf("\n  Compliance: %v", c)
	if x != ExtendedComplianceUnspecified {
		s += fmt.Sprintf(" %v", x)
	}
	return s
}
