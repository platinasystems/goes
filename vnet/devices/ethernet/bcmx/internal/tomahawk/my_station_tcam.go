package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

const n_my_station_tcam_entry = 1 << 10

// Key and mask for station TCAM.  Match when search & mask == key */
type my_station_tcam_key struct {
	m.LogicalPort
	m.Vlan
	m.EthernetAddress
}

func (key *my_station_tcam_key) tcamEncode(mask *my_station_tcam_key, isSet bool) (x, y my_station_tcam_key) {
	x.LogicalPort, y.LogicalPort = key.LogicalPort.TcamEncode(&mask.LogicalPort, isSet)
	x.Vlan, y.Vlan = key.Vlan.TcamEncode(mask.Vlan, isSet)
	x.EthernetAddress, y.EthernetAddress = key.EthernetAddress.TcamEncode(&mask.EthernetAddress, isSet)
	return
}

// What to do station entry matches.
type my_station_tcam_data struct {
	copy_to_cpu bool
	drop        bool
	// Enable termination of various types of packets.
	ip4_unicast_enable   bool
	ip6_unicast_enable   bool
	ip4_multicast_enable bool
	ip6_multicast_enable bool
	mpls_enable          bool
	arp_rarp_enable      bool
	fcoe_enable          bool
	trill_enable         bool
	mac_in_mac_enable    bool
}

type my_station_tcam_entry struct {
	data      my_station_tcam_data
	valid     bool
	key, mask my_station_tcam_key
}

func (r *my_station_tcam_key) getSet(b []uint32, lo int, isSet bool) int {
	i := r.EthernetAddress.MemGetSet(b, lo, isSet)
	i = r.Vlan.MemGetSet(b, i, isSet)
	i = r.LogicalPort.MemGetSet(b, i, isSet)
	return lo + 80
}

func (r *my_station_tcam_data) getSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSet1(&r.mac_in_mac_enable, b, i, isSet)
	i = m.MemGetSet1(&r.mpls_enable, b, i, isSet)
	i = m.MemGetSet1(&r.trill_enable, b, i, isSet)
	i = m.MemGetSet1(&r.ip4_unicast_enable, b, i, isSet)
	i = m.MemGetSet1(&r.ip6_unicast_enable, b, i, isSet)
	i = m.MemGetSet1(&r.arp_rarp_enable, b, i, isSet)
	i = m.MemGetSet1(&r.fcoe_enable, b, i, isSet)
	i = m.MemGetSet1(&r.ip4_multicast_enable, b, i, isSet)
	i = m.MemGetSet1(&r.ip6_multicast_enable, b, i, isSet)
	i = m.MemGetSet1(&r.drop, b, i, isSet)
	i = m.MemGetSet1(&r.copy_to_cpu, b, i, isSet)
	// bit 11 is reserved
	i += 1
	return i
}

func (r *my_station_tcam_entry) MemBits() int { return 174 }
func (r *my_station_tcam_entry) MemGetSet(b []uint32, isSet bool) {
	i := m.MemGetSet1(&r.valid, b, 0, isSet)
	var key, mask my_station_tcam_key
	if isSet {
		key, mask = r.key.tcamEncode(&r.mask, isSet)
	}
	i = key.getSet(b, i, isSet)
	if i != 81 {
		panic("81")
	}
	i = mask.getSet(b, i, isSet)
	if i != 161 {
		panic("161")
	}
	if !isSet {
		key, mask = r.key.tcamEncode(&r.mask, isSet)
	}
	i = r.data.getSet(b, i, isSet)
}

type my_station_tcam_mem m.MemElt

func (r *my_station_tcam_mem) geta(q *DmaRequest, v *my_station_tcam_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *my_station_tcam_mem) seta(q *DmaRequest, v *my_station_tcam_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *my_station_tcam_mem) get(q *DmaRequest, v *my_station_tcam_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *my_station_tcam_mem) set(q *DmaRequest, v *my_station_tcam_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type my_station_tcam_entry_only_mem m.MemElt
type my_station_tcam_data_only_mem m.MemElt

//go:generate gentemplate -d Package=tomahawk -id my_station_tcam -d PoolType=my_station_tcam_pool -d Type=my_station_tcam_entry -d Data=entries github.com/platinasystems/go/elib/pool.tmpl

type my_station_tcam_main struct {
	pool             my_station_tcam_pool
	poolIndexByEntry map[my_station_tcam_entry]uint
}

func (tm *my_station_tcam_main) addDel(t *tomahawk, e *my_station_tcam_entry, isDel bool) (i uint, ok bool) {
	if tm.poolIndexByEntry == nil {
		tm.poolIndexByEntry = make(map[my_station_tcam_entry]uint)
		tm.pool.SetMaxLen(n_my_station_tcam_entry)
	}

	q := t.getDmaReq()
	f := my_station_tcam_entry{}
	f.key = e.key
	f.mask = e.mask
	if i, ok = tm.poolIndexByEntry[f]; !ok && isDel {
		return
	}
	if isDel {
		pe := &tm.pool.entries[i]
		pe.valid = false
		t.rx_pipe_mems.my_station_tcam[i].set(q, pe)
		tm.pool.PutIndex(i)
		delete(tm.poolIndexByEntry, f)
	} else {
		if !ok {
			i = tm.pool.GetIndex()
			tm.poolIndexByEntry[f] = i
		}
		pe := &tm.pool.entries[i]
		*pe = *e
		pe.valid = true
		t.rx_pipe_mems.my_station_tcam[i].set(q, pe)
	}
	q.Do()
	return
}
