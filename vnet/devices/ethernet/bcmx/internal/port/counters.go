package port

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

// Port counters
type Counter_type int

const (
	Rx_64_byte_packets Counter_type = iota
	Rx_65_to_127_byte_packets
	Rx_128_to_255_byte_packets
	Rx_256_to_511_byte_packets
	Rx_512_to_1023_byte_packets
	Rx_1024_to_1518_byte_packets
	Rx_1519_to_1522_byte_vlan_packets
	Rx_1519_to_2047_byte_packets
	Rx_2048_to_4096_byte_packets
	Rx_4096_to_9216_byte_packets
	Rx_9217_to_16383_byte_packets
	Rx_packets
	Rx_unicast_packets
	Rx_multicast_packets
	Rx_broadcast_packets
	Rx_crc_error_packets
	Rx_control_packets
	Rx_flow_control_packets
	Rx_pfc_packets
	Rx_unsupported_opcode_control_packets
	Rx_unsupported_dst_address_control_packets
	Rx_src_address_not_unicast_packets
	Rx_alignment_error_packets
	Rx_802_3_length_error_packets
	Rx_code_error_packets
	Rx_false_carrier_events
	Rx_oversize_packets
	Rx_jabber_packets
	Rx_mtu_check_error_packets
	Rx_mac_sec_crc_matched_packets
	Rx_promiscuous_packets
	Rx_1tag_vlan_packets
	Rx_2tag_vlan_packets
	Rx_truncated_packets
	Rx_good_packets
	Rx_xon_to_xoff_priority_0
	Rx_xon_to_xoff_priority_1
	Rx_xon_to_xoff_priority_2
	Rx_xon_to_xoff_priority_3
	Rx_xon_to_xoff_priority_4
	Rx_xon_to_xoff_priority_5
	Rx_xon_to_xoff_priority_6
	Rx_xon_to_xoff_priority_7
	Rx_pfc_priority_0
	Rx_pfc_priority_1
	Rx_pfc_priority_2
	Rx_pfc_priority_3
	Rx_pfc_priority_4
	Rx_pfc_priority_5
	Rx_pfc_priority_6
	Rx_pfc_priority_7
	unused_0x33
	Rx_undersize_packets
	Rx_fragment_packets
	Rx_eee_lpi_events
	Rx_eee_lpi_duration
	unused_0x38
	unused_0x39
	unused_0x3a
	unused_0x3b
	unused_0x3c
	Rx_bytes
	Rx_runt_bytes
	Rx_runt_packets

	// Tx counters start at 0x40
	Tx_64_byte_packets
	Tx_65_to_127_byte_packets
	Tx_128_to_255_byte_packets
	Tx_256_to_511_byte_packets
	Tx_512_to_1023_byte_packets
	Tx_1024_to_1518_byte_packets
	Tx_1519_to_1522_byte_vlan_packets
	Tx_1519_to_2047_byte_packets
	Tx_2048_to_4096_byte_packets
	Tx_4096_to_9216_byte_packets
	Tx_9217_to_16383_byte_packets
	Tx_good_packets
	Tx_packets
	Tx_unicast_packets
	Tx_multicast_packets
	Tx_broadcast_packets
	Tx_flow_control_packets
	Tx_pfc_packets
	Tx_jabber_packets
	Tx_fcs_errors
	Tx_control_packets
	Tx_oversize
	Tx_single_deferral_packets
	Tx_multiple_deferral_packets
	Tx_single_collision_packets
	Tx_multiple_collision_packets
	Tx_late_collision_packets
	Tx_excessive_collision_packets
	Tx_fragments
	Tx_system_error_packets
	Tx_1tag_vlan_packets
	Tx_2tag_vlan_packets
	Tx_runt_packets
	Tx_fifo_underrun_packets
	Tx_pfc_priority0_packets
	Tx_pfc_priority1_packets
	Tx_pfc_priority2_packets
	Tx_pfc_priority3_packets
	Tx_pfc_priority4_packets
	Tx_pfc_priority5_packets
	Tx_pfc_priority6_packets
	Tx_pfc_priority7_packets
	Tx_eee_lpi_events
	Tx_eee_lpi_duration
	unused_0x6c
	unused_0x6d
	Tx_total_collisions
	Tx_bytes
	N_counters
)

