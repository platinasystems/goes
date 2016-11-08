// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build snake

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/packet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

var SnakeLoopMode loopbackType = loopbackPhy // change to loopbackExtCable for ext cables

// Set all ports to forwarding state for spanning tree group 0.
func (t *fe1a) vlan_init() {
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

	for vlan := 1; vlan < 2; vlan++ {
		var e rx_vlan_entry
		e.valid = true
		e.spanning_tree_group = stg
		e.members[rx] = t.all_ports
		e.members[tx] = e.members[rx]
		// e.flex_counter_ref.alloc(t, 1, 0xf, BlockIpipe, "vlan %d", vlan)

		var f tx_vlan_entry
		f.valid = true
		f.spanning_tree_group = stg
		f.members = e.members[rx]
		f.untagged_members = e.members[rx]
		// f.flex_counter_ref.alloc(t, 1, 0xf, BlockEpipe, "vlan %d", vlan)

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

func (t *fe1a) l2_init() {
	q := t.getDmaReq()

	{
		for i := snake_port(0); i < n_snake_port; i++ {
			e := source_trunk_map_entry{}
			l3_iif := i.l3_iif_for_port()
			e.index = uint16(l3_iif)
			e.lport_profile_index = uint8(i.to_pipe())
			t.rx_pipe_mems.source_trunk_map[i.to_pipe()].set(q, &e)
		}
		// Arrange for cpu port to inject packets just like any other port.
		{
			e := source_trunk_map_entry{}
			cpu := phys_port_cpu
			e.lport_profile_index = uint8(cpu.toPipe())
			// Inherit l3 iif and vrf from port 0.
			e.index = uint16(snake_port(0).l3_iif_for_port())
			t.rx_pipe_mems.source_trunk_map[cpu.toPipe()].set(q, &e)
		}
		q.Do()
	}
}

var next_hop_dst_ethernet_address = m.EthernetAddress{0, 1, 2, 3, 4, 5}

type snake_port uint

const n_snake_port = 32

func (i snake_port) to_phys() physical_port_number {
	return phys_port_data_lo + 4*physical_port_number(i)
}
func (i snake_port) to_pipe() pipe_port_number { return i.to_phys().toPipe() }
func (i snake_port) vrf_for_port() uint        { return uint(i) + 1 }
func (i snake_port) l3_iif_for_port() uint     { return uint(i) + 1 }
func (i snake_port) next_hop_for_port() uint   { return uint(i) + 1 }

func (t *fe1a) l3_init() {
	q := t.getDmaReq()

	{
		d := my_station_tcam_entry{}
		d.key.EthernetAddress = m.EthernetAddress{0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0x82}
		d.mask.EthernetAddress = m.EthernetAddress{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		d.data = my_station_tcam_data{
			ip4_unicast_enable:   true,
			ip6_unicast_enable:   true,
			ip4_multicast_enable: true,
			ip6_multicast_enable: true,
			mpls_enable:          true,
		}
		d.valid = true
		for i := snake_port(0); i < n_snake_port; i++ {
			e := d
			e.key.LogicalPort.Set(uint(i.to_pipe()))
			t.rx_pipe_mems.my_station_tcam[i].set(q, &e)
		}
		e := d
		e.key.LogicalPort.Set(uint(phys_port_cpu.toPipe()))
		t.rx_pipe_mems.my_station_tcam[n_snake_port].set(q, &e)
		q.Do()
	}

	const (
		next_hop_index      = 1
		next_hop_index_alpm = 2
		tx_l3_intf_index    = 1
	)

	{
		for i := snake_port(0); i < n_snake_port; i++ {
			e := l3_defip_entry{}
			e[0].key = l3_defip_tcam_key{
				key_type:   l3_defip_ip4,
				Vrf:        m.Vrf(i.vrf_for_port()),
				Ip4Address: m.Ip4Address{0, 0, 0, 0},
			}
			e[0].mask = l3_defip_tcam_key{
				key_type:   l3_defip_key_type(0xff),
				Vrf:        m.Vrf(^uint16(0)),
				Ip4Address: m.Ip4Address{0, 0, 0, 0},
			}
			next_port := (i + 1) % n_snake_port
			if SnakeLoopMode == loopbackExtCable {
				// external cable snake - flip indexes
				if next_port%2 == 0 {
					next_port += 1
				} else {
					next_port -= 1
				}
			}
			e[0].next_hop = m.NextHop{Index: uint16(next_port.next_hop_for_port())}
			e[0].is_valid = true
			t.rx_pipe_mems.l3_defip[i].set(q, &e)
		}
		q.Do()
	}

	// rx next hop
	{
		for i := snake_port(0); i < n_snake_port; i++ {
			e := rx_next_hop_entry{}
			e.rx_next_hop_type = rx_next_hop_type_tunnel
			l3_oif := i.l3_iif_for_port()
			e.index = uint16(l3_oif)

			if false {
				e.drop = true
				e.copy_to_cpu = true
			} else {
				e.LogicalPort.Set(uint(i.to_pipe()))
			}
			nh := i.next_hop_for_port()
			t.rx_pipe_mems.l3_next_hop[nh].set(q, &e)
			// large mtu => no drops
			t.rx_pipe_mems.l3_interface_mtu[m.Unicast][l3_oif].Set(&q.DmaRequest, BlockRxPipe, sbus.Duplicate, 0x3fff)
		}
		q.Do()
	}

	// tx next hop
	{
		for i := snake_port(0); i < n_snake_port; i++ {
			e := l3_unicast_tx_next_hop{}
			e.disable_dst_ethernet_address_rewrite = true
			e.disable_src_ethernet_address_rewrite = true
			e.disable_ip_ttl_decrement = true
			nh := i.next_hop_for_port()
			l3_oif := i.l3_iif_for_port()
			e.l3_intf_index = uint16(l3_oif)
			// e.flex_counter_ref.alloc(t, 0, 0xf, BlockEpipe, "l3 next_hop port %d", i)
			t.tx_pipe_mems.l3_next_hop[nh].set(q, &e)
		}
		q.Do()
	}

	// Add l2 user entry to drop re-written packet when/if it ever comes back.
	if false {
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

	// rx l3 iif profile
	{
		e := rx_l3_interface_profile_entry{
			ip4_enable:           true,
			ip6_enable:           true,
			ip4_multicast_enable: true,
			ip6_multicast_enable: true,
		}
		t.rx_pipe_mems.l3_interface_profile[0].set(q, &e)
		q.Do()
	}

	// rx l3 iif
	{
		for i := snake_port(0); i < n_snake_port; i++ {
			iif := rx_l3_interface_entry{}
			oif := tx_l3_interface_entry{}
			vrf := vrf_entry{}
			oif.outer_vlan.id = 1
			l3_iif := i.l3_iif_for_port()
			iif.vrf = uint16(i.vrf_for_port())
			// iif.flex_counter_ref.alloc(t, 2, 0xf, BlockIpipe, "l3 iif port %d", i)
			t.rx_pipe_mems.l3_interface[l3_iif].set(q, &iif)
			t.tx_pipe_mems.l3_interface[l3_iif].set(q, &oif)
			// vrf.flex_counter_ref.alloc(t, 3, 0xf, BlockIpipe, "l3 vrf port %d", i)
			t.rx_pipe_mems.vrf[i.vrf_for_port()].set(q, &vrf)
		}
		q.Do()
	}
}

func (t *fe1a) cpu_rx_init() {
	q := t.getDmaReq()

	// Enable ip4 (but not ip6) dst lookup miss to cpu.
	// V6 miss will be switched to cpu via punt next hop.
	t.rx_pipe_regs.cpu_control_1.set(q, 1<<10)

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

func (t *fe1a) snake_init() {
	t.cpu_rx_init()
	t.vlan_init()
	t.l2_init()
	t.l3_init()
}

func (t *fe1a) garbage_dump_init() {
	t.snake_init()
}
