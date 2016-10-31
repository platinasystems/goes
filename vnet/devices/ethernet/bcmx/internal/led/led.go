// Copyright 2016 Platina Systems, Inc. All rights reserved.

package led

import (
	"github.com/platinasystems/go/elib/hw"

	"fmt"
	"time"
)

type reg hw.Reg32

func (r *reg) get() uint32  { return (*hw.Reg32)(r).Get() }
func (r *reg) set(v uint32) { (*hw.Reg32)(r).Set(v) }
func (r *reg) get8() uint8  { return uint8(r.get() & 0xff) }
func (r *reg) set8(v uint8) {
	val := r.get()
	val = (val & 0xffffff00) | (uint32(v) & 0xff)
	r.set(val)
}

type Regs struct {
	// [9:4] scan start delay
	// [3:1] intra port scan delay
	// [0] enable uproc
	control reg
	// [9] initializing
	// [8] running
	// [7:0] program counter
	status                         reg
	scan_chain_assembly_start_addr reg
	_                              reg
	// port order remap => used to 'remap' the LED id to front panel port id:
	// each port has 6 bits and 4 ports per register, so:
	// 31..24 - not used
	// 23..18 - port_id = (reg# * 4)+3
	// 17..12 - port_id = (reg# * 3)+2
	// 11..6  - port_id = (reg# * 2)+1
	// 5..0   - port_id = reg# * 4
	port_order_remap [16]reg
	clock_params     reg
	// upper 2 bits of scan out counter [9:0] [1:0]
	scan_out_count_upper reg
	tm_control           reg
	clock_divider        reg
	_                    [0x400 - 0x60]byte

	// beginning at offset 0xa0: 1 byte per from panel port
	// [0] lane 0 link status => 100G, 50G, 25G/10G
	// [1] lane 1 link status =>       50G, 25G/10G
	// [2] lane 2 link status =>            25G/10G
	// [3] lane 3 link status =>            25G/10G
	// [5:4] port mode => 0x00: 100G
	//                    0x01: 2x50G
	//                    0x10: 4x{25G or 10G}
	//                    0x11 - undefined
	// [6] maintenance => RED LED: 0 = off, 1 = on
	// [7] turbo mode - not currently used
	data_ram [256]reg

	program_ram [256]reg
	_           [0x1000 - 0xc00]byte
}

// port mode => lane link status information
const (
	oneLane  = 0 << 4 // 100G
	twoLane  = 1 << 4 // 50G
	fourLane = 2 << 4 // 25G/10G
)

type Led struct {
	bank                 int
	Regs                 *Regs
	data_ram_port_offset uint
}

type status int

const (
	running status = iota
	initializing
	disabled
	unknown
)

var status_strings = []string{
	running:      "Running",
	initializing: "Initializing",
	disabled:     "Disabled",
	unknown:      "Unknown",
}

func (s status) String() string { return status_strings[s] }

// Mutually exclusive states
func (l *Led) getStatus() (s status) {
	r := l.Regs
	var v [2]uint32
	v[0] = r.control.get()
	if v[0]&(1<<0) == 0 {
		return disabled
	}
	v[1] = r.status.get()
	if v[1]&(1<<9) != 0 {
		return initializing
	}
	if v[1]&(1<<8) != 0 {
		return running
	}
	return unknown
}

// unReset the led Microprocessor
func (l *Led) enable(enable bool) {
	r := l.Regs
	v := r.control.get()
	const (
		enable_uproc = 1 << 0
	)
	if enable {
		v |= enable_uproc
	} else {
		v &^= enable_uproc
	}
	r.control.set(v)
}

// This version has support for:
// - lane 0..3 link status => 100G, 50G, 25G, 10G
// - port mode: 100G, 50G, 25G, and 10G
// - maintenance mode: RED LED: 0 = off, 1 = on
// - port lane status periodic cycle: 1 to 4 in order
//
// Color Scheme:
//    - No Color: Link Down (for 100G port)
//    - Amber:    Link Down (for breakout: 50G/25G/10G ports)
//    - Green:    Link Up
//    - Red:      Maintenance Mode
var firmware = [...]uint8{
	0x02, 0x00, 0x60, 0xE1, 0x12, 0xA0, 0xF8, 0x15, 0x61, 0xE3, 0x67, 0x9D, 0x71, 0xAA, 0xCA, 0x30,
	0x70, 0x80, 0xDA, 0x10, 0x74, 0x1C, 0x12, 0x01, 0x61, 0xE5, 0x77, 0x20, 0x12, 0x03, 0x61, 0xE5,
	0x16, 0xE0, 0xCE, 0xE2, 0x74, 0xA2, 0x16, 0xE5, 0xEE, 0xE6, 0x71, 0xA2, 0x12, 0x40, 0x61, 0xE2,
	0x77, 0x80, 0x06, 0xE1, 0xF2, 0x04, 0xD2, 0x40, 0x74, 0x02, 0x86, 0xE0, 0x16, 0xE0, 0xDA, 0x3F,
	0x70, 0x4C, 0xDA, 0x80, 0x74, 0x64, 0x12, 0x00, 0x61, 0xE0, 0x77, 0x64, 0x86, 0xE6, 0x86, 0xE0,
	0x16, 0xE6, 0xDA, 0x04, 0x70, 0x58, 0x77, 0x64, 0x12, 0x00, 0x61, 0xE6, 0x12, 0x80, 0x61, 0xE2,
	0x12, 0x81, 0x61, 0xE0, 0x3A, 0x20, 0x16, 0xE3, 0xCA, 0x30, 0x70, 0x6F, 0x16, 0xE6, 0xF1, 0x28,
	0x32, 0x00, 0x32, 0x01, 0xB7, 0x97, 0x75, 0xB2, 0x16, 0xE0, 0xCA, 0x04, 0x74, 0xB2, 0x77, 0xA2,
	0x67, 0x8C, 0x75, 0x86, 0x77, 0x66, 0xCA, 0x30, 0x70, 0xA2, 0x77, 0xBA, 0x16, 0xE3, 0x02, 0x00,
	0xCA, 0x30, 0x70, 0x97, 0x16, 0xE6, 0x01, 0x16, 0xE3, 0x18, 0x06, 0xE1, 0x57, 0x16, 0xE3, 0x1A,
	0x06, 0x57, 0x32, 0x0E, 0x87, 0x32, 0x0E, 0x87, 0x77, 0x32, 0x32, 0x0F, 0x87, 0x32, 0x0E, 0x87,
	0x77, 0x32, 0x32, 0x0E, 0x87, 0x32, 0x0F, 0x87, 0x77, 0x32, 0x32, 0x0F, 0x87, 0x32, 0x0F, 0x87,
	0x77, 0x32,
}

