// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build zion

package main

import (
	"flag"
	"fmt"
	"unsafe"

	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm"
	"github.com/platinasystems/vnetdevices/optics/sfp"
)

type qsfp_cpld_regs struct {
	revision  reg8
	interrupt [2]reg8
	present   [2]reg8
	// active low reset
	active_low_reset          [2]reg8
	low_power_mode_active_low [2]reg8
	module_select_active_low  [2]reg8
}

var (
	dummy       byte
	regsPointer = unsafe.Pointer(&dummy)
	regsAddr    = uintptr(unsafe.Pointer(&dummy))
)

type reg8 byte

func get_qsfp_cpld_regs() *qsfp_cpld_regs { return (*qsfp_cpld_regs)(regsPointer) }
func (r *reg8) offset() uint8             { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func i2cDo(busIndex, busAddr int, rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(busIndex)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(busAddr)
	if err != nil {
		return
	}

	err = bus.Do(rw, regOffset, size, data)
	return
}

func (m *zionMain) i2cDo(addr int, rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) {
	err := i2cDo(m.i2cBusIndex, addr, rw, regOffset, size, data)
	if err != nil {
		panic(err)
	}
}

func (r *reg8) get(m *zionCpld) byte {
	var data i2c.SMBusData
	zion.i2cDo(m.busAddress, i2c.Read, r.offset(), i2c.ByteData, &data)
	return data[0]
}

func (r *reg8) set(m *zionCpld, v uint8) {
	var data i2c.SMBusData
	data[0] = v
	zion.i2cDo(m.busAddress, i2c.Write, r.offset(), i2c.ByteData, &data)
}

type zionCpld struct {
	busAddress int
}

type zionMain struct {
	i2cBusIndex   int
	muxBusAddress int
	qsfpBusAdress int
	cplds         [4]zionCpld
	qsfps         [32]sfp.QsfpModule
}

var zion = &zionMain{
	i2cBusIndex:   0,
	muxBusAddress: 0x70,
	qsfpBusAdress: 0x50,
	cplds: [4]zionCpld{
		{busAddress: 0x78},
		{busAddress: 0x79},
		{busAddress: 0x7a},
		{busAddress: 0x7b},
	},
}

func zionOpticsInit(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	// Have mux address all devices.
	zion.i2cDo(zion.muxBusAddress, i2c.Write, 0, i2c.ByteData, &i2c.SMBusData{0xff})

	r := get_qsfp_cpld_regs()
	for ci := range zion.cplds {
		cp := &zion.cplds[ci]
		present := elib.Word(r.present[0].get(cp))
		reset := elib.Word(r.active_low_reset[0].get(cp))
		toReset := elib.Word(0)
		var pi int
		for p := present; p != 0; {
			p, pi = p.NextSet()
			m := elib.Word(1) << uint(pi)
			if reset&m == 0 {
				toReset |= m
			}
		}
		if toReset != 0 {
			r.active_low_reset[0].set(cp, uint8(reset|toReset))
		}

		// Read eprom for all present modules & enable tx laser.
		for p := present; p != 0; {
			p, pi = p.NextSet()
			portIndex := 8*ci + pi
			mod := &zion.qsfps[portIndex]
			mod.BusIndex = zion.i2cBusIndex
			mod.BusAddress = zion.qsfpBusAdress
			r.module_select_active_low[0].set(cp, ^uint8(1<<uint(pi)))
			r.low_power_mode_active_low[0].set(cp, 0xff)
			mod.Present()
			mod.TxEnable(0xf, 0xf) // enable all lanes
			fmt.Fprintf(w, "%d: %+v\n", portIndex, mod)
		}
	}
	return
}

func init() {
	vnet.AddInit(func(v *vnet.Vnet) {
		v.CliAdd(&cli.Command{
			Name:   "zion optics init",
			Action: zionOpticsInit,
		})
	})
}

var phyConfigs = []bcm.PhyConfig{
	bcm.PhyConfig{
		Index: 0,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 1,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 2,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 3,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 4,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 5,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 6,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 7,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
	},
	bcm.PhyConfig{
		Index: 8,
		Rx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
		Tx_logical_lane_by_phys_lane: []uint8{1, 0, 3, 2},
	},
	bcm.PhyConfig{
		Index: 9,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 10,
		Rx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
		Tx_logical_lane_by_phys_lane: []uint8{1, 0, 3, 2},
	},
	bcm.PhyConfig{
		Index: 11,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 12,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{3, 2, 1, 0},
	},
	bcm.PhyConfig{
		Index: 13,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 14,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{3, 2, 1, 0},
	},
	bcm.PhyConfig{
		Index: 15,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 16,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{1, 0, 3, 2},
	},
	bcm.PhyConfig{
		Index: 17,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 18,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{3, 2, 1, 0},
	},
	bcm.PhyConfig{
		Index: 19,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 20,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{3, 2, 1, 0},
	},
	bcm.PhyConfig{
		Index: 21,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 22,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{3, 2, 1, 0},
	},
	bcm.PhyConfig{
		Index: 23,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 24,
		Rx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 25,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 26,
		Rx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 27,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 28,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 29,
		Rx_logical_lane_by_phys_lane: []uint8{0, 3, 2, 1},
		Tx_logical_lane_by_phys_lane: []uint8{2, 3, 0, 1},
	},
	bcm.PhyConfig{
		Index: 30,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	bcm.PhyConfig{
		Index: 31,
		Rx_logical_lane_by_phys_lane: []uint8{2, 1, 0, 3},
		Tx_logical_lane_by_phys_lane: []uint8{0, 1, 2, 3},
	},
	// TSCE Eagle management ports.
	bcm.PhyConfig{
		Index:        0,
		IsManagement: true,
	},
}

func (p *platform) boardInit() (err error) {
	p.AddressBlock = ethernet.AddressBlock{
		Base:  ethernet.Address{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0},
		Count: 256,
	}
	return
}

func (p *platform) boardPortInit(s bcm.Switch) (err error) {
	cf := bcm.SwitchConfig{}

	var (
		nPorts uint
		speed  float64
		pm     uint
		lb     bool
	)

	flag.UintVar(&nPorts, "n-ports", 32, "Number of ports (32 64 128)")
	flag.Float64Var(&speed, "speed", 100e9, "Port speed in bits per second")
	flag.UintVar(&pm, "pm", 4, "PortMode(0-4x1, 1-2x1_1x2, 2-1x2_2x1, 3-2x2, 4-1x4)")
	flag.BoolVar(&lb, "lb", false, "true or false")
	flag.Parse()

	switch nPorts {
	case 32, 64, 96, 128:
		break
	default:
		panic("nports")
	}

	cf.Ports = make([]bcm.PortConfig, nPorts)

	switch nPorts {
	case 32:
		for i := range cf.Ports {
			cf.Ports[i] = bcm.PortConfig{
				PortBlockIndex:  uint(i),
				SpeedBitsPerSec: speed,
				LaneMask:        0xf,
			}
		}
	case 64:
		for i := range cf.Ports {
			cf.Ports[i] = bcm.PortConfig{
				PortBlockIndex:  uint(i) / 2,
				SpeedBitsPerSec: speed,
				LaneMask:        0x3 << (2 * (uint(i) % 2)),
			}
		}
	case 96:
		if pm == 2 {
			// 2x1_1x2
			for i := range cf.Ports {
				if i%3 == 0 {
					cf.Ports[i] = bcm.PortConfig{
						PortBlockIndex:  uint(i) / 3,
						SpeedBitsPerSec: speed,
						LaneMask:        0x3 << 2,
					}
				} else if (i+2)%3 == 0 {
					cf.Ports[i] = bcm.PortConfig{
						PortBlockIndex:  (uint(i) - 1) / 3,
						SpeedBitsPerSec: speed / 2,
						LaneMask:        0x1 << 1,
					}
				} else if (i+1)%3 == 0 {
					cf.Ports[i] = bcm.PortConfig{
						PortBlockIndex:  (uint(i) - 2) / 3,
						SpeedBitsPerSec: speed / 2,
						LaneMask:        0x1 << 0,
					}
				}
			}
		} else if pm == 1 {
			// 1x2_2x1
			for i := range cf.Ports {
				if i%3 == 0 {
					cf.Ports[i] = bcm.PortConfig{
						PortBlockIndex:  uint(i) / 3,
						SpeedBitsPerSec: speed / 2,
						LaneMask:        0x1 << 3,
					}
				} else if (i+2)%3 == 0 {
					cf.Ports[i] = bcm.PortConfig{
						PortBlockIndex:  (uint(i) - 1) / 3,
						SpeedBitsPerSec: speed / 2,
						LaneMask:        0x1 << 2,
					}
				} else if (i+1)%3 == 0 {
					cf.Ports[i] = bcm.PortConfig{
						PortBlockIndex:  (uint(i) - 2) / 3,
						SpeedBitsPerSec: speed,
						LaneMask:        0x3 << 0,
					}
				}
			}
		} else {
			panic("Unsupported 2/1/1 configuration")
		}
	case 128:
		for i := range cf.Ports {
			cf.Ports[i] = bcm.PortConfig{
				PortBlockIndex:  uint(i) / 4,
				SpeedBitsPerSec: speed,
				LaneMask:        0x1 << (uint(i) % 4),
			}
		}
	}

	for i := range cf.Ports {
		cf.Ports[i].PhyInterface = bcm.PhyInterfaceOptics
	}

	// Management ports.
	for i := 0; i < 2; i++ {
		cf.Ports = append(cf.Ports, bcm.PortConfig{
			PortBlockIndex:  0,
			IsManagement:    true,
			SpeedBitsPerSec: 10e9,
			LaneMask:        1 << (2 * uint(i)),
			PhyInterface:    bcm.PhyInterfaceCR,
		})
	}

	cf.Phys = phyConfigs[:]
	for i := range cf.Phys {
		// For zion FC8-FC31 go to front panel ports 0-23; FC0-7 go to front panels ports 24-31
		cf.Phys[i].FrontPanelIndex = (cf.Phys[i].Index - 8) % 32
	}

	cf.Configure(p.Vnet, s)
	return
}
