// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"

	"fmt"
	"time"
	"unsafe"
)

type fuc_uint16 uint16
type fuc_uint8 uint8
type euc_uint16 uint16
type euc_uint8 uint8

func (phy *Tsce) get_uc_mem() *tsce_uc_mem { return (*tsce_uc_mem)(m.RegsBasePointer) }
func (phy *Tscf) get_uc_mem() *tscf_uc_mem { return (*tscf_uc_mem)(m.RegsBasePointer) }

func getSetTscf(q *DmaRequest, laneMask m.LaneMask, r unsafe.Pointer, write_data *uint32, log2NBytes int) (read_data uint32) {
	addr := uint32(uintptr(r) - m.RegsBaseAddress)
	addr |= 0x20000000 // ARM AHB device address space.

	regs := get_tscf_regs()

	// Set read/write data size.
	m, v := uint16(3), uint16(log2NBytes)
	if write_data == nil {
		m, v = m<<4, v<<4
	}

	regs.uc.ahb_control.Modify(q, laneMask, v, m)

	if write_data != nil {
		regs.uc.write_address.Set(q, laneMask, addr)
		regs.uc.write_data.Set(q, laneMask, *write_data)
		read_data = *write_data
	} else {
		regs.uc.read_address.Set(q, laneMask, addr)
		regs.uc.read_data.Get(q, laneMask, &read_data)
	}
	q.Do()

	return
}

func getSetTsce(q *DmaRequest, laneMask m.LaneMask, r unsafe.Pointer, write_data *uint32, log2NBytes int) (read_data uint32) {
	addr := uint16(uintptr(r) - m.RegsBaseAddress)

	regs := get_tsce_regs()

	// Select data memory access mode thru mdio
	regs.uc.command1.Modify(q, laneMask, 2<<7, 3<<7)

	// Access-width selection - 0 for word; 1 for byte
	if log2NBytes > 0 {
		regs.uc.command1.Modify(q, laneMask, 0<<9, 1<<9)
	} else {
		regs.uc.command1.Modify(q, laneMask, 1<<9, 1<<9)
	}

	regs.uc.address.Set(q, laneMask, addr)

	if write_data != nil {
		regs.uc.write_data.Set(q, laneMask, uint16(*write_data))
		read_data = *write_data
	} else {
		var tmp uint16
		regs.uc.read_data.Get(q, laneMask, &tmp)
		read_data = uint32(tmp)
	}
	q.Do()

	return
}

func (r *euc_uint16) Get(q *DmaRequest, laneMask m.LaneMask) uint16 {
	return uint16(getSetTsce(q, laneMask, unsafe.Pointer(r), nil, 1))
}

func (r *euc_uint16) Set(q *DmaRequest, laneMask m.LaneMask, v uint16) {
	u := uint32(v)
	getSetTsce(q, laneMask, unsafe.Pointer(r), &u, 1)
}

func (r *euc_uint8) Get(q *DmaRequest, laneMask m.LaneMask) uint8 {
	return uint8(getSetTsce(q, laneMask, unsafe.Pointer(r), nil, 0))
}

func (r *euc_uint8) Set(q *DmaRequest, laneMask m.LaneMask, v uint8) {
	u := uint32(v)
	getSetTsce(q, laneMask, unsafe.Pointer(r), &u, 0)
}

func (r *fuc_uint16) Get(q *DmaRequest, laneMask m.LaneMask) uint16 {
	return uint16(getSetTscf(q, laneMask, unsafe.Pointer(r), nil, 1))
}

func (r *fuc_uint16) Set(q *DmaRequest, laneMask m.LaneMask, v uint16) {
	u := uint32(v)
	getSetTscf(q, laneMask, unsafe.Pointer(r), &u, 1)
}

func (r *fuc_uint8) Get(q *DmaRequest, laneMask m.LaneMask) uint8 {
	return uint8(getSetTscf(q, laneMask, unsafe.Pointer(r), nil, 0))
}

func (r *fuc_uint8) Set(q *DmaRequest, laneMask m.LaneMask, v uint8) {
	u := uint32(v)
	getSetTscf(q, laneMask, unsafe.Pointer(r), &u, 0)
}

