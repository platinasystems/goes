// Copyright © 2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mk1

import (
	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/sriovs"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/bus/pci"
	"github.com/platinasystems/go/vnet/devices/ethernet/ixge"
	fe1_plugin "github.com/platinasystems/go/vnet/devices/ethernet/switch/plugins/fe1"
	"github.com/platinasystems/go/vnet/ethernet"
	ipcli "github.com/platinasystems/go/vnet/ip/cli"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
	"github.com/platinasystems/go/vnet/pg"
	fe1_platform "github.com/platinasystems/go/vnet/platforms/fe1"
	"github.com/platinasystems/go/vnet/unix"

	"fmt"
	"net"
	"time"
)

func newSriovs(ver uint) error {
	if ver > 0 {
		sriovs.VfName = func(port, subport uint) string {
			return fmt.Sprintf("eth-%d-%d", port+1, subport+1)
		}
	}
	eth0, err := net.InterfaceByName("eth0")
	if err != nil {
		return err
	}
	mac := sriovs.Mac(eth0.HardwareAddr)
	// skip over eth0, eth1, and eth2
	mac.Plus(3)
	sriovs.VfMac = mac.VfMac
	return sriovs.New(vfs)
}

func delSriovs() error { return sriovs.Del(vfs) }

func vlan_for_port(port, subport sriovs.Vf) (vf sriovs.Vf) {
	// physical port number for data ports are numbered starting at 1.
	// (phys 0 is cpu port...)
	phys := sriovs.Vf(1)

	// 4 sub-ports per port; mk1 ports are even/odd swapped.
	phys += 4 * (port ^ 1)

	phys += subport

	// Vlan is 1 plus physical port number.
	return sriovs.Vf(1 + phys)
}

// The vfs table is 0 based and is adjusted to 1 based beta and production
// units with VfName
var vfs = make_vfs()

func make_vfs() [][]sriovs.Vf {
	// pf0 = fe1 pipes 0 & 1; only 32 vfs.
	// pf1 = fe1 pipes 2 & 3; only 32 vfs.
	const (
		n_port        = 32
		n_sub_port    = 1
		n_pf          = 2
		n_port_per_pf = n_port / n_pf
	)
	var pfs [n_pf][n_port / n_pf]sriovs.Vf
	for port := sriovs.Vf(0); port < n_port; port++ {
		for subport := sriovs.Vf(0); subport < n_sub_port; subport++ {
			vf := port<<sriovs.PortShift | subport<<sriovs.SubPortShift | vlan_for_port(port, subport)
			pf := port / n_port_per_pf
			i := port%n_port_per_pf + n_port_per_pf*subport
			if i < sriovs.Vf(len(pfs[pf])) {
				pfs[pf][i] = vf
			}
		}
	}
	return [][]sriovs.Vf{pfs[0][:], pfs[1][:]}
}

func PlatformInit(v *vnet.Vnet, p *fe1_platform.Platform) (err error) {
	fns, err := sriovs.NumvfsFns()
	p.SriovMode = err == nil && len(fns) > 0
	err = nil

	// Select packages we want to run with.
	m4 := ip4.Init(v)
	m6 := ip6.Init(v)
	ethernet.Init(v, m4, m6)
	if !p.SriovMode {
		ixge.Init(v, ixge.Config{DisableUnix: true, PuntNode: "fe1-single-tagged-punt"})
	} else if err = newSriovs(p.Version); err != nil {
		return
	}
	pci.Init(v)
	pg.Init(v)
	ipcli.Init(v)
	unix.Init(v)

	gpio := pca9535_main{
		bus_index:   0,
		bus_address: 0x74,
	}
	if err = gpio.do(gpio.led_output_enable); err != nil {
		return
	}
	if !p.DisableGpioSwitchReset {
		if err = gpio.do(gpio.switch_reset); err != nil {
			return
		}
	}

	fe1_plugin.Init(v, p)
	fe1_plugin.AddPlatform(v, p)

	return
}

func PlatformExit(v *vnet.Vnet, p *fe1_platform.Platform) (err error) {
	if p.SriovMode {
		if err = delSriovs(); err != nil {
			return
		}
	}
	return
}

type pca9535_main struct {
	bus_index, bus_address int
}

func (m *pca9535_main) do(f func(bus *i2c.Bus) error) (err error) {
	var bus i2c.Bus
	if err = bus.Open(m.bus_index); err != nil {
		return
	}
	defer bus.Close()
	if err = bus.ForceSlaveAddress(m.bus_address); err != nil {
		return
	}
	return f(&bus)
}

const (
	pca9535_reg_input_0  = iota // read-only input bits [7:0]
	pca9535_reg_input_1         // read-only input bits [15:8]
	pca9535_reg_output_0        // output bits [7:0] (default 1)
	pca9535_reg_output_1        // output [15:8]
	pca9535_reg_invert_polarity_0
	pca9535_reg_invert_polarity_1
	pca9535_reg_is_input_0 // 1 => pin is input; 0 => output
	pca9535_reg_is_input_1 // defaults are 1 (pin is input)
)

// MK1 pin usage.
const (
	mk1_pca9535_pin_switch_reset = 1 << iota
	_
	mk1_pca9535_pin_led_output_enable
)

// MK1 board front panel port LED's require PCA9535 GPIO device configuration
// to provide an output signal that allows LED operation.
func (m *pca9535_main) led_output_enable(bus *i2c.Bus) (err error) {
	var d i2c.SMBusData
	// Set pin to output (default is input and default value is high which we assume).
	if err = bus.Read(pca9535_reg_is_input_0, i2c.ByteData, &d); err != nil {
		return
	}
	d[0] &^= mk1_pca9535_pin_led_output_enable
	return bus.Write(pca9535_reg_is_input_0, i2c.ByteData, &d)
}

// Hard reset switch via gpio pins on MK1 board.
func (m *pca9535_main) switch_reset(bus *i2c.Bus) (err error) {
	const reset_bits = mk1_pca9535_pin_switch_reset

	var val, dir i2c.SMBusData

	// Set direction to output.
	if err = bus.Read(pca9535_reg_is_input_0, i2c.ByteData, &dir); err != nil {
		return
	}
	if dir[0]&reset_bits != 0 {
		dir[0] &^= reset_bits
		if err = bus.Write(pca9535_reg_is_input_0, i2c.ByteData, &dir); err != nil {
			return
		}
	}

	// Set output low & wait 2 us minimum.
	if err = bus.Read(pca9535_reg_output_0, i2c.ByteData, &val); err != nil {
		return
	}
	val[0] &^= reset_bits
	if err = bus.Write(pca9535_reg_output_0, i2c.ByteData, &val); err != nil {
		return
	}
	time.Sleep(2 * time.Microsecond)

	// Set output hi & wait 2 ms minimum before pci activity.
	val[0] |= reset_bits
	if err = bus.Write(pca9535_reg_output_0, i2c.ByteData, &val); err != nil {
		return
	}
	// Need to wait a long time else the switch does not show up in pci bus and pci discovery fails.
	time.Sleep(100 * time.Millisecond)

	return
}
