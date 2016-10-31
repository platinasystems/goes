// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

// RX debug counters
type rx_debug_counter_type uint8

const (
	rx_debug_ip4_l3_drops                              rx_debug_counter_type = 0
	rx_debug_ip4_l3_packets                            rx_debug_counter_type = 1
	rx_debug_ip4_header_errors                         rx_debug_counter_type = 2
	rx_debug_ip4_routed_multicast_packets              rx_debug_counter_type = 3
	rx_debug_ip6_l3_drops                              rx_debug_counter_type = 4
	rx_debug_ip6_l3_packets                            rx_debug_counter_type = 5
	rx_debug_ip6_header_errors                         rx_debug_counter_type = 6
	rx_debug_ip6_routed_multicast_packets              rx_debug_counter_type = 7
	rx_debug_ibp_discard_cbp_full_drops                rx_debug_counter_type = 8
	rx_debug_good_unicast_packets                      rx_debug_counter_type = 9
	rx_debug_spanning_tree_state_not_forwarding_drops  rx_debug_counter_type = 10
	rx_debug_policy_drop                               rx_debug_counter_type = 11
	rx_debug_multicast_l2_packets                      rx_debug_counter_type = 12
	rx_debug_fp_drops                                  rx_debug_counter_type = 13
	rx_debug_multicast_l2_l3_drops                     rx_debug_counter_type = 14
	rx_debug_zero_port_bitmap_drops                    rx_debug_counter_type = 15
	rx_debug_hg_ipc_pause                              rx_debug_counter_type = 16
	rx_debug_denial_of_service_l3_header_error_packets rx_debug_counter_type = 19
	rx_debug_denial_of_service_l4_header_error_packets rx_debug_counter_type = 20
	rx_debug_denial_of_service_icmp_error_packets      rx_debug_counter_type = 21
	rx_debug_denial_of_service_fragment_error_packets  rx_debug_counter_type = 22
	rx_debug_l3_mtu_exceeded_punts                     rx_debug_counter_type = 23
	rx_debug_tunnel_packets                            rx_debug_counter_type = 24
	rx_debug_tunnel_error_packets                      rx_debug_counter_type = 25
	rx_debug_vlan_drops                                rx_debug_counter_type = 26
	rx_debug_hi_gig_header_error                       rx_debug_counter_type = 27
	rx_debug_multicast_index_error_drops               rx_debug_counter_type = 28
	rx_debug_lag_failover_loopback_packets             rx_debug_counter_type = 29
	rx_debug_lag_backup_port_down_drops                rx_debug_counter_type = 30
	rx_debug_parity_error_drops                        rx_debug_counter_type = 31
	rx_debug_hg_uc_lookup_cases                        rx_debug_counter_type = 32
	rx_debug_hg_mc_lookup_cases                        rx_debug_counter_type = 33
	rx_debug_urfp_drop                                 rx_debug_counter_type = 34
	rx_debug_vfp_drop                                  rx_debug_counter_type = 35
	rx_debug_dst_discard_drop                          rx_debug_counter_type = 36
	rx_debug_class_based_move_drop                     rx_debug_counter_type = 37
	rx_debug_mac_limit_exceeded_no_drop                rx_debug_counter_type = 38
	rx_debug_src_dst_mac_equal_drop                    rx_debug_counter_type = 39
	rx_debug_mac_limit_exceeded_drops                  rx_debug_counter_type = 40
	rx_debug_fcoe_class_2_drops                        rx_debug_counter_type = 56
	rx_debug_fcoe_class_2_frames                       rx_debug_counter_type = 57
	rx_debug_fcoe_class_3_drops                        rx_debug_counter_type = 58
	rx_debug_fcoe_class_3_frames                       rx_debug_counter_type = 59
)

type rx_counter_type uint8

const (
	rx_ip4_l3_drops rx_counter_type = iota
	rx_ip4_l3_packets
	rx_ip4_header_errors
	rx_ip4_routed_multicast_packets
	rx_ip6_l3_drops
	rx_ip6_l3_packets
	rx_ip6_header_errors
	rx_ip6_routed_multicast_packets
	rx_ibp_discard_cbp_full_drops
	rx_unicast_packets
	rx_spanning_tree_state_not_forwarding_drops
	rx_debug_0
	rx_debug_1
	rx_debug_2
	rx_debug_3
	rx_debug_4
	rx_debug_5
	rx_hi_gig_unknown_hgi_packets
	rx_hi_gig_control_packets
	rx_hi_gig_broadcast_packets
	rx_hi_gig_l2_multicast_packets
	rx_hi_gig_l3_multicast_packets
	rx_hi_gig_unknown_opcode_packets
	rx_debug_6
	rx_debug_7
	rx_debug_8
	rx_trill_packets
	rx_trill_trill_drops
	rx_trill_non_trill_drops
	rx_niv_frame_error_drops
	rx_niv_forwarding_error_drops
	rx_niv_frame_vlan_tagged
	rx_ecn_counter
	n_rx_counters
)

var rx_counter_prefix = "rx pipe "

var rx_counter_names = [...]string{
	rx_ip4_l3_drops:                             "ip4 l3 drops",
	rx_ip4_l3_packets:                           "ip4 l3 packets",
	rx_ip4_header_errors:                        "ip4 header errors",
	rx_ip4_routed_multicast_packets:             "ip4 routed multicast packets",
	rx_ip6_l3_drops:                             "ip6 l3 drops",
	rx_ip6_l3_packets:                           "ip6 l3 packets",
	rx_ip6_header_errors:                        "ip6 header errors",
	rx_ip6_routed_multicast_packets:             "ip6 routed multicast packets",
	rx_ibp_discard_cbp_full_drops:               "ibp discard cbp full drops",
	rx_unicast_packets:                          "unicast packets",
	rx_spanning_tree_state_not_forwarding_drops: "spanning tree state not forwarding drops",
	rx_debug_0:                       "debug 0",
	rx_debug_1:                       "debug 1",
	rx_debug_2:                       "debug 2",
	rx_debug_3:                       "debug 3",
	rx_debug_4:                       "debug 4",
	rx_debug_5:                       "debug 5",
	rx_hi_gig_unknown_hgi_packets:    "hi gig unknown hgi packets",
	rx_hi_gig_control_packets:        "hi gig control packets",
	rx_hi_gig_broadcast_packets:      "hi gig broadcast packets",
	rx_hi_gig_l2_multicast_packets:   "hi gig l2 multicast packets",
	rx_hi_gig_l3_multicast_packets:   "hi gig l3 multicast packets",
	rx_hi_gig_unknown_opcode_packets: "hi gig unknown opcode packets",
	rx_debug_6:                       "debug 6",
	rx_debug_7:                       "debug 7",
	rx_debug_8:                       "debug 8",
	rx_trill_packets:                 "trill packets",
	rx_trill_trill_drops:             "trill trill drops",
	rx_trill_non_trill_drops:         "trill non trill drops",
	rx_niv_frame_error_drops:         "niv frame error drops",
	rx_niv_forwarding_error_drops:    "niv forwarding error drops",
	rx_niv_frame_vlan_tagged:         "vlan tagged packets",
	rx_ecn_counter:                   "ecn counter",
}
