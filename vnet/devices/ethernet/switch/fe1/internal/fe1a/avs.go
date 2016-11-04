// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file is used for the FE1 Adaptive Voltage Scaling (AVS)interface.
//
// The FE1 AVS is a power-saving technique of the digital 1.0V supply.
// And, it maintains performance under various operating conditions.
//
// Presently, the Platina Systems TOR will not use the AVS support.
// If that changes in the future, then this file can be updated to include
// the AVS support (will require interfacing to the IR3595 device) and implement
// the associated driver support. Until then, the FE1 1.0V supply is
// set to 1.0V (will not vary from 1.05V maximum to .85V minimum).
//
package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"unsafe"
)

type avs_reg uint32

func (r *avs_reg) set(q *DmaRequest, v uint32)  { q.SetReg32(v, BlockTop, r.address(), sbus.Unique0) }
func (r *avs_reg) get(q *DmaRequest, v *uint32) { q.GetReg32(v, BlockTop, r.address(), sbus.Unique0) }
func (r *avs_reg) getDo(q *DmaRequest) (v uint32) {
	r.get(q, &v)
	q.Do()
	return
}

func (r *avs_reg) address() sbus.Address { return sbus.GenReg | sbus.Address(r.offset()) }
func (r *avs_reg) offset() uint          { return uint(uintptr(unsafe.Pointer(r))-m.RegsBaseAddress) << 8 }

type avs_regs struct {
	sw_control                           avs_reg
	sw_measurement_unit_busy             avs_reg
	measurements_init_pvt_monitor        avs_reg
	measurements_init_cen_rosc           [2]avs_reg
	measurements_init_rmt_rosc           [4]avs_reg
	measurements_init_pow_wdog           avs_reg
	sequencer_init                       avs_reg
	sequencer_mask_pvt_monitor           avs_reg
	sequencer_mask_cen_rosc              [2]avs_reg
	sequencer_mask_rmt_rosc              [5]avs_reg
	enable_default_pvt_monitor           avs_reg
	enable_default_cen_rosc              [2]avs_reg
	rosc_measurement_time_control        avs_reg
	rosc_counting_mode                   avs_reg
	interrupt_pow_watchdog_enable        [3]avs_reg
	interrupt_sw_measurement_done_enable avs_reg
	last_measured_sensor                 avs_reg
	interrupt_status                     avs_reg
	interrupt_status_clear               avs_reg
	remote_sensor_type                   [3]avs_reg
	avs_registers_locks                  avs_reg
	temperature_reset_enable             avs_reg
	temperature_threshold                avs_reg
	idle_state_0_cen_rosc                [2]avs_reg
	adc_settling_time                    avs_reg
	avs_spare                            [2]avs_reg
}
