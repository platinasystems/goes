// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

// TX debug counters
type tx_debug_counter_type uint8

const (
	tx_debug_ip4_unicast_packets                      tx_debug_counter_type = 0
	tx_debug_ip4_unicast_aged_and_drop_packets        tx_debug_counter_type = 1
	tx_debug_ip4_multicast_packets                    tx_debug_counter_type = 2
	tx_debug_ip4_multicast_aged_and_drop_packets      tx_debug_counter_type = 3
	tx_debug_ip6_unicast_packets                      tx_debug_counter_type = 4
	tx_debug_ip6_unicast_aged_and_drop_packets        tx_debug_counter_type = 5
	tx_debug_ip6_multicast_packets                    tx_debug_counter_type = 6
	tx_debug_ip6_multicast_aged_and_drop_packets      tx_debug_counter_type = 7
	tx_debug_tunnel_packets                           tx_debug_counter_type = 8
	tx_debug_tunnel_error_packets                     tx_debug_counter_type = 9
	tx_debug_ttl_threshold_reached_drops              tx_debug_counter_type = 10
	tx_debug_untagged_or_l3_cfi_set_drops             tx_debug_counter_type = 11
	tx_debug_vlan_tagged_packets                      tx_debug_counter_type = 12
	tx_debug_invalid_vlan_drops                       tx_debug_counter_type = 13
	tx_debug_vxlt_table_miss_drops                    tx_debug_counter_type = 14
	tx_debug_spanning_tree_state_not_forwarding_drops tx_debug_counter_type = 15
	tx_debug_packet_aged_drops                        tx_debug_counter_type = 16
	tx_debug_l2_multicast_drops                       tx_debug_counter_type = 17
	tx_debug_packets_dropped                          tx_debug_counter_type = 18
	tx_debug_mirror_flag                              tx_debug_counter_type = 19
	tx_debug_sip_link_local_drops                     tx_debug_counter_type = 20
	tx_debug_hg_l3_unicast_packets                    tx_debug_counter_type = 21
	tx_debug_hg_l3_multicast_packets                  tx_debug_counter_type = 22
	tx_debug_hg2_unknown_drops                        tx_debug_counter_type = 23
	tx_debug_hg_unknown_drops                         tx_debug_counter_type = 24
	tx_debug_l2_mtu_fail_drops                        tx_debug_counter_type = 25
	tx_debug_parity_error_drops                       tx_debug_counter_type = 26
	tx_debug_ip_length_check_drops                    tx_debug_counter_type = 27
	tx_debug_module_id_gt_31_drops                    tx_debug_counter_type = 28
	tx_debug_byte_additions_too_large_drops           tx_debug_counter_type = 29
	tx_debug_fcoe_class_2_tx_frames                   tx_debug_counter_type = 31
	tx_debug_fcoe_class_3_tx_frames                   tx_debug_counter_type = 32
	tx_debug_protection_data_drops                    tx_debug_counter_type = 37
)

type tx_counter_type uint8

const (
	tx_debug_0 tx_counter_type = iota
	tx_debug_1
	tx_debug_2
	tx_debug_3
	tx_debug_4
	tx_debug_5
	tx_debug_6
	tx_debug_7
	tx_debug_8
	tx_debug_9
	tx_debug_a
	tx_debug_b
	tx_trill_packets
	tx_trill_access_port_drops
	tx_trill_non_trill_drops
	tx_ecn_errors
	tx_purge_cell_error_drops
	n_tx_counters
)

var tx_counter_prefix = "tx pipe "

var tx_counter_names = [...]string{
	tx_debug_0:                 "debug 0",
	tx_debug_1:                 "debug 1",
	tx_debug_2:                 "debug 2",
	tx_debug_3:                 "debug 3",
	tx_debug_4:                 "debug 4",
	tx_debug_5:                 "debug 5",
	tx_debug_6:                 "debug 6",
	tx_debug_7:                 "debug 7",
	tx_debug_8:                 "debug 8",
	tx_debug_9:                 "debug 9",
	tx_debug_a:                 "debug a",
	tx_debug_b:                 "debug b",
	tx_trill_packets:           "trill packets",
	tx_trill_access_port_drops: "trill access port drops",
	tx_trill_non_trill_drops:   "trill non trill drops",
	tx_ecn_errors:              "ecn errors",
	tx_purge_cell_error_drops:  "purge cell error drops",
}
