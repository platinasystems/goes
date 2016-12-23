// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package fantray

import (
	"strconv"
	"unsafe"

	"github.com/platinasystems/go/goes/internal/eeprom"
	"github.com/platinasystems/go/goes/internal/i2c"
	"github.com/platinasystems/go/goes/internal/log"
	"github.com/platinasystems/go/goes/internal/redis"
)

var (
	dummy       byte
	regsPointer = unsafe.Pointer(&dummy)
	regsAddr    = uintptr(unsafe.Pointer(&dummy))
)

type FanStat struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

const (
	fanTrayLeds = 0x33
	minRpm      = 2000
)

var fanTrayLedOff = []uint8{0x0, 0x0, 0x0, 0x0}
var fanTrayLedGreen = []uint8{0x20, 0x02, 0x20, 0x02}
var fanTrayLedYellow = []uint8{0x10, 0x01, 0x10, 0x01}
var fanTrayLedBits = []uint8{0x30, 0x03, 0x30, 0x03}
var fanTrayDirBits = []uint8{0x80, 0x08, 0x80, 0x08}
var fanTrayAbsBits = []uint8{0x40, 0x04, 0x40, 0x04}
var deviceVer byte

func getFanGpioRegs() *fanGpioRegs { return (*fanGpioRegs)(regsPointer) }

func (r *reg8) offset() uint8   { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16r) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func (h *FanStat) i2cDo(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(h.Bus)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(h.Addr)
	if err != nil {
		return
	}

	err = bus.Do(rw, regOffset, size, data)
	return
}

func (h *FanStat) i2cDoMux(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(h.MuxBus)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(h.MuxAddr)
	if err != nil {
		return
	}

	err = bus.Do(rw, regOffset, size, data)
	return
}

func (r *reg8) get(h *FanStat) byte {
	var data i2c.SMBusData
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fangpio: get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: get8 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Read, r.offset(), i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: get8 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	return data[0]
}

func (r *reg16) get(h *FanStat) (v uint16) {
	var data i2c.SMBusData
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fangpio: get16: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: get16 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: get16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	return uint16(data[0])<<8 | uint16(data[1])
}

func (r *reg16r) get(h *FanStat) (v uint16) {
	var data i2c.SMBusData
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fangpio: get16r: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: get16r MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: get16r #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	return uint16(data[1])<<8 | uint16(data[0])
}

func (r *reg8) set(h *FanStat, v uint8) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fangpio: set8: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: set8 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	data[0] = v
	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Write, r.offset(), i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: set8 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *reg16) set(h *FanStat, v uint16) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fangpio: set16: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: set16 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: set16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *reg16r) set(h *FanStat, v uint16) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fangpio: set16r: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: set16r MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	data[1] = uint8(v >> 8)
	data[0] = uint8(v)
	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fangpio: set16r #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *regi16) get(h *FanStat) (v int16) { v = int16((*reg16)(r).get(h)); return }
func (r *regi16) set(h *FanStat, v int16)  { (*reg16)(r).set(h, uint16(v)) }

func (h *FanStat) FanTrayLedInit() {
	r := getFanGpioRegs()

	e := eeprom.Device{
		BusIndex:   0,
		BusAddress: 0x55,
	}
	e.GetInfo()
	deviceVer = e.Fields.DeviceVersion
	if deviceVer == 0xff || deviceVer == 0x00 {
		fanTrayLedGreen = []uint8{0x10, 0x01, 0x10, 0x01}
		fanTrayLedYellow = []uint8{0x20, 0x02, 0x20, 0x02}
	} else {
		fanTrayLedGreen = []uint8{0x20, 0x02, 0x20, 0x02}
		fanTrayLedYellow = []uint8{0x10, 0x01, 0x10, 0x01}
	}

	r.Output[0].set(h, 0xff&(fanTrayLedOff[2]|fanTrayLedOff[3]))
	r.Output[1].set(h, 0xff&(fanTrayLedOff[0]|fanTrayLedOff[1]))
	r.Config[0].set(h, 0xff^fanTrayLeds)
	r.Config[1].set(h, 0xff^fanTrayLeds)
	log.Print("notice: fan tray led init complete")
}

func (h *FanStat) FanTrayStatus(i uint8) string {
	var s string
	var f string

	if deviceVer == 0xff || deviceVer == 0x00 {
		log.Print("jlp: here")
		fanTrayLedGreen = []uint8{0x10, 0x01, 0x10, 0x01}
		fanTrayLedYellow = []uint8{0x20, 0x02, 0x20, 0x02}
	} else {
		fanTrayLedGreen = []uint8{0x20, 0x02, 0x20, 0x02}
		fanTrayLedYellow = []uint8{0x10, 0x01, 0x10, 0x01}
	}

	r := getFanGpioRegs()
	n := 0
	i--

	if i < 2 {
		n = 1
	}

	o := r.Output[n].get(h)
	d := 0xff ^ fanTrayLedBits[i]
	o &= d

	if (r.Input[n].get(h) & fanTrayAbsBits[i]) != 0 {
		//fan tray is not present, turn LED off
		s = "not installed"
		o |= fanTrayLedOff[i]
	} else {
		//get fan tray air direction
		if (r.Input[n].get(h) & fanTrayDirBits[i]) != 0 {
			f = "front->back"
		} else {
			f = "back->front"
		}

		//check fan speed is above minimum
		f1 := "fan_tray." + strconv.Itoa(int(i+1)) + ".1.rpm"
		f2 := "fan_tray." + strconv.Itoa(int(i+1)) + ".2.rpm"
		s1, _ := redis.Hget(redis.DefaultHash, f1)
		s2, _ := redis.Hget(redis.DefaultHash, f2)
		r1, _ := strconv.ParseInt(s1, 10, 64)
		r2, _ := strconv.ParseInt(s2, 10, 64)

		if s1 == "" && s2 == "" {
			o |= fanTrayLedYellow[i]
		} else if (r1 > minRpm) && (r2 > minRpm) {
			s = "ok" + "." + f
			o |= fanTrayLedGreen[i]
		} else {
			s = "warning check fan tray"
			o |= fanTrayLedYellow[i]
		}
	}

	r.Output[n].set(h, o)
	return s
}
