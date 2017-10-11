// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mk1

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/optics/sfp"
	fe1_platform "github.com/platinasystems/go/vnet/platforms/fe1"

	"syscall"
	"time"
)

const (
	mux0_addr           = 0x70
	mux1_addr           = 0x71
	qsfp_addr           = 0x50
	qsfp_gpio_base_addr = 0x20
)

func i2cMuxSelectPort(port uint) {
	// Select 2 level mux.
	i2c.Do(0, mux0_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (port / 8)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
	i2c.Do(0, mux1_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (port % 8)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
}

func readWriteQsfp(addr uint8, b []byte, isWrite bool) (err error) {
	i, n_left := 0, len(b)

	for n_left >= 2 {
		err = i2c.Do(0, qsfp_addr, func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			if isWrite {
				d[0] = b[i+0]
				d[1] = b[i+1]
				err = bus.ReadWrite(i2c.Write, addr+uint8(i), i2c.WordData, &d)
			} else {
				err = bus.ReadWrite(i2c.Read, addr+uint8(i), i2c.WordData, &d)
				if err == nil {
					b[i+0] = d[0]
					b[i+1] = d[1]
				}
			}
			return
		})
		if err != nil {
			return
		}
		n_left -= 2
		i += 2
	}

	for n_left > 0 {
		err = i2c.Do(0, qsfp_addr, func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			if isWrite {
				d[0] = b[i+0]
				err = bus.ReadWrite(i2c.Write, addr+uint8(i), i2c.ByteData, &d)
			} else {
				err = bus.ReadWrite(i2c.Read, addr+uint8(i), i2c.ByteData, &d)
				if err == nil {
					b[i+0] = d[0]
				}
			}
			return
		})
		if err != nil {
			return
		}
		n_left -= 1
		i += 1
	}
	return
}

type qsfpStatus struct {
	// 1 => qsfp module is present
	is_present uint64
	// 1 => interrupt active
	interrupt_status uint64
}
type qsfpSignals struct {
	qsfpStatus

	// 1 => low power mode
	is_low_power_mode uint64
	// 1 => in reset; 0 not in reset
	is_reset_active uint64
}

