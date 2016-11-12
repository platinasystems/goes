// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"

	"fmt"
)

func (phy *HundredGig) fu(tag string) { phy.fu2(tag, 0) }

func (phy *HundredGig) fu2(tag string, bi m.PortBlockIndex) {
	if bi != 99 && phy.PortBlock.PortBlockIndex != bi {
		return
	}

	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	mem := phy.get_uc_mem()

	type T struct {
		status           uint16
		top_user_control uint16
		config_word      uint16
		uc_status        uint8
		restarts         uint8
		resets           uint8
		pmd_locks        uint8
		link_status      uint16
		pll              uint16
	}
	type U struct {
		core  T
		lanes [4]T
	}

	u := &U{}

	lm := m.LaneMask(1 << 0)
	u.core.status = r.dig.core_datapath_reset_status.GetDo(q, lm)
	u.core.uc_status = mem.status_byte.Get(q, 1<<0)
	u.core.config_word = (*fuc_uint16)(&mem.config_word).Get(q, lm)
	laneMask := m.LaneMask(0xf << 0)
	laneMask.Foreach(func(l m.LaneMask) {
		lm := m.LaneMask(1 << l)
		u.lanes[l].status = r.clock_and_reset.lane_data_path_reset_status.GetDo(q, lm)
		u.lanes[l].uc_status = mem.lanes[l].uc_stopped.Get(q, lm)
		u.lanes[l].top_user_control = r.dig.top_user_control.GetDo(q, lm)
		u.lanes[l].config_word = (*fuc_uint16)(&mem.lanes[l].config_word).Get(q, lm)
		u.lanes[l].restarts = mem.lanes[l].restart_counter.Get(q, lm)
		u.lanes[l].resets = mem.lanes[l].reset_counter.Get(q, lm)
		u.lanes[l].pmd_locks = mem.lanes[l].pmd_lock_counter.Get(q, lm)
		u.lanes[l].link_status = r.rx_x4.pcs_live_status.GetDo(q, lm)
		u.lanes[l].pll = r.pll.multiplier.GetDo(q, lm)
	})
	fmt.Printf("%s %s: core status %04x uc status %02x config 0x%04x uc version 0x%04x\n",
		tag, phy.PortBlock.Name(),
		u.core.status, u.core.uc_status, u.core.config_word, mem.common_ucode_version.Get(q, lm))
	for i := range u.lanes {
		fmt.Printf("  lane %d: status %04x uc %02x top %04x config 0x%04x restarts %d resets %d pmd locks %d link_status %x pll 0x%x\n",
			i, u.lanes[i].status, u.lanes[i].uc_status, u.lanes[i].top_user_control, u.lanes[i].config_word,
			u.lanes[i].restarts, u.lanes[i].resets, u.lanes[i].pmd_locks,
			u.lanes[i].link_status, u.lanes[i].pll)
	}
}

func uc_control_cmd(q *DmaRequest, subCmd uc_sub_command) (err error) {
	r := get_hundred_gig_controller()
	c := uc_cmd{command: uc_cmd_control, sub_command: subCmd}
	err = c.do(q, 1<<0, &r.uc_cmd)
	return
}

type uc_event_code uint8

const (
	uc_unknown uc_event_code = iota
	uc_entry_to_dsc_reset
	uc_release_user_reset
	uc_exit_from_dsc_reset
	uc_entry_to_core_reset
	uc_release_user_core_reset
	uc_active_restart_condition
	uc_exit_from_restart
	uc_write_tr_coarse_lock
	uc_cl72_ready_for_command
	uc_each_write_to_cl72_tx_change_request
	uc_frame_lock
	uc_local_rx_trained
	uc_dsc_lock
	uc_first_rx_pmd_lock
	uc_pmd_restart_from_cl72_cmd_intf_timeout
	uc_lp_rx_ready
	uc_stop_event_log
	uc_general_event_0
	uc_general_event_1
	uc_general_event_2
	uc_error_event
	uc_num_timestamp_wraparound_maxout
	uc_restart_pmd_on_cdr_lock_lost
	uc_sm_status_restart
	uc_core_programming
	uc_lane_programming
	uc_restart_pmd_on_close_eye
	uc_restart_pmd_on_dfe_tap_config
	uc_cl72_auto_polarity_change
	uc_restart_from_cl72_max_timeout
	uc_cl72_local_tx_changed
	uc_first_write_to_cl72_tx_change_request
	uc_frame_unlock
	uc_entry_to_core_pll1_reset
	uc_release_user_core_pll1_reset
	uc_active_wait_for_sig
	uc_exit_from_wait_for_sig
	uc_start_vga_tuning
	uc_start_fx_dfe_tuning
	uc_start_pf_tuning
	uc_start_eye_meas_tuning
	uc_start_main_tap_tuning
	uc_start_fl_tap_tuning
	uc_exit_from_dsc_init
	uc_entry_to_dsc_init
	uc_dsc_pause
	uc_dsc_uc_tune
	uc_dsc_done
	uc_restart_pmd_on_short_channel_detected
)