type tsce_uc_mem struct {
	_ [0x50]byte

	config_word tsce_uc_core_config_word_reg

	temp_frc_val               euc_uint16
	common_ucode_version       euc_uint16
	avg_tmon_reg13bit          euc_uint16
	trace_mem_rd_idx           euc_uint16
	trace_mem_wr_idx           euc_uint16
	temp_idx                   euc_uint8
	event_log_level            euc_uint8
	common_ucode_minor_version euc_uint8
	afe_hardware_version       euc_uint8
	status_byte                euc_uint8
	diag_max_time_control      euc_uint8
	diag_max_err_control       euc_uint8

	_ [0x400 - 0x63]byte

	lanes [4]struct {
		config_word tsce_uc_lane_config_word_reg

		retune_after_restart               euc_uint8
		clk90_offset_adjust                euc_uint8
		clk90_offset_override              euc_uint8
		event_log_level                    euc_uint8
		disable_startup_functions          euc_uint8
		disable_dfe_startup_functions      euc_uint8
		disable_steady_state_functions     euc_uint8
		disable_dfe_steady_state_functions euc_uint8
		restart_counter                    euc_uint8
		reset_counter                      euc_uint8
		pmd_lock_counter                   euc_uint8
		horizontal_eye_left                euc_uint8
		horizontal_eye_right               euc_uint8
		vertical_eye_upper                 euc_uint8
		vertical_eye_lower                 euc_uint8
		uc_stopped                         euc_uint8
		link_time                          euc_uint16

		diag_status euc_uint16

		diag_read_pointer euc_uint8
		diag_mode         euc_uint8
		usr_var           [2]euc_uint16

		_ [0x100 - 0x1c]byte
	}
}

type tscf_uc_mem struct {
	_ [0x400]byte

	config_word tscf_uc_core_config_word_reg

	temp_frc_val               fuc_uint16
	common_ucode_version       fuc_uint16
	avg_tmon_reg13bit          fuc_uint16
	trace_mem_rd_idx           fuc_uint16
	trace_mem_wr_idx           fuc_uint16
	temp_idx                   fuc_uint8
	event_log_level            fuc_uint8
	common_ucode_minor_version fuc_uint8
	afe_hardware_version       fuc_uint8
	status_byte                fuc_uint8
	diag_max_time_control      fuc_uint8
	diag_max_err_control       fuc_uint8
	misc_ctrl_byte             fuc_uint8
	config_pll1_word           fuc_uint16

	_ [0x420 - 0x416]byte

	lanes [4]struct {
		config_word tscf_uc_lane_config_word_reg

		retune_after_restart               fuc_uint8
		clk90_offset_adjust                fuc_uint8
		clk90_offset_override              fuc_uint8
		event_log_level                    fuc_uint8
		dummy                              fuc_uint8
		cl93_frc_byte                      fuc_uint8
		disable_startup_functions          fuc_uint16
		disable_steady_state_functions     fuc_uint16
		disable_dfe_startup_functions      fuc_uint8
		disable_dfe_steady_state_functions fuc_uint8
		restart_counter                    fuc_uint8
		reset_counter                      fuc_uint8
		pmd_lock_counter                   fuc_uint8
		horizontal_eye_left                fuc_uint8
		horizontal_eye_right               fuc_uint8
		vertical_eye_upper                 fuc_uint8
		vertical_eye_lower                 fuc_uint8
		uc_stopped                         fuc_uint8
		link_time                          fuc_uint16

		diag_status fuc_uint16

		diag_read_pointer       fuc_uint8
		diag_mode               fuc_uint8
		main_tap_estimate       fuc_uint16
		phase_horizontal_offset fuc_uint8

		_ [0x130 - 0x1f]byte
	}
}

// Microcontroller commands.
type uc_command uint8
type uc_sub_command uint8