var CounterNames = [...]string{
	Rx_64_byte_packets:                         "rx 64 byte packets",
	Rx_65_to_127_byte_packets:                  "rx 65 to 127 byte packets",
	Rx_128_to_255_byte_packets:                 "rx 128 to 255 byte packets",
	Rx_256_to_511_byte_packets:                 "rx 256 to 511 byte packets",
	Rx_512_to_1023_byte_packets:                "rx 512 to 1023 byte packets",
	Rx_1024_to_1518_byte_packets:               "rx 1024 to 1518 byte packets",
	Rx_1519_to_1522_byte_vlan_packets:          "rx 1519 to 1522 byte vlan packets",
	Rx_1519_to_2047_byte_packets:               "rx 1519 to 2047 byte packets",
	Rx_2048_to_4096_byte_packets:               "rx 2048 to 4096 byte packets",
	Rx_4096_to_9216_byte_packets:               "rx 4096 to 9216 byte packets",
	Rx_9217_to_16383_byte_packets:              "rx 9217 to 16383 byte packets",
	Rx_packets:                                 "rx packets",
	Rx_unicast_packets:                         "rx unicast packets",
	Rx_multicast_packets:                       "rx multicast packets",
	Rx_broadcast_packets:                       "rx broadcast packets",
	Rx_crc_error_packets:                       "rx crc error packets",
	Rx_control_packets:                         "rx control packets",
	Rx_flow_control_packets:                    "rx flow control packets",
	Rx_pfc_packets:                             "rx pfc packets",
	Rx_unsupported_opcode_control_packets:      "rx unsupported opcode control packets",
	Rx_unsupported_dst_address_control_packets: "rx unsupported dst address control packets",
	Rx_src_address_not_unicast_packets:         "rx src address not_unicast packets",
	Rx_alignment_error_packets:                 "rx alignment error packets",
	Rx_802_3_length_error_packets:              "rx 802.3 length_error packets",
	Rx_code_error_packets:                      "rx code error packets",
	Rx_false_carrier_events:                    "rx false carrier events",
	Rx_oversize_packets:                        "rx oversize packets",
	Rx_jabber_packets:                          "rx jabber packets",
	Rx_mtu_check_error_packets:                 "rx mtu check error packets",
	Rx_mac_sec_crc_matched_packets:             "rx mac sec crc_matched packets",
	Rx_promiscuous_packets:                     "rx promiscuous packets",
	Rx_1tag_vlan_packets:                       "rx 1tag vlan packets",
	Rx_2tag_vlan_packets:                       "rx 2tag vlan packets",
	Rx_truncated_packets:                       "rx truncated packets",
	Rx_good_packets:                            "rx good packets",
	Rx_xon_to_xoff_priority_0:                  "rx xon to xoff priority 0",
	Rx_xon_to_xoff_priority_1:                  "rx xon to xoff priority 1",
	Rx_xon_to_xoff_priority_2:                  "rx xon to xoff priority 2",
	Rx_xon_to_xoff_priority_3:                  "rx xon to xoff priority 3",
	Rx_xon_to_xoff_priority_4:                  "rx xon to xoff priority 4",
	Rx_xon_to_xoff_priority_5:                  "rx xon to xoff priority 5",
	Rx_xon_to_xoff_priority_6:                  "rx xon to xoff priority 6",
	Rx_xon_to_xoff_priority_7:                  "rx xon to xoff priority 7",
	Rx_pfc_priority_0:                          "rx pfc priority 0",
	Rx_pfc_priority_1:                          "rx pfc priority 1",
	Rx_pfc_priority_2:                          "rx pfc priority 2",
	Rx_pfc_priority_3:                          "rx pfc priority 3",
	Rx_pfc_priority_4:                          "rx pfc priority 4",
	Rx_pfc_priority_5:                          "rx pfc priority 5",
	Rx_pfc_priority_6:                          "rx pfc priority 6",
	Rx_pfc_priority_7:                          "rx pfc priority 7",
	Rx_undersize_packets:                       "rx undersize packets",
	Rx_fragment_packets:                        "rx fragment packets",
	Rx_eee_lpi_events:                          "rx eee lpi events",
	Rx_eee_lpi_duration:                        "rx eee lpi duration",
	Rx_bytes:                                   "rx bytes",
	Rx_runt_bytes:                              "rx runt bytes",
	Rx_runt_packets:                            "rx runt packets",
	Tx_64_byte_packets:                         "tx 64 byte packets",
	Tx_65_to_127_byte_packets:                  "tx 65 to 127 byte packets",
	Tx_128_to_255_byte_packets:                 "tx 128 to 255 byte packets",
	Tx_256_to_511_byte_packets:                 "tx 256 to 511 byte packets",
	Tx_512_to_1023_byte_packets:                "tx 512 to 1023 byte packets",
	Tx_1024_to_1518_byte_packets:               "tx 1024 to 1518 byte packets",
	Tx_1519_to_1522_byte_vlan_packets:          "tx 1519 to 1522 byte_vlan packets",
	Tx_1519_to_2047_byte_packets:               "tx 1519 to 2047 byte packets",
	Tx_2048_to_4096_byte_packets:               "tx 2048 to 4096 byte packets",
	Tx_4096_to_9216_byte_packets:               "tx 4096 to 9216 byte packets",
	Tx_9217_to_16383_byte_packets:              "tx 9217 to 16383 byte packets",
	Tx_good_packets:                            "tx good packets",
	Tx_packets:                                 "tx packets",
	Tx_unicast_packets:                         "tx unicast packets",
	Tx_multicast_packets:                       "tx multicast packets",
	Tx_broadcast_packets:                       "tx broadcast packets",
	Tx_flow_control_packets:                    "tx flow_control packets",
	Tx_pfc_packets:                             "tx pfc packets",
	Tx_jabber_packets:                          "tx jabber packets",
	Tx_fcs_errors:                              "tx fcs errors",
	Tx_control_packets:                         "tx control packets",
	Tx_oversize:                                "tx oversize",
	Tx_single_deferral_packets:                 "tx single deferral packets",
	Tx_multiple_deferral_packets:               "tx multiple deferral packets",
	Tx_single_collision_packets:                "tx single collision packets",
	Tx_multiple_collision_packets:              "tx multiple collision packets",
	Tx_late_collision_packets:                  "tx late collision packets",
	Tx_excessive_collision_packets:             "tx excessive collision packets",
	Tx_fragments:                               "tx fragments",
	Tx_system_error_packets:                    "tx system error packets",
	Tx_1tag_vlan_packets:                       "tx 1tag vlan packets",
	Tx_2tag_vlan_packets:                       "tx 2tag vlan packets",
	Tx_runt_packets:                            "tx runt packets",
	Tx_fifo_underrun_packets:                   "tx fifo underrun packets",
	Tx_pfc_priority0_packets:                   "tx pfc priority 0 packets",
	Tx_pfc_priority1_packets:                   "tx pfc priority 1 packets",
	Tx_pfc_priority2_packets:                   "tx pfc priority 2 packets",
	Tx_pfc_priority3_packets:                   "tx pfc priority 3 packets",
	Tx_pfc_priority4_packets:                   "tx pfc priority 4 packets",
	Tx_pfc_priority5_packets:                   "tx pfc priority 5 packets",
	Tx_pfc_priority6_packets:                   "tx pfc priority 6 packets",
	Tx_pfc_priority7_packets:                   "tx pfc priority 7 packets",
	Tx_eee_lpi_events:                          "tx eee lpi events",
	Tx_eee_lpi_duration:                        "tx eee lpi duration",
	Tx_total_collisions:                        "tx total collisions",
	Tx_bytes:                                   "tx bytes",
}

