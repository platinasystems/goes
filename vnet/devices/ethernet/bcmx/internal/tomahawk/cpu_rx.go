// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

const (
	n_cpu_cos_map = 128
)

type cpu_rx_reason uint8

const (
	cpu_rx_reason_internal_pri                    cpu_rx_reason = 0 // 0-3
	cpu_rx_reason_non_unicast                     cpu_rx_reason = 4
	cpu_rx_reason_switched                        cpu_rx_reason = 5
	cpu_rx_reason_is_mirror                       cpu_rx_reason = 6
	cpu_rx_reason_bpdu                            cpu_rx_reason = 7
	cpu_rx_reason_ip_multicast_reserved           cpu_rx_reason = 8
	cpu_rx_reason_dhcp                            cpu_rx_reason = 9
	cpu_rx_reason_igmp                            cpu_rx_reason = 10
	cpu_rx_reason_arp                             cpu_rx_reason = 11
	cpu_rx_reason_unknown_vlan                    cpu_rx_reason = 12
	cpu_rx_reason_private_vlan_mismatch           cpu_rx_reason = 13
	cpu_rx_reason_dos_attack                      cpu_rx_reason = 14
	cpu_rx_reason_parity_error                    cpu_rx_reason = 15
	cpu_rx_reason_higig_control                   cpu_rx_reason = 16
	cpu_rx_reason_ip_ttl_check_fails              cpu_rx_reason = 17
	cpu_rx_reason_ip_options_present              cpu_rx_reason = 18
	cpu_rx_reason_l2_src_miss                     cpu_rx_reason = 19
	cpu_rx_reason_l2_dst_miss                     cpu_rx_reason = 20
	cpu_rx_reason_l2_move                         cpu_rx_reason = 21
	cpu_rx_reason_l2_copy_to_cpu_set              cpu_rx_reason = 22
	cpu_rx_reason_l2_non_unicast_miss             cpu_rx_reason = 23
	cpu_rx_reason_l3_src_miss                     cpu_rx_reason = 24
	cpu_rx_reason_l3_dst_miss                     cpu_rx_reason = 25
	cpu_rx_reason_l3_src_move                     cpu_rx_reason = 26
	cpu_rx_reason_multicast_miss                  cpu_rx_reason = 27
	cpu_rx_reason_ip_multicast_miss               cpu_rx_reason = 28
	cpu_rx_reason_l3_header_error                 cpu_rx_reason = 29
	cpu_rx_reason_martian_address                 cpu_rx_reason = 30
	cpu_rx_reason_tunnel_error                    cpu_rx_reason = 31
	cpu_rx_reason_higig_header_error              cpu_rx_reason = 32
	cpu_rx_reason_multicast_index_error           cpu_rx_reason = 33
	cpu_rx_reason_vfp_match                       cpu_rx_reason = 34
	cpu_rx_reason_class_based_move                cpu_rx_reason = 35
	cpu_rx_reason_congestion_cnm                  cpu_rx_reason = 36
	cpu_rx_reason_e2e_hol_ibp                     cpu_rx_reason = 37
	cpu_rx_reason_class_tag_packets               cpu_rx_reason = 38
	cpu_rx_reason_next_hop                        cpu_rx_reason = 39
	cpu_rx_reason_src_rpf_check_fail              cpu_rx_reason = 40
	cpu_rx_reason_filter_match                    cpu_rx_reason = 41
	cpu_rx_reason_icmp_redirect                   cpu_rx_reason = 42
	cpu_rx_reason_sflow_sample_src                cpu_rx_reason = 43
	cpu_rx_reason_sflow_sample_dst                cpu_rx_reason = 44
	cpu_rx_reason_l3_mtu_exceeded                 cpu_rx_reason = 45
	cpu_rx_reason_mpls_label_miss                 cpu_rx_reason = 46
	cpu_rx_reason_mpls_invalid_action             cpu_rx_reason = 47
	cpu_rx_reason_mpls_invalid_payload            cpu_rx_reason = 48
	cpu_rx_reason_mpls_ttl_check_fail             cpu_rx_reason = 49
	cpu_rx_reason_mpls_sequence_number_failure    cpu_rx_reason = 50
	cpu_rx_reason_mpls_unknown_ach                cpu_rx_reason = 51
	cpu_rx_reason_mmrp                            cpu_rx_reason = 52
	cpu_rx_reason_srp                             cpu_rx_reason = 53
	cpu_rx_reason_time_sync_unknown_version       cpu_rx_reason = 54
	cpu_rx_reason_mpls_router_alert_label         cpu_rx_reason = 55
	cpu_rx_reason_mpls_illegal_reserved_label     cpu_rx_reason = 56
	cpu_rx_reason_vlan_translate_miss             cpu_rx_reason = 57
	cpu_rx_reason_amt_tunnel_control              cpu_rx_reason = 58
	cpu_rx_reason_time_sync                       cpu_rx_reason = 59
	cpu_rx_reason_oam_slowpath                    cpu_rx_reason = 60
	cpu_rx_reason_oam_error                       cpu_rx_reason = 61
	cpu_rx_reason_l2_marked                       cpu_rx_reason = 62
	cpu_rx_reason_l3_mac_addr_bind_fail           cpu_rx_reason = 63
	cpu_rx_reason_my_station_copy_to_cpu          cpu_rx_reason = 64
	cpu_rx_reason_niv_drop_reason                 cpu_rx_reason = 65 // 3 bits
	cpu_rx_reason_trill_drop_reason               cpu_rx_reason = 68 // 3 bits
	cpu_rx_reason_l2_gre_src_ip_miss              cpu_rx_reason = 71
	cpu_rx_reason_l2_gre_vpnid_miss               cpu_rx_reason = 72
	cpu_rx_reason_bfd_slowpath                    cpu_rx_reason = 73
	cpu_rx_reason_bfd_error                       cpu_rx_reason = 74
	cpu_rx_reason_oam_lmdm                        cpu_rx_reason = 75
	cpu_rx_reason_congestion_cnm_proxy            cpu_rx_reason = 76
	cpu_rx_reason_congestion_cnm_proxy_error      cpu_rx_reason = 77
	cpu_rx_reason_vxlan_src_ip_miss               cpu_rx_reason = 78
	cpu_rx_reason_vxlan_vpnid_miss                cpu_rx_reason = 79
	cpu_rx_reason_fcoe_zone_check_fail            cpu_rx_reason = 80
	cpu_rx_reason_nat_drop_reason                 cpu_rx_reason = 81 // 3 bits
	cpu_rx_reason_ip_multicast_interface_mismatch cpu_rx_reason = 84
	cpu_rx_reason_sflow_sample_source_flex        cpu_rx_reason = 85
)

