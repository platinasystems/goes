package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

const (
	n_l2_dedicated_banks            = 2 // banks 0 and 1
	n_l2_dedicated_buckets_per_bank = 1 << 10
	n_l2_entry_per_bucket           = 4
)

type l2_entry_key_type uint8

const (
	l2_entry_key_type_outer_vlan_and_mac l2_entry_key_type = iota
	l2_entry_key_type_outer_vlan
	l2_entry_key_type_outer_and_inner_vlan
	l2_entry_key_type_vfi_and_mac
	l2_entry_key_type_niv_vif
	l2_entry_key_type_trill_unicast
	l2_entry_key_type_trill_non_unicast_tree_vlan_and_mac
	l2_entry_key_type_trill_non_unicast_tree_vlan
	_
	l2_entry_key_type_port_extender_etag
)

type l2_entry struct {
	valid    bool
	key_type l2_entry_key_type
	// 14 bits
	vlan_vfi uint16
	m.EthernetAddress
}

func (r *l2_entry) MemBits() int { return 105 }
func (r *l2_entry) MemGetSet(b []uint32, isSet bool) {
	panic("not yet")
}

type l2_entry_mem m.MemElt

func (r *l2_entry_mem) geta(q *DmaRequest, v *l2_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l2_entry_mem) seta(q *DmaRequest, v *l2_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l2_entry_mem) get(q *DmaRequest, v *l2_entry) { r.geta(q, v, sbus.DataSplit0) }
func (r *l2_entry_mem) set(q *DmaRequest, v *l2_entry) { r.seta(q, v, sbus.DataSplit0) }

const n_l2_user_entry = 1 << 9

type l2_user_entry_type m.TcamUint8

const (
	l2_user_entry_type_vlan_mac l2_user_entry_type = iota
	l2_user_entry_type_vlan_vfi
)

type l2_user_entry_key struct {
	typ l2_user_entry_type
	// Vlan or VFI depending on entry type.
	vlan_vfi m.TcamUint16
	m.EthernetAddress
}

// What to do station entry matches.
type l2_user_entry_data struct {
	cpu                                   bool
	drop                                  bool
	disable_src_ethernet_address_learning bool
	is_bpdu                               bool
	is_l2_protocol_packet                 bool

	priority_valid bool
	priority       uint8

	ifpClass uint16
	m.LogicalPort
}

type l2_user_entry struct {
	valid     bool
	data      l2_user_entry_data
	key, mask l2_user_entry_key
}

func (r *l2_user_entry_key) getSet(b []uint32, lo int, isSet bool) int {
	i := r.EthernetAddress.MemGetSet(b, lo, isSet)
	i = m.MemGetSetUint16((*uint16)(&r.vlan_vfi), b, i+13, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&r.typ), b, i, i, isSet)
	return lo + 80
}

func (key l2_user_entry_type) tcamEncode(mask l2_user_entry_type, isSet bool) (x, y l2_user_entry_type) {
	a, b := m.TcamUint8(key).TcamEncode(m.TcamUint8(mask), isSet)
	x, y = l2_user_entry_type(a), l2_user_entry_type(b)
	return
}

func (key *l2_user_entry_key) tcamEncode(mask *l2_user_entry_key, isSet bool) (x, y l2_user_entry_key) {
	x.vlan_vfi, y.vlan_vfi = key.vlan_vfi.TcamEncode(mask.vlan_vfi, isSet)
	x.typ, y.typ = key.typ.tcamEncode(mask.typ, isSet)
	x.EthernetAddress, y.EthernetAddress = key.EthernetAddress.TcamEncode(&mask.EthernetAddress, isSet)
	return
}

func (r *l2_user_entry_data) getSet(b []uint32, lo int, isSet bool) int {
	i := lo
	i = m.MemGetSetUint8(&r.priority, b, i+4, i, isSet)
	i = m.MemGetSet1(&r.priority_valid, b, i, isSet)
	i = m.MemGetSet1(&r.cpu, b, i, isSet)
	i = m.MemGetSet1(&r.drop, b, i, isSet)
	i = r.LogicalPort.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&r.disable_src_ethernet_address_learning, b, i, isSet)
	i = m.MemGetSetUint16(&r.ifpClass, b, i+9, i, isSet)
	i = m.MemGetSet1(&r.is_bpdu, b, i, isSet)
	i = m.MemGetSet1(&r.is_l2_protocol_packet, b, i, isSet)
	return i
}

func (r *l2_user_entry) MemBits() int { return 200 }
func (r *l2_user_entry) MemGetSet(b []uint32, isSet bool) {
	i := m.MemGetSet1(&r.valid, b, 0, isSet)
	var key, mask l2_user_entry_key
	if isSet {
		key, mask = r.key.tcamEncode(&r.mask, isSet)
	}
	i = key.getSet(b, i, isSet)
	i = mask.getSet(b, i, isSet)
	if !isSet {
		key, mask = r.key.tcamEncode(&r.mask, isSet)
	}
	i = r.data.getSet(b, i, isSet)
}

type l2_user_entry_mem m.MemElt

func (r *l2_user_entry_mem) geta(q *DmaRequest, v *l2_user_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l2_user_entry_mem) seta(q *DmaRequest, v *l2_user_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l2_user_entry_mem) get(q *DmaRequest, v *l2_user_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l2_user_entry_mem) set(q *DmaRequest, v *l2_user_entry) {
	r.seta(q, v, sbus.Duplicate)
}