// Ordering of counters for display.
var Counter_order = [vnet.NRxTx][]Counter_type{
	vnet.Rx: []Counter_type{
		Rx_packets,
		Rx_bytes,
		Rx_64_byte_packets,
		Rx_65_to_127_byte_packets,
		Rx_128_to_255_byte_packets,
		Rx_256_to_511_byte_packets,
		Rx_512_to_1023_byte_packets,
		Rx_1024_to_1518_byte_packets,
		Rx_1519_to_1522_byte_vlan_packets,
		Rx_1519_to_2047_byte_packets,
		Rx_2048_to_4096_byte_packets,
		Rx_4096_to_9216_byte_packets,
		Rx_9217_to_16383_byte_packets,
		Rx_good_packets,
		Rx_unicast_packets,
		Rx_multicast_packets,
		Rx_broadcast_packets,
		Rx_crc_error_packets,
		Rx_control_packets,
		Rx_flow_control_packets,
		Rx_pfc_packets,
		Rx_unsupported_opcode_control_packets,
		Rx_unsupported_dst_address_control_packets,
		Rx_src_address_not_unicast_packets,
		Rx_alignment_error_packets,
		Rx_802_3_length_error_packets,
		Rx_code_error_packets,
		Rx_false_carrier_events,
		Rx_oversize_packets,
		Rx_jabber_packets,
		Rx_mtu_check_error_packets,
		Rx_mac_sec_crc_matched_packets,
		Rx_promiscuous_packets,
		Rx_1tag_vlan_packets,
		Rx_2tag_vlan_packets,
		Rx_truncated_packets,
		Rx_xon_to_xoff_priority_0,
		Rx_xon_to_xoff_priority_1,
		Rx_xon_to_xoff_priority_2,
		Rx_xon_to_xoff_priority_3,
		Rx_xon_to_xoff_priority_4,
		Rx_xon_to_xoff_priority_5,
		Rx_xon_to_xoff_priority_6,
		Rx_xon_to_xoff_priority_7,
		Rx_pfc_priority_0,
		Rx_pfc_priority_1,
		Rx_pfc_priority_2,
		Rx_pfc_priority_3,
		Rx_pfc_priority_4,
		Rx_pfc_priority_5,
		Rx_pfc_priority_6,
		Rx_pfc_priority_7,
		Rx_undersize_packets,
		Rx_fragment_packets,
		Rx_eee_lpi_events,
		Rx_eee_lpi_duration,
		Rx_runt_bytes,
		Rx_runt_packets,
	},
	vnet.Tx: []Counter_type{
		Tx_packets,
		Tx_bytes,
		Tx_64_byte_packets,
		Tx_65_to_127_byte_packets,
		Tx_128_to_255_byte_packets,
		Tx_256_to_511_byte_packets,
		Tx_512_to_1023_byte_packets,
		Tx_1024_to_1518_byte_packets,
		Tx_1519_to_1522_byte_vlan_packets,
		Tx_1519_to_2047_byte_packets,
		Tx_2048_to_4096_byte_packets,
		Tx_4096_to_9216_byte_packets,
		Tx_9217_to_16383_byte_packets,
		Tx_good_packets,
		Tx_unicast_packets,
		Tx_multicast_packets,
		Tx_broadcast_packets,
		Tx_flow_control_packets,
		Tx_pfc_packets,
		Tx_jabber_packets,
		Tx_fcs_errors,
		Tx_control_packets,
		Tx_oversize,
		Tx_single_deferral_packets,
		Tx_multiple_deferral_packets,
		Tx_single_collision_packets,
		Tx_multiple_collision_packets,
		Tx_late_collision_packets,
		Tx_excessive_collision_packets,
		Tx_fragments,
		Tx_system_error_packets,
		Tx_1tag_vlan_packets,
		Tx_2tag_vlan_packets,
		Tx_runt_packets,
		Tx_fifo_underrun_packets,
		Tx_pfc_priority0_packets,
		Tx_pfc_priority1_packets,
		Tx_pfc_priority2_packets,
		Tx_pfc_priority3_packets,
		Tx_pfc_priority4_packets,
		Tx_pfc_priority5_packets,
		Tx_pfc_priority6_packets,
		Tx_pfc_priority7_packets,
		Tx_eee_lpi_events,
		Tx_eee_lpi_duration,
		Tx_total_collisions,
	},
}

// Dma port counters and add to running count.
func (p *PortBlock) GetCounters(port m.Porter, v *[N_counters]uint64) {
	r, _, _, _ := p.get_regs()
	q := p.dmaReq()
	pi := m.GetSubPortIndex(port)

	var buf [2 * N_counters]uint32
	q.Add(&sbus.DmaCmd{
		Command: sbus.Command{
			Opcode: sbus.ReadRegister,
			Block:  p.SbusBlock,
		},
		Address: r.counters[0][pi].address(),
		Rx:      buf[:],
		Count:   uint(N_counters),
		Log2SbusAddressIncrement: m.Log2NRegPorts,
	})
	q.Do()
	for i := range v {
		v[i] = uint64(buf[2*i+0]) | uint64(buf[2*i+1])<<32
	}
}