type cpu_cos_map_key [5]uint32

func (e *cpu_cos_map_key) getSetReason(r cpu_rx_reason, v *uint8, nbits int, isSet bool) {
	m.MemGetSetUint8(v, e[:], int(r)+nbits-1, int(r), isSet)
}

func (e *cpu_cos_map_key) getSetReason1(r cpu_rx_reason, v bool, isSet bool) bool {
	var i uint8
	if isSet && v {
		i = 1
	}
	e.getSetReason(r, &i, 1, isSet)
	if !isSet {
		v = i != 0
	}
	return v
}
func (e *cpu_cos_map_key) getReason(r cpu_rx_reason) bool    { return e.getSetReason1(r, false, false) }
func (e *cpu_cos_map_key) setReason(r cpu_rx_reason, v bool) { e.getSetReason1(r, v, true) }

func (e *cpu_cos_map_key) memGetSet(b []uint32, i int, isSet bool) int {
	for j := 0; j < 4; j++ {
		i = m.MemGetSetUint32(&e[j], b, i+31, i, isSet)
	}
	i = m.MemGetSetUint32(&e[4], b, i+11, i, isSet) // last 12 bits
	return i
}

func (key *cpu_cos_map_key) tcamEncode(mask *cpu_cos_map_key, isSet bool) (x, y cpu_cos_map_key) {
	for j := range key {
		a, b := m.TcamUint32(key[j]).TcamEncode(m.TcamUint32(mask[j]), isSet)
		x[j], y[j] = uint32(a), uint32(b)
	}
	return
}

type cpu_cos_map_data struct {
	// cpu queue number: 0-47
	cpu_cos_queue              uint8
	rqe_queue_number           bool
	truncate_copy_to_144_bytes bool
	strength                   uint8
}

func (e *cpu_cos_map_data) memGetSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSetUint8(&e.cpu_cos_queue, b, i+5, i, isSet)
	i = m.MemGetSet1(&e.rqe_queue_number, b, i, isSet)
	i = m.MemGetSet1(&e.truncate_copy_to_144_bytes, b, i, isSet)
	i = m.MemGetSetUint8(&e.strength, b, i+6, i, isSet)
	return i
}

type cpu_cos_map_entry struct {
	key, mask cpu_cos_map_key
	data      cpu_cos_map_data
	valid     bool
}

func (e *cpu_cos_map_entry) MemBits() int { return 296 }
func (e *cpu_cos_map_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	var key, mask cpu_cos_map_key
	if isSet {
		key, mask = e.key.tcamEncode(&e.mask, isSet)
	}
	i = key.memGetSet(b, i, isSet)
	i = mask.memGetSet(b, i, isSet)
	if !isSet {
		e.key, e.mask = key.tcamEncode(&mask, isSet)
	}
	i = e.data.memGetSet(b, i, isSet)
	if i != 296 {
		panic("296")
	}
}

type cpu_cos_map_mem m.MemElt

func (r *cpu_cos_map_mem) geta(q *DmaRequest, v *cpu_cos_map_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *cpu_cos_map_mem) seta(q *DmaRequest, v *cpu_cos_map_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *cpu_cos_map_mem) get(q *DmaRequest, v *cpu_cos_map_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *cpu_cos_map_mem) set(q *DmaRequest, v *cpu_cos_map_entry) {
	r.seta(q, v, sbus.Duplicate)
}