var uc_event_code_strings = [...]string{
	uc_unknown:                                "uc_unknown",
	uc_entry_to_dsc_reset:                     "uc_entry_to_dsc_reset",
	uc_release_user_reset:                     "uc_release_user_reset",
	uc_exit_from_dsc_reset:                    "uc_exit_from_dsc_reset",
	uc_entry_to_core_reset:                    "uc_entry_to_core_reset",
	uc_release_user_core_reset:                "uc_release_user_core_reset",
	uc_active_restart_condition:               "uc_active_restart_condition",
	uc_exit_from_restart:                      "uc_exit_from_restart",
	uc_write_tr_coarse_lock:                   "uc_write_tr_coarse_lock",
	uc_cl72_ready_for_command:                 "uc_cl72_ready_for_command",
	uc_each_write_to_cl72_tx_change_request:   "uc_each_write_to_cl72_tx_change_request",
	uc_frame_lock:                             "uc_frame_lock",
	uc_local_rx_trained:                       "uc_local_rx_trained",
	uc_dsc_lock:                               "uc_dsc_lock",
	uc_first_rx_pmd_lock:                      "uc_first_rx_pmd_lock",
	uc_pmd_restart_from_cl72_cmd_intf_timeout: "uc_pmd_restart_from_cl72_cmd_intf_timeout",
	uc_lp_rx_ready:                            "uc_lp_rx_ready",
	uc_stop_event_log:                         "uc_stop_event_log",
	uc_general_event_0:                        "uc_general_event_0",
	uc_general_event_1:                        "uc_general_event_1",
	uc_general_event_2:                        "uc_general_event_2",
	uc_error_event:                            "uc_error_event",
	uc_num_timestamp_wraparound_maxout:        "uc_num_timestamp_wraparound_maxout",
	uc_restart_pmd_on_cdr_lock_lost:           "uc_restart_pmd_on_cdr_lock_lost",
	uc_sm_status_restart:                      "uc_sm_status_restart",
	uc_core_programming:                       "uc_core_programming",
	uc_lane_programming:                       "uc_lane_programming",
	uc_restart_pmd_on_close_eye:               "uc_restart_pmd_on_close_eye",
	uc_restart_pmd_on_dfe_tap_config:          "uc_restart_pmd_on_dfe_tap_config",
	uc_cl72_auto_polarity_change:              "uc_cl72_auto_polarity_change",
	uc_restart_from_cl72_max_timeout:          "uc_restart_from_cl72_max_timeout",
	uc_cl72_local_tx_changed:                  "uc_cl72_local_tx_changed",
	uc_first_write_to_cl72_tx_change_request:  "uc_first_write_to_cl72_tx_change_request",
	uc_frame_unlock:                           "uc_frame_unlock",
	uc_entry_to_core_pll1_reset:               "uc_entry_to_core_pll1_reset",
	uc_release_user_core_pll1_reset:           "uc_release_user_core_pll1_reset",
	uc_active_wait_for_sig:                    "uc_active_wait_for_sig",
	uc_exit_from_wait_for_sig:                 "uc_exit_from_wait_for_sig",
	uc_start_vga_tuning:                       "uc_start_vga_tuning",
	uc_start_fx_dfe_tuning:                    "uc_start_fx_dfe_tuning",
	uc_start_pf_tuning:                        "uc_start_pf_tuning",
	uc_start_eye_meas_tuning:                  "uc_start_eye_meas_tuning",
	uc_start_main_tap_tuning:                  "uc_start_main_tap_tuning",
	uc_start_fl_tap_tuning:                    "uc_start_fl_tap_tuning",
	uc_exit_from_dsc_init:                     "uc_exit_from_dsc_init",
	uc_entry_to_dsc_init:                      "uc_entry_to_dsc_init",
	uc_dsc_pause:                              "uc_dsc_pause",
	uc_dsc_uc_tune:                            "uc_dsc_uc_tune",
	uc_dsc_done:                               "uc_dsc_done",
	uc_restart_pmd_on_short_channel_detected:  "uc_restart_pmd_on_short_channel_detected",
}