const (
	uc_cmd_null uc_command = iota
	uc_cmd_control
	uc_cmd_horizontal_eye_offset
	uc_cmd_vertical_eye_offset
	uc_cmd_debug
	uc_cmd_diagnostics
	uc_cmd_read_lane_byte
	uc_cmd_write_lane_byte
	uc_cmd_read_core_byte
	uc_cmd_write_core_byte
	uc_cmd_read_lane_word
	uc_cmd_write_lane_word
	uc_cmd_read_core_word
	uc_cmd_write_core_word
	uc_cmd_event_log_control
	uc_cmd_event_log_read
	uc_cmd_capture_bit_error_rate_start
	uc_cmd_read_diagnostic_data_byte
	uc_cmd_read_diagnostic_data_word
	uc_cmd_capture_bit_error_rate_end
	uc_cmd_compute_ucode_crc
	uc_cmd_freeze_steady_state
)

var uc_cmd_strings = []string{
	uc_cmd_null:                         "null",
	uc_cmd_control:                      "control",
	uc_cmd_horizontal_eye_offset:        "horizontal eye offset",
	uc_cmd_vertical_eye_offset:          "vertical eye offset",
	uc_cmd_debug:                        "debug",
	uc_cmd_diagnostics:                  "diagnostics",
	uc_cmd_read_lane_byte:               "read lane byte",
	uc_cmd_write_lane_byte:              "write lane byte",
	uc_cmd_read_core_byte:               "read core byte",
	uc_cmd_write_core_byte:              "write core byte",
	uc_cmd_read_lane_word:               "read lane word",
	uc_cmd_write_lane_word:              "write lane word",
	uc_cmd_read_core_word:               "read core word",
	uc_cmd_write_core_word:              "write core word",
	uc_cmd_event_log_control:            "event log control",
	uc_cmd_event_log_read:               "event log read",
	uc_cmd_capture_bit_error_rate_start: "capture bit error rate start",
	uc_cmd_read_diagnostic_data_byte:    "read diagnostic data byte",
	uc_cmd_read_diagnostic_data_word:    "read diagnostic data word",
	uc_cmd_capture_bit_error_rate_end:   "capture bit error rate end",
	uc_cmd_compute_ucode_crc:            "compute ucode crc",
	uc_cmd_freeze_steady_state:          "freeze steady state",
}

// uc_cmd_control sub commands.
const (
	uc_cmd_control_stop_gracefully uc_sub_command = iota
	uc_cmd_control_stop_immediate
	uc_cmd_control_resume
)

// uc_cmd_diagnostics sub commands.
const (
	uc_cmd_diagnostics_none uc_sub_command = iota
	uc_cmd_diagnostics_passive
	uc_cmd_diagnostics_density
	uc_cmd_diagnostics_disable
	uc_cmd_diagnostics_start_vertical_eye_scan
	uc_cmd_diagnostics_start_horizontal_eye_scan
	uc_cmd_diagnostics_get_eye_sample
)

// uc_cmd_event_log_read sub commands.
const (
	uc_cmd_event_log_read_start uc_sub_command = iota
	uc_cmd_event_log_read_next
	uc_cmd_event_log_read_done
)

// uc_cmd_debug sub commands.
const (
	uc_cmd_debug_die_temperature uc_sub_command = iota
	uc_cmd_debug_timestamp
	uc_cmd_debug_lane_index
	uc_cmd_debug_lane_timer
)

type uc_cmd struct {
	command     uc_command
	sub_command uc_sub_command
	timeout     time.Duration
	in, out     *uint16
	outAux      *uint16
}

func (c *uc_cmd) do(q *DmaRequest, laneMask m.LaneMask, r *uc_cmd_regs) (err error) {
	if c.in != nil {
		r.data.Set(q, laneMask, *c.in)
	}

	timeout := c.timeout
	if timeout == 0 {
		timeout = 100 * time.Millisecond
	}

	error_tag := ""
	defer func() {
		if len(error_tag) > 0 {
			err = fmt.Errorf("uc command %s: %s (sub command %d)", error_tag, uc_cmd_strings[c.command], c.sub_command)
		}
	}()

	control := uint16(c.command) & 0x3f
	control |= uint16(c.sub_command) << 8
	r.control.Set(q, laneMask, control)
	q.Do()

	start := time.Now()
	for {
		time.Sleep(100 * time.Microsecond)
		s := r.control.GetDo(q, laneMask)
		if s&(1<<7) != 0 {
			if s&(1<<6) != 0 {
				error_tag = "error"
			} else if c.out != nil {
				*c.out = r.data.GetDo(q, laneMask)
			}
			if c.outAux != nil {
				*c.outAux = uint16(s >> 8)
			}
			return
		}
		if time.Since(start) > timeout {
			error_tag = "timeout"
			return
		}
	}
}

