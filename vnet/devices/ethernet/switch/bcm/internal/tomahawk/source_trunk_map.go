// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/sbus"
)

type source_port_type uint8

type source_trunk_map_main struct{}

const (
	source_port_type_normal source_port_type = iota
	source_port_type_trunk
	_
	source_port_type_hp_mesh
)

func (t *source_port_type) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(t), b, i+1, i, isSet)
}

type source_trunk_map_entry struct {
	source_port_type

	trunk_id uint16

	ifp_class_id uint16

	// 8 bit index into ing_vlan_range compression table.
	outer_vlan_range_compression_index uint8
	inner_vlan_range_compression_index uint8

	// vrf/outer vlan/l3 iif depending on setting in lport_profile_table.
	index uint16

	vfp_port_group_id uint8
	port_group_id     uint16

	src_virtual_port       uint16
	src_virtual_port_valid bool

	ip4_multicast_l2_enable                          bool
	ip6_multicast_l2_enable                          bool
	disable_vlan_membership_and_spanning_tree_checks bool

	trill_rbridge_nickname_index uint8

	lport_profile_index uint8
	ipbm_index          uint8
}

func (e *source_trunk_map_entry) MemBits() int { return 122 }
func (e *source_trunk_map_entry) MemGetSet(b []uint32, isSet bool) {
	i := e.source_port_type.MemGetSet(b, 0, isSet)
	i = m.MemGetSetUint16(&e.trunk_id, b, i+10, i, isSet)
	i = 17 // skip reserved bits [16:13]
	i = m.MemGetSetUint16(&e.ifp_class_id, b, i+11, i, isSet)
	i = m.MemGetSetUint8(&e.outer_vlan_range_compression_index, b, i+7, i, isSet)
	i = m.MemGetSetUint16(&e.index, b, i+13, i, isSet)
	i = m.MemGetSetUint8(&e.vfp_port_group_id, b, i+7, i, isSet)
	i = m.MemGetSetUint8(&e.trill_rbridge_nickname_index, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.inner_vlan_range_compression_index, b, i+7, i, isSet)
	i = m.MemGetSetUint16(&e.port_group_id, b, i+11, i, isSet)
	i = m.MemGetSetUint16(&e.src_virtual_port, b, i+14, i, isSet)
	i = m.MemGetSet1(&e.src_virtual_port_valid, b, i, isSet)
	i = m.MemGetSet1(&e.ip6_multicast_l2_enable, b, i, isSet)
	i = m.MemGetSet1(&e.ip4_multicast_l2_enable, b, i, isSet)
	i = m.MemGetSet1(&e.disable_vlan_membership_and_spanning_tree_checks, b, i, isSet)
	i = m.MemGetSetUint8(&e.lport_profile_index, b, i+7, i, isSet)
	i = m.MemGetSetUint8(&e.ipbm_index, b, i+5, i, isSet)
	if i != 114 {
		panic("114")
	}
}

type source_trunk_map_mem m.MemElt

func (r *source_trunk_map_mem) geta(q *DmaRequest, v *source_trunk_map_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *source_trunk_map_mem) seta(q *DmaRequest, v *source_trunk_map_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *source_trunk_map_mem) get(q *DmaRequest, v *source_trunk_map_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *source_trunk_map_mem) set(q *DmaRequest, v *source_trunk_map_entry) {
	r.seta(q, v, sbus.Duplicate)
}