type uc_event struct {
	code uc_event_code
	lane uint8
	data [4]byte
	time uint64
}

func (e *uc_event) String() string {
	return fmt.Sprintf("%.4e: lane %d, %s",
		float64(e.time)*1e-5, e.lane,
		elib.Stringer(uc_event_code_strings[:], int(e.code)))
}

func (phy *HundredGig) dump_event_log() {
	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	c := uc_cmd{command: uc_cmd_event_log_read}
	var err error
	defer func() {
		if err != nil {
			fmt.Printf("%s: %s\n", phy.PortBlock.Name(), err)
		}
	}()

	err = uc_control_cmd(q, uc_cmd_control_stop_immediate)
	if err != nil {
		return
	}

	laneMask := m.LaneMask(1 << 0)

	c.sub_command = uc_cmd_event_log_read_start
	err = c.do(q, laneMask, &r.uc_cmd)
	if err != nil {
		return
	}

	b := []byte{}
	n_zero := 0
	for {
		var v [2]uint16
		d := uc_cmd{
			command: uc_cmd_event_log_read, sub_command: uc_cmd_event_log_read_next,
			out:    &v[0],
			outAux: &v[1],
		}
		d.sub_command = uc_cmd_event_log_read_next
		err := d.do(q, laneMask, &r.uc_cmd)
		if err != nil {
			return
		}

		// 0 type implies end of log.
		if byte(v[0]) == 0 {
			n_zero++
		}

		b = append(b, byte(v[0]))

		if v[1] == 1 {
			break
		}
	}

	c.sub_command = uc_cmd_event_log_read_done
	err = c.do(q, laneMask, &r.uc_cmd)
	if err != nil {
		return
	}

	err = uc_control_cmd(q, uc_cmd_control_resume)
	if err != nil {
		return
	}

	es := []uc_event{}
	i := 0
	for {
		if i >= len(b) {
			break
		}
		x := b[i]
		if x == 0 {
			break
		}

		// Time wrap.
		if x == 0xff {
			if i+2 >= len(b) {
				break
			}
			y := b[i+1]<<8 | b[i+2]
			dt := uint64(y) * (1 << 16)
			for j := range es {
				es[j].time += dt
			}
			i += 3
		} else {
			nBytes := int(x >> 5)
			if i+nBytes-1 >= len(b) {
				break
			}
			e := uc_event{
				lane: uint8(x & 0x1f),
				time: uint64(b[i+1]<<8 | b[i+2]),
				code: uc_event_code(b[i+3]),
			}
			if nBytes > 4 {
				copy(e.data[:], b[i+4:i+nBytes])
			}
			es = append(es, e)
			i += nBytes
		}
	}

	fmt.Printf("%s: event log\n", phy.PortBlock.Name())
	l := len(es)
	for i := range es {
		fmt.Printf("  %s\n", &es[l-1-i])
	}

	return
}

func (ss *switchSelect) showPhyEventLog(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	portMap := make(map[int]bool)
	ss.SelectAll()
	for !in.End() {
		var port int
		switch {
		case in.Parse("ce%d", &port):
			portMap[port] = true
		case in.Parse("d%*ev %v", ss):
		default:
			err = cli.ParseError
			return
		}
	}

	for _, s := range ss.Switches {
		ps := s.GetPorts()
		for i := range ps {
			p := ps[i]
			if p.GetPortCommon().IsManagement {
				continue
			}
			if len(portMap) > 0 {
				if _, ok := portMap[i]; !ok {
					continue
				}
			}
			phy := p.GetPhy().(*HundredGig)
			phy.dump_event_log()
		}
	}
	return
}

type switchSelect struct{ m.SwitchSelect }

func Init(v *vnet.Vnet) {
	ss := &switchSelect{}
	ss.Vnet = v
	cmds := []cli.Command{
		cli.Command{
			Name:   "show fe1 phy event-log",
			Action: ss.showPhyEventLog,
		},
		cli.Command{
			Name:   "show fe1 eyescan",
			Action: ss.showEyeScan,
		},
		cli.Command{
			Name:   "show fe1 port-status phy",
			Action: ss.showPortStatus,
		},
	}
	for i := range cmds {
		v.CliAdd(&cmds[i])
	}
}