type tscf_uc_core_config_word_reg fuc_uint16
type tsce_uc_core_config_word_reg euc_uint16

type uc_core_config_word struct {
	vco_rate_in_hz                        float64
	core_config_from_pcs                  bool
	disable_write_pll_charge_pump_current bool
}

const (
	tsce_vco_rate_unit = 250e6
	tscf_vco_rate_unit = 62.5e6
)

func (w *uc_core_config_word) fromTscfUint16(v uint16) {
	w.vco_rate_in_hz = tscf_vco_rate_unit * float64(224+(v&0xff))
	w.core_config_from_pcs = v&(1<<8) != 0
	w.disable_write_pll_charge_pump_current = v&(1<<9) != 0
}

func (w *uc_core_config_word) toTscfUint16() (v uint16) {
	v = (uint16(w.vco_rate_in_hz/tscf_vco_rate_unit+.5) - 224) & 0xff
	if w.core_config_from_pcs {
		v |= 1 << 8
	}
	if w.disable_write_pll_charge_pump_current {
		v |= 1 << 9
	}
	return
}

func (w *uc_core_config_word) fromTsceUint16(v uint16) {
	w.core_config_from_pcs = v&(1<<0) != 0
	w.vco_rate_in_hz = tsce_vco_rate_unit * float64(22+((v>>1)&0x1f))
}

func (w *uc_core_config_word) toTsceUint16() (v uint16) {
	if w.core_config_from_pcs {
		v |= 1 << 0
	}
	v |= ((uint16(w.vco_rate_in_hz/tsce_vco_rate_unit+.5) - 22) & 0x1f) << 1
	return
}

func (r *tscf_uc_core_config_word_reg) Get(q *DmaRequest, laneMask m.LaneMask) (w uc_core_config_word) {
	v := (*fuc_uint16)(r).Get(q, laneMask)
	w.fromTscfUint16(v)
	return
}

func (r *tscf_uc_core_config_word_reg) Set(q *DmaRequest, laneMask m.LaneMask, w *uc_core_config_word) {
	v := w.toTscfUint16()
	(*fuc_uint16)(r).Set(q, laneMask, v)
}

func (r *tsce_uc_core_config_word_reg) Get(q *DmaRequest, laneMask m.LaneMask) (w uc_core_config_word) {
	v := (*euc_uint16)(r).Get(q, laneMask)
	w.fromTsceUint16(v)
	return
}

func (r *tsce_uc_core_config_word_reg) Set(q *DmaRequest, laneMask m.LaneMask, w *uc_core_config_word) {
	v := w.toTsceUint16()
	(*euc_uint16)(r).Set(q, laneMask, v)
}

type tscf_uc_lane_config_word_reg fuc_uint16
type tsce_uc_lane_config_word_reg euc_uint16

type uc_lane_config_media_type uint8

const (
	uc_lane_config_media_backplane uc_lane_config_media_type = iota
	uc_lane_config_media_copper_cable
	uc_lane_config_media_optics
)

type uc_lane_config_word struct {
	cl72_restart_timeout_enable bool
	cl72_auto_polarity_enable   bool
	scrambling_disable          bool
	unreliable_los              bool
	force_br_dfe_on             bool
	dfe_low_power_mode          bool
	dfe_on                      bool
	autoneg_enable              bool
	lane_config_from_pcs        bool
	media_type                  uc_lane_config_media_type
}