// Clear Data RAM and Program RAM
func (l *Led) initMem() {
	var x uint8
	r := l.Regs
	for i := range r.program_ram {
		if i < len(firmware) {
			x = firmware[i]
		} else {
			// unused locations must be set to 0
			x = 0
		}
		r.program_ram[i].set8(uint8(x))
	}
	// init the upper half of the data ram
	for i := 0x80; i < 0xff; i++ {
		r.data_ram[i].set8(uint8(0))
	}
}

func (l *Led) SetPortState(port_index uint, mask uint8, isSet bool) {
	var index uint
	// FIXME: this is an MK1 100G specific map; offset by 4 for 100G!
	//              Led[0]          Led[1]
	// addr 0xA0    port 0          port  8
	// addr 0xBC    port 7          port  .
	// addr 0xC0    port 24         port  .
	// addr 0xDC    port 31         port 23

	// lane index
	lane_index := port_index % 4
	// plughole index
	port_index = port_index / 4
	if port_index == 0 || port_index < 8 {
		index = port_index * 4
	} else if port_index == 8 || port_index < 24 {
		index = (port_index - 8) * 4
	} else if port_index == 24 || port_index < 32 {
		index = (port_index - 16) * 4
	}
	// port index in context of led ram
	index += lane_index

	if false {
		fmt.Printf("SetPortState: port_index = %d lane_index = %d index = 0x%x isSet %v mask %x\n",
			port_index, lane_index, index, isSet, mask)
		fmt.Printf("SetPortState: port_index = %d\n", port_index)
		fmt.Printf("offset = 0x%x\n", index)
		fmt.Printf("led data ram index = 0x%x\n", l.data_ram_port_offset+index)
	}
	mem_index := l.data_ram_port_offset + index

	// This is a remap of front-panel leds to match the phy remap
	if true {
		remap_mem_index := mapLedLight[mem_index]
		if remap_mem_index != 0 {
			mem_index = remap_mem_index
		}
	}

	r := &l.Regs.data_ram[mem_index]
	v := uint8(r.get8())
	if isSet {
		v |= mask
	} else {
		v &^= mask
	}
	r.set8(v)
}

func (l *Led) DumpDataRam() {
	for i := 0; i < 256; i++ {
		r := &l.Regs.data_ram[i]
		v := uint8(r.get8())
		fmt.Printf("led_data[0x%x] = 0x%x\n", i, v)
	}
}

var bank0PortMap = []int{
	0x61969b,
	0x71d79f,
	0x411493,
	0x515597,
	0x20928b,
	0x30d38f,
	0x1083,
	0x105187,
	0xe39ebb,
	0xf3dfbf,
	0xc31cb3,
	0xd35db7,
	0xa29aab,
	0xb2dbaf,
	0x8218a3,
	0x9259a7,
}

var bank1PortMap = []int{
	0x105187,
	0x1083,
	0x30d38f,
	0x20928b,
	0x515597,
	0x411493,
	0x71d79f,
	0x61969b,
	0x9259a7,
	0x8218a3,
	0xb2dbaf,
	0xa29aab,
	0xd35db7,
	0xc31cb3,
	0xf3dfbf,
	0xe39ebb,
}

var mapLedLight = map[uint]uint{
	0xa0: 0xa4,
	0xa4: 0xa0,
	0xa8: 0xac,
	0xac: 0xa8,
	0xb0: 0xb4,
	0xb4: 0xb0,
	0xb8: 0xbc,
	0xbc: 0xb8,
	0xc0: 0xc4,
	0xc4: 0xc0,
	0xc8: 0xcc,
	0xcc: 0xc8,
	0xd0: 0xd4,
	0xd4: 0xd0,
	0xd8: 0xdc,
	0xdc: 0xd8,
}

func (l *Led) initPortMap() {
	var v int
	r := l.Regs
	for i := range r.port_order_remap {
		if l.bank == 0 {
			v = bank0PortMap[i]
		} else {
			v = bank1PortMap[i]
		}
		l.Regs.port_order_remap[i].set(uint32(v))
	}
}

func (l *Led) Init(data_ram_port_offset uint, bank int) {
	l.bank = bank
	l.data_ram_port_offset = data_ram_port_offset
	l.initMem()
	l.initPortMap()
	l.enable(true)

	// Poll LED Control Status: should be initializing, then running..
	start := time.Now()
	for {
		if l.getStatus() == running {
			break
		}
		if time.Since(start) >= 100*time.Millisecond {
			panic("led timeout: waiting for running state")
		}
	}
}
