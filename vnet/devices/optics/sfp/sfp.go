// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sfp

import (
	"fmt"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/platinasystems/go/internal/i2c"
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

type QsfpModule struct {
	// Read in when module is inserted and taken out of reset.
	sfpRegs SfpRegs

	signals [QsfpNSignal]QsfpSignal

	BusIndex   int
	BusAddress int

	Config QsfpModuleConfig
}

var (
	dummy       byte
	regsPointer = unsafe.Pointer(&dummy)
	regsAddr    = uintptr(unsafe.Pointer(&dummy))
)

func getQsfpRegs() *qsfpRegs { return (*qsfpRegs)(regsPointer) }
func upperMemoryPageRegs() unsafe.Pointer {
	return unsafe.Pointer(uintptr(regsPointer) + unsafe.Offsetof((*qsfpRegs)(nil).upperMemory))
}
func getQsfpThresholdRegs() *qsfpThresholdRegs { return (*qsfpThresholdRegs)(upperMemoryPageRegs()) }

func (r *reg8) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func (m *QsfpModule) i2cDo(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(m.BusIndex)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(m.BusAddress)
	if err != nil {
		return
	}

	err = bus.Do(rw, regOffset, size, data)
	return
}

func (r *reg8) get(m *QsfpModule) byte {
	var data i2c.SMBusData
	err := m.i2cDo(i2c.Read, r.offset(), i2c.ByteData, &data)
	if err != nil {
		panic(err)
	}
	return data[0]
}

func (r *reg8) setErr(m *QsfpModule, v uint8) error {
	var data i2c.SMBusData
	data[0] = v
	return m.i2cDo(i2c.Write, r.offset(), i2c.ByteData, &data)
}

func (r *reg8) set(m *QsfpModule, v uint8) {
	err := r.setErr(m, v)
	if err != nil {
		panic(err)
	}
}

func (r *reg16) get(m *QsfpModule) (v uint16) {
	var data i2c.SMBusData
	err := m.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
	return uint16(data[0])<<8 | uint16(data[1])
}

func (r *reg16) set(m *QsfpModule, v uint16) {
	var data i2c.SMBusData
	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	err := m.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
}

func (r *regi16) get(m *QsfpModule) (v int16) { v = int16((*reg16)(r).get(m)); return }
func (r *regi16) set(m *QsfpModule, v int16)  { (*reg16)(r).set(m, uint16(v)) }

func (r *QsfpSignal) get() (v bool) {
	// GPIO
	return
}

func (r *QsfpSignal) set(v bool) {
	// GPIO
}

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

func (m *QsfpModule) Present() {
	r := getQsfpRegs()

	// Wait for module to become ready.
	start := time.Now()
	for r.status.get(m)&(1<<0) != 0 {
		if time.Since(start) >= 100*time.Millisecond {
			panic("timeout")
		}
	}

	// Read EEPROM.
	if r.upperMemoryMapPageSelect.get(m) != 0 {
		r.upperMemoryMapPageSelect.set(m, 0)
	}
	p := (*[128]byte)(unsafe.Pointer(&m.sfpRegs))
	for i := byte(0); i < 128; i++ {
		p[i] = r.upperMemory[i].get(m)
	}

	// Might as well select page 3 forever.
	// If write fails ENXIO then optics module does not support write and we ignore page 3.
	err := r.upperMemoryMapPageSelect.setErr(m, 3)
	if errno, ok := err.(syscall.Errno); !ok || errno != syscall.ENXIO {
		panic(err)
	}
	if err == nil {
		tr := getQsfpThresholdRegs()
		m.Config.TemperatureInCelsius.get(m, &tr.temperature, TemperatureToCelsius)
		m.Config.SupplyVoltageInVolts.get(m, &tr.supplyVoltage, SupplyVoltageToVolts)
		m.Config.RxPowerInWatts.get(m, &tr.rxPower, RxPowerToWatts)
		m.Config.TxBiasCurrentInAmps.get(m, &tr.txBiasCurrent, TxBiasCurrentToAmps)
	}
}

func (m *QsfpModule) TxEnable(enableMask, laneMask uint) uint {
	r := getQsfpRegs()
	was := r.txDisable.get(m)
	disableMask := byte(^enableMask)
	is := 0xf & ((was &^ byte(laneMask)) | disableMask)
	if is != was {
		r.txDisable.set(m, is)
	}
	return uint(was)
}

func trim(b []byte) string {
	// Strip trailing nulls.
	if i := strings.IndexByte(string(b), 0); i >= 0 {
		b = b[:i]
	}
	return strings.TrimSpace(string(b))
}

func (r *SfpRegs) String() string {
	s := fmt.Sprintf("Id: %s, Connector Type: %s", r.Id.String(), r.ConnectorType.String())
	s += fmt.Sprintf("\n  Vendor: %s, Part Number %s, Revision %s, Serial %s, Date %s",
		trim(r.VendorName[:]), trim(r.VendorPartNumber[:]), trim(r.VendorRevision[:]),
		trim(r.VendorSerialNumber[:]), trim(r.VendorDateCode[:]))
	return s
}

func (m *QsfpModule) String() string {
	return m.sfpRegs.String()
}