// j == 0 => abs_l + int_l
// j == 1 => lpmode + rst_l
func readSignals(j uint) (v [2]uint32) {
	i2c.Do(0, mux0_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (4 + j)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
	// Read 0x20 21 22 23 to get 32 bits of status.
	for i := 0; i < 4; i++ {
		i2c.Do(0, qsfp_gpio_base_addr+i,
			func(bus *i2c.Bus) (err error) {
				var d i2c.SMBusData
				err = bus.Read(0, i2c.WordData, &d)
				v[i/2] |= (uint32(d[0]) | uint32(d[1])<<8) << (16 * uint(i%2))
				return
			})
	}
	return
}

const m32 = 1<<32 - 1

func (s *qsfpStatus) read() {
	v := readSignals(0)
	s.is_present = m32 &^ uint64(v[0])
	s.interrupt_status = m32 &^ uint64(v[1])
}

func (s *qsfpSignals) read() {
	s.qsfpStatus.read()
	v := readSignals(1)
	s.is_low_power_mode = uint64(v[0])
	s.is_reset_active = m32 &^ uint64(v[1])
}

func writeSignal(port uint, is_rst_l bool, value uint64) {
	i2c.Do(0, mux0_addr, func(bus *i2c.Bus) (err error) {
		var d i2c.SMBusData
		d[0] = 1 << (4 + 1)
		err = bus.Write(0, i2c.ByteData, &d)
		return
	})
	// 0x20 0x21 for lpmode
	slave := qsfp_gpio_base_addr + int(port/16)
	if is_rst_l {
		slave += 2 // 0x22 0x23 for rst_l
	}
	i2c.Do(0, slave,
		func(bus *i2c.Bus) (err error) {
			var d i2c.SMBusData
			d[0] = uint8(value >> (8 * (port / 8)))
			reg := uint8(0)
			if port%16 >= 8 {
				reg = 1
			}
			err = bus.Write(reg, i2c.ByteData, &d)
			return
		})
	return
}
func (s *qsfpSignals) writeLpmode(port uint) { writeSignal(port, false, s.is_low_power_mode) }
func (s *qsfpSignals) writeReset(port uint)  { writeSignal(port, true, ^s.is_reset_active) }

type qsfpState struct {
}

func (m *qsfpMain) signalChange(signal sfp.QsfpSignal, changedPorts, newValues uint64) {
	elib.Word(changedPorts).ForeachSetBit(func(i uint) {
		port := i ^ 1 // mk1 port swapping
		mod := m.module_by_port[port]
		v := newValues&(1<<i) != 0
		q := &mod.q
		q.SetSignal(signal, v)
	})
}

func (m *qsfpMain) poll() {
	sequence := 0
	for {
		old := m.current
		// Read initial state only first time; else read just status (presence + interrupt status).
		if sequence == 0 {
			m.current.read()
		} else {
			m.current.qsfpStatus.read()
		}
		new := m.current
		// Do lpmode/reset first; presence next; interrupt status last.
		// Presence change will have correct reset state when sequence == 0.
		if sequence == 0 {
			if d := new.is_low_power_mode ^ old.is_low_power_mode; d != 0 {
				m.signalChange(sfp.QsfpLowPowerMode, d, new.is_low_power_mode)
			}
			if d := new.is_reset_active ^ old.is_reset_active; d != 0 {
				m.signalChange(sfp.QsfpResetIsActive, d, new.is_reset_active)
			}
		}
		if d := new.is_present ^ old.is_present; d != 0 {
			m.signalChange(sfp.QsfpModuleIsPresent, d, new.is_present)
		}
		if d := new.interrupt_status ^ old.interrupt_status; d != 0 {
			m.signalChange(sfp.QsfpInterruptStatus, d, new.interrupt_status)
		}
		sequence++
		time.Sleep(1 * time.Second)
	}
}

type qsfpModule struct {
	// Index into m.current.* bitmaps.
	port_index uint
	m          *qsfpMain
	q          sfp.QsfpModule
}

type qsfpMain struct {
	current        qsfpSignals
	module_by_port []qsfpModule
}

func (q *qsfpModule) SfpReset(is_active bool) {
	mask := uint64(1) << q.port_index
	was_active := q.m.current.is_reset_active&mask != 0
	if is_active != was_active {
		q.m.current.is_reset_active ^= mask
		q.m.current.writeReset(q.port_index)
	}
}
func (q *qsfpModule) SfpSetLowPowerMode(is bool) {
	mask := uint64(1) << q.port_index
	was := q.m.current.is_low_power_mode&mask != 0
	if is != was {
		q.m.current.is_low_power_mode ^= mask
		q.m.current.writeLpmode(q.port_index)
	}
}
func (q *qsfpModule) SfpReadWrite(offset uint, p []uint8, isWrite bool) (write_ok bool) {
	i2cMuxSelectPort(q.port_index)
	err := readWriteQsfp(uint8(offset), p, isWrite)
	if write_ok = err == nil; !write_ok {
		if errno, ok := err.(syscall.Errno); !ok || errno != syscall.ENXIO {
			panic(err)
		}
	}
	return
}

func qsfpInit(v *vnet.Vnet, p *fe1_platform.Platform) {
	m := &qsfpMain{}

	p.QsfpModules = make(map[fe1_platform.SwitchPort]*sfp.QsfpModule)
	m.module_by_port = make([]qsfpModule, 32)
	for port := range m.module_by_port {
		q := &m.module_by_port[port]
		q.port_index = uint(port ^ 1)
		q.m = m
		q.q.Init(q)
		sp := fe1_platform.SwitchPort{Switch: 0, Port: uint8(port)}
		p.QsfpModules[sp] = &q.q
	}

	go m.poll()
}
