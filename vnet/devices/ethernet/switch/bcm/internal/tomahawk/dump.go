// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build !snake

package tomahawk

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/packet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/sbus"
)

var SnakeLoopMode loopbackType = loopbackNone

// Set all ports to forwarding state for spanning tree group 0.
func (t *tomahawk) vlan_init() {
	q := t.getDmaReq()

	const stg = 1

	// Group 0 gets all ports in forwarding state.
	{
		var e vlan_spanning_tree_group_entry
		for i := range e {
			e[i] = m.SpanningTreeForwarding
		}
		t.rx_pipe_mems.vlan_spanning_tree_group[stg].set(q, &e)
		t.tx_pipe_mems.vlan_spanning_tree_group[stg].set(q, &e)
	}
	q.Do()

	// Disable l2 learning.
	t.rx_pipe_mems.vlan_profile[0].Set(&q.DmaRequest, BlockRxPipe, sbus.Duplicate, 1<<2)
	q.Do()

	// Profile 0: Arrange for untagged packets to have internal vlan added (from port table).
	// This makes packet tagged which we don't want; so disabled.
	if false {
		var e rx_vlan_tag_action_entry
		e[ut_otag_action] = vlan_tag_action_add
		t.rx_pipe_mems.vlan_tag_action_profile[0].set(q, &e)
		q.Do()
	}

	for vlan := 0; vlan < 2; vlan++ {
		var e rx_vlan_entry
		e.valid = true
		e.spanning_tree_group = stg
		e.members[rx] = t.all_ports
		e.members[tx] = e.members[rx]
		// e.flex_counter_ref.alloc(t, 1, 0xf, BlockRxPipe)

		var f tx_vlan_entry
		f.valid = true
		f.spanning_tree_group = stg
		f.members = e.members[rx]
		f.untagged_members = e.members[rx]
		// f.flex_counter_ref.alloc(t, 1, 0xf, BlockEpipe)

		t.tx_pipe_mems.vlan[vlan].set(q, &f)
		t.rx_pipe_mems.vlan[vlan].set(q, &e)
	}
	q.Do()

	{
		var e vlan_range_entry
		e[0].min = 0
		e[0].max = 4095
		t.rx_pipe_mems.vlan_range[0].set(q, &e)
		q.Do()
	}

	{
		e := vlan_protocol_data_entry{
			outer_vlan: 1,
		}
		t.rx_pipe_mems.vlan_protocol_data[0][0].set(q, &e)
		t.rx_pipe_mems.vlan_protocol[0].Set(&q.DmaRequest, BlockRxPipe, sbus.Duplicate, 0x8100|1<<18|1<<19|1<<20)
		q.Do()
	}
}

func (t *tomahawk) l2_init() {
	q := t.getDmaReq()

	if false {
		for i := 0; i < 16<<10; i++ {
			e := &l2_multicast_entry{}
			e.ports.or(&t.all_ports)
			e.valid = true
			t.rx_pipe_mems.l2_multicast[i].set(q, e)
			q.Do()
		}
	}
}

var next_hop_dst_ethernet_address = m.EthernetAddress{0, 1, 2, 3, 4, 5}