func (w *uc_lane_config_word) fromTscfUint16(v uint16) {
	if v&(1<<10) != 0 {
		w.cl72_restart_timeout_enable = true
	}
	if v&(1<<9) != 0 {
		w.cl72_auto_polarity_enable = true
	}
	if v&(1<<8) != 0 {
		w.scrambling_disable = true
	}
	if v&(1<<7) != 0 {
		w.unreliable_los = true
	}
	w.media_type = uc_lane_config_media_type((v >> 5) & 3)
	if v&(1<<4) != 0 {
		w.force_br_dfe_on = true
	}
	if v&(1<<3) != 0 {
		w.dfe_low_power_mode = true
	}
	if v&(1<<2) != 0 {
		w.dfe_on = true
	}
	if v&1<<1 != 0 {
		w.autoneg_enable = true
	}
	if v&(1<<0) != 0 {
		w.lane_config_from_pcs = true
	}
	return
}

func (w *uc_lane_config_word) toTscfUint16() (v uint16) {
	v = uint16(0)
	if w.cl72_restart_timeout_enable {
		v |= 1 << 10
	}
	if w.cl72_auto_polarity_enable {
		v |= 1 << 9
	}
	if w.scrambling_disable {
		v |= 1 << 8
	}
	if w.unreliable_los {
		v |= 1 << 7
	}
	v |= uint16(w.media_type) << 5
	if w.force_br_dfe_on {
		v |= 1 << 4
	}
	if w.dfe_low_power_mode {
		v |= 1 << 3
	}
	if w.dfe_on {
		v |= 1 << 2
	}
	if w.autoneg_enable {
		v |= 1 << 1
	}
	if w.lane_config_from_pcs {
		v |= 1 << 0
	}
	return
}

func (r *tscf_uc_lane_config_word_reg) Get(q *DmaRequest, laneMask m.LaneMask) (w uc_lane_config_word) {
	v := (*fuc_uint16)(r).Get(q, laneMask)
	w.fromTscfUint16(v)
	return
}

func (r *tscf_uc_lane_config_word_reg) Set(q *DmaRequest, laneMask m.LaneMask, w *uc_lane_config_word) {
	v := w.toTscfUint16()
	(*fuc_uint16)(r).Set(q, laneMask, v)
}

func (w *uc_lane_config_word) fromTsceUint16(v uint16) {
	if v&(1<<9) != 0 {
		w.cl72_restart_timeout_enable = true
	}
	if v&(1<<8) != 0 {
		w.cl72_auto_polarity_enable = true
	}
	if v&(1<<7) != 0 {
		w.scrambling_disable = true
	}
	if v&(1<<6) != 0 {
		w.unreliable_los = true
	}
	w.media_type = uc_lane_config_media_type((v >> 4) & 3)
	if v&(1<<3) != 0 {
		w.force_br_dfe_on = true
	}
	if v&(1<<2) != 0 {
		w.dfe_on = true
	}
	if v&(1<<1) != 0 {
		w.autoneg_enable = true
	}
	if v&(1<<0) != 0 {
		w.lane_config_from_pcs = true
	}
	return
}

func (w *uc_lane_config_word) toTsceUint16() (v uint16) {
	v = uint16(0)
	if w.cl72_restart_timeout_enable {
		v |= 1 << 9
	}
	if w.cl72_auto_polarity_enable {
		v |= 1 << 8
	}
	if w.scrambling_disable {
		v |= 1 << 7
	}
	if w.unreliable_los {
		v |= 1 << 6
	}
	v |= uint16(w.media_type) << 4
	if w.force_br_dfe_on {
		v |= 1 << 3
	}
	if w.dfe_on {
		v |= 1 << 2
	}
	if w.autoneg_enable {
		v |= 1 << 1
	}
	if w.lane_config_from_pcs {
		v |= 1 << 0
	}
	return
}

func (r *tsce_uc_lane_config_word_reg) Get(q *DmaRequest, laneMask m.LaneMask) (w uc_lane_config_word) {
	v := (*euc_uint16)(r).Get(q, laneMask)
	w.fromTsceUint16(v)
	return
}

func (r *tsce_uc_lane_config_word_reg) Set(q *DmaRequest, laneMask m.LaneMask, w *uc_lane_config_word) {
	v := w.toTsceUint16()
	(*euc_uint16)(r).Set(q, laneMask, v)
}
