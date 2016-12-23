// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package led

import (
	"strconv"
	"strings"
	"unsafe"

	"github.com/platinasystems/go/eeprom"
	"github.com/platinasystems/go/goes/internal/gpio"
	"github.com/platinasystems/go/goes/internal/log"
	"github.com/platinasystems/go/goes/internal/redis"
	"github.com/platinasystems/go/i2c"
)

const (
	maxFanTrays = 4
	maxPsu      = 2
)

var (
	dummy         byte
	regsPointer   = unsafe.Pointer(&dummy)
	regsAddr      = uintptr(unsafe.Pointer(&dummy))
	lastFanStatus [maxFanTrays]string
	lastPsuStatus [maxPsu]string
	psuLed             = []uint8{0x8, 0x10}
	psuLedYellow       = []uint8{0x8, 0x10}
	psuLedOff          = []uint8{0x04, 0x01}
	sysLed        byte = 0x1
	sysLedGreen   byte = 0x1
	sysLedYellow  byte = 0xc
	sysLedOff     byte = 0x80
	fanLed        byte = 0x6
	fanLedGreen   byte = 0x2
	fanLedYellow  byte = 0x6
	fanLedOff     byte = 0x0
	deviceVer     byte
	saveFanSpeed  string
	forceFanSpeed bool
)

type LedCon struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
	Reset    string
}

func getLedRegs() *ledRegs { return (*ledRegs)(regsPointer) }