func (t *tomahawk) l3_init() {
	q := t.getDmaReq()

	if false {

		const (
			next_hop_index      = 1 // must be non-zero; zero => invalid
			next_hop_index_alpm = 2
			l3_intf_index       = 1 // must be non-zero; zero => invalid
		)

		// l3 defip: 0/0 points to next hop 1
		{
			e := l3_defip_entry{}
			e[0].key = l3_defip_tcam_key{
				key_type:   l3_defip_ip4,
				Ip4Address: m.Ip4Address{0, 0, 0, 0},
			}
			e[0].mask = l3_defip_tcam_key{
				key_type:   0xff,
				Ip4Address: m.Ip4Address{0, 0, 0, 0},
			}
			e[0].next_hop = m.NextHop{Index: next_hop_index}
			e[0].is_valid = true
			e[0].bucket_index = 0
			e[0].sub_bucket_index = 1
			t.rx_pipe_mems.l3_defip[0].set(q, &e)
			q.Do()
		}

		// l3 defip alpm
		if false {
			e := l3_defip_alpm_ip4_entry{}
			e.is_valid = true
			e.next_hop = m.NextHop{Index: next_hop_index_alpm}
			e.dst_length = 32
			e.dst = m.Ip4Address{0x5, 0x6, 0x7, 0x8}
			e.sub_bucket_index = 1
			t.rx_pipe_mems.l3_defip_alpm_ipv4[0][0][0].set(q, &e)
			q.Do()
		}

		// rx next hop
		{
			e := rx_next_hop_entry{}
			e.rx_next_hop_type = rx_next_hop_type_tunnel
			e.index = l3_intf_index
			if false {
				e.drop = true
				e.copy_to_cpu = true
			} else {
				e.LogicalPort.Set(uint(phys_port_cpu.toPipe()))
			}
			t.rx_pipe_mems.l3_next_hop[next_hop_index].set(q, &e)
			t.rx_pipe_mems.l3_next_hop[next_hop_index_alpm].set(q, &e)
			// large mtu => no drops
			t.rx_pipe_mems.l3_interface_mtu[m.Unicast][l3_intf_index].Set(&q.DmaRequest, BlockRxPipe, sbus.Duplicate, 0x3fff)
			q.Do()
		}

		// tx next hop
		{
			e := [2]l3_unicast_tx_next_hop{}
			e[0].dst_ethernet_address = next_hop_dst_ethernet_address
			e[0].disable_dst_ethernet_address_rewrite = true
			e[0].disable_src_ethernet_address_rewrite = true
			e[0].disable_ip_ttl_decrement = true
			e[0].l3_intf_index = l3_intf_index
			// e[0].flex_counter_ref.alloc(t, 0, 0xf, BlockEpipe)
			t.tx_pipe_mems.l3_next_hop[next_hop_index].set(q, &e[0])

			e[1] = e[0]
			// e[1].flex_counter_ref.alloc(t, 0, 0xf, BlockEpipe)
			t.tx_pipe_mems.l3_next_hop[next_hop_index_alpm].set(q, &e[1])

			q.Do()
		}

		// Add l2 user entry to drop re-written packet when/if it ever comes back.
		if true {
			e := l2_user_entry{
				valid: true,
				key: l2_user_entry_key{
					typ:             l2_user_entry_type_vlan_mac,
					vlan_vfi:        1,
					EthernetAddress: next_hop_dst_ethernet_address,
				},
				mask: l2_user_entry_key{
					typ:             1,
					vlan_vfi:        0xfff,
					EthernetAddress: m.EthernetAddress{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				data: l2_user_entry_data{
					drop: true,
					cpu:  false,
				},
			}
			{
				const (
					pipe physical_port_number = 0
					port physical_port_number = 0
				)
				pipePort := uint((phys_port_data_lo + (8*pipe+port)*4).toPipe())
				e.data.LogicalPort.Set(pipePort)
			}

			t.rx_pipe_mems.l2_user_entry[0].set(q, &e)
			q.Do()
		}
	}
}

func (t *tomahawk) cpu_rx_init() {
	q := t.getDmaReq()

	// Specifiy exception packets to be sent to cpu.
	{
		const (
			ip_unicast_ttl_equals_1   = 1 << 27
			ip_multicast_ttl_equals_1 = 1 << 27
			unknown_multicast         = 1 << 3
		)
		// For IP packets with ttl = 1 either you punt to cpu here or packet gets dropped.
		t.rx_pipe_regs.cpu_control_1.set(q, ip_unicast_ttl_equals_1|ip_multicast_ttl_equals_1|unknown_multicast)
		q.Do()
	}

	// Enable arp request/reply to be sent to cpu.
	t.rx_pipe_regs.protocol_pkt_control[0].set(q, 1<<4|1<<6)
	q.Do()

	{
		e := [3]cpu_cos_map_entry{}

		// Re-direct arp packets to cos queue 1.
		e[0].key.setReason(cpu_rx_reason_arp, true)
		e[0].mask.setReason(cpu_rx_reason_arp, true)
		e[0].data.cpu_cos_queue = uint8(packet.Rx_next_punt)
		e[0].valid = true

		// Re-direct ip4 dst miss to cos queue 2.
		e[1].key.setReason(cpu_rx_reason_l3_dst_miss, true)
		e[1].mask.setReason(cpu_rx_reason_l3_dst_miss, true)
		e[1].data.cpu_cos_queue = uint8(packet.Rx_next_punt)
		e[1].valid = true

		// Punt everything else.
		e[2].data.cpu_cos_queue = uint8(packet.Rx_next_punt)
		e[2].valid = true

		for i := range e {
			t.rx_pipe_mems.cpu_cos_map[i].set(q, &e[i])
		}
		q.Do()
	}
}

func (t *tomahawk) garbage_dump_init() {
	t.cpu_rx_init()
	t.vlan_init()
	t.l2_init()
	t.l3_init()
}