func (r *reg8) offset() uint8   { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16r) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func (h *LedCon) i2cDo(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
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

func (h *LedCon) i2cDoMux(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
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

func (r *reg8) get(h *LedCon) byte {
	var data i2c.SMBusData
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in ledgpio: get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("ledgpio: get8 MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("ledgpio: get8 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	return data[0]
}

func (r *reg16) get(h *LedCon) (v uint16) {
	var data i2c.SMBusData
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in ledgpio: get16: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("ledgpio: get16 MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("ledgpio: get16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	return uint16(data[0])<<8 | uint16(data[1])
}

func (r *reg16r) get(h *LedCon) (v uint16) {
	var data i2c.SMBusData
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in ledgpio: get16r: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("ledgpio: get16r MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("ledgpio: get16r #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	return uint16(data[1])<<8 | uint16(data[0])
}

func (r *reg8) set(h *LedCon, v uint8) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in ledgpio: set8: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("ledgpio: set8 MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("ledgpio: set8 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *reg16) set(h *LedCon, v uint16) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in ledgpio: set16: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("ledgpio: set16 MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("ledgpio: set16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *reg16r) set(h *LedCon, v uint16) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in ledgpio: set16r: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("ledgpio: set16r MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("ledgpio: set16r #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *regi16) get(h *LedCon) (v int16) { v = int16((*reg16)(r).get(h)); return }
func (r *regi16) set(h *LedCon, v int16)  { (*reg16)(r).set(h, uint16(v)) }

func (h *LedCon) LedFpInit() {
	var d byte

	pin, found := gpio.Pins[h.Reset]
	if found {
		pin.SetValue(true)
	}

	e := eeprom.Device{
		BusIndex:   0,
		BusAddress: 0x55,
	}
	e.GetInfo()
	deviceVer = e.Fields.DeviceVersion
	if deviceVer == 0xff || deviceVer == 0x00 {
		psuLed = []uint8{0x0c, 0x03}
		psuLedYellow = []uint8{0x00, 0x00}
		psuLedOff = []uint8{0x04, 0x01}
		sysLed = 0xc0
		sysLedGreen = 0x0
		sysLedYellow = 0xc
		sysLedOff = 0x80
		fanLed = 0x30
		fanLedGreen = 0x10
		fanLedYellow = 0x20
		fanLedOff = 0x30
	}
	// save initial fan speed
	saveFanSpeed, _ = redis.Hget(redis.DefaultHash, "fan_tray.speed")
	forceFanSpeed = false

	r := getLedRegs()
	o := r.Output0.get(h)

	//on bmc boot up set front panel SYS led to green, FAN led to yellow, let PSU drive PSU LEDs
	d = 0xff ^ (sysLed | fanLed)
	o &= d
	o |= sysLedGreen | fanLedYellow
	r.Output0.set(h, o)
	o = r.Config0.get(h)
	o |= psuLed[0] | psuLed[1]
	o &= (sysLed | fanLed) ^ 0xff
	r.Config0.set(h, o)
}

func (h *LedCon) LedStatus() {
	r := getLedRegs()
	var o, c uint8
	var d byte

	if deviceVer == 0xff || deviceVer == 0x00 {
		psuLed = []uint8{0x0c, 0x03}
		psuLedYellow = []uint8{0x00, 0x00}
		psuLedOff = []uint8{0x04, 0x01}
		sysLed = 0xc0
		sysLedGreen = 0x0
		sysLedYellow = 0xc
		sysLedOff = 0x80
		fanLed = 0x30
		fanLedGreen = 0x10
		fanLedYellow = 0x20
		fanLedOff = 0x30
	}

	allFanGood := true
	fanStatChange := false
	for j := 1; j <= maxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j))+".status")
		if !strings.Contains(p, "ok") {
			allFanGood = false
		}
		if lastFanStatus[j-1] != p {
			fanStatChange = true
			//if any fan tray is failed or not installed, set front panel FAN led to yellow
			//log.Print ("last: ",lastFanStatus[j-1], " present: ", p)
			if strings.Contains(p, "warning") && !strings.Contains(lastFanStatus[j-1], "not installed") {
				o = r.Output0.get(h)
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedYellow
				r.Output0.set(h, o)
				log.Print("warning: fan tray ", j, " failure")
				if !forceFanSpeed {
					redis.Hset(redis.DefaultHash, "fan_tray.speed", "high")
					forceFanSpeed = true
				}
			} else if strings.Contains(p, "not installed") {
				o = r.Output0.get(h)
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedYellow
				r.Output0.set(h, o)
				log.Print("warning: fan tray ", j, " not installed")
				if !forceFanSpeed {
					redis.Hset(redis.DefaultHash, "fan_tray.speed", "high")
					forceFanSpeed = true
				}
			} else if strings.Contains(lastFanStatus[j-1], "not installed") && (strings.Contains(p, "warning") || strings.Contains(p, "ok")) {
				log.Print("notice: fan tray ", j, " installed")
			}
		}
		lastFanStatus[j-1] = p
	}

	if allFanGood && !forceFanSpeed {
		saveFanSpeed, _ = redis.Hget(redis.DefaultHash, "fan_tray.speed")
	}

	//if any fan tray is failed or not installed, set front panel FAN led to yellow
	if fanStatChange {
		if allFanGood {
			// if all fan trays have "ok" status, set front panel FAN led to green
			allStat := true
			for i := range lastFanStatus {
				if lastFanStatus[i] == "" {
					allStat = false
				}
			}
			if allStat {
				o = r.Output0.get(h)
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedGreen
				r.Output0.set(h, o)
				log.Print("notice: all fan trays up")
				redis.Hset(redis.DefaultHash, "fan_tray.speed", saveFanSpeed)
				forceFanSpeed = false
			}
		}

	}

	for j := 0; j < maxPsu; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "psu"+strconv.Itoa(j+1)+".status")
		if lastPsuStatus[j] != p {
			o = r.Output0.get(h)
			c = r.Config0.get(h)

			//if PSU is not installed or installed and powered on, set front panel PSU led to off or green (PSU drives)
			if strings.Contains(p, "not_installed") || strings.Contains(p, "powered_on") {
				c |= psuLed[j]
			} else if strings.Contains(p, "powered_off") {
				//if PSU is installed but powered off, set front panel PSU led to yellow
				d = 0xff ^ psuLed[j]
				o &= d
				o |= psuLedYellow[j]
				c &= (psuLed[j]) ^ 0xff
			}
			r.Output0.set(h, o)
			r.Config0.set(h, c)
			lastPsuStatus[j] = p
			if p != "" {
				log.Print("notice: psu", j+1, " ", p)
			}
		}
	}
}
