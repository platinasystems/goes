// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"
)

type adjacency struct {
	rx      rx_next_hop_entry
	tx      tx_next_hop_entry
	counter tx_pipe_pipe_counter_ref
}

//go:generate gentemplate -d Package=fe1a -id adjacency -d PoolType=adjacency_pool -d Type=adjacency -d Data=entries github.com/platinasystems/go/elib/pool.tmpl

const (
	adj_index_invalid = iota // hardware enforced invalid
	n_special_adj_index
)

type adjacency_main struct {
	adjacency_pool
	family_adj_index_by_ai [ip.NFamily]elib.Uint32Vec
	special_adj_index      [n_special_adj_index]uint32
}

func (t *fe1a) adjacency_main_init() {
	am := &t.adjacency_main
	am.adjacency_pool.SetMaxLen(n_next_hop)
	for i := uint(0); i < n_special_adj_index; i++ {
		ai := am.adjacency_pool.GetIndex()
		if ai != i {
			panic("special adjacency index must agree")
		}
		am.special_adj_index[i] = uint32(ai)
	}
	im4 := ip4.GetMain(t.Vnet)
	im4.RegisterAdjAddDelHook(t.adj_add_del)
	im4.RegisterFibAddDelHook(t.ip4_fib_add_del)
	im4.RegisterAdjSyncCounterHook(t.adj_sync_counters)
	im4.RegisterAdjGetCounterHook(t.adj_get_counter)

	// Configure ECMP hash control.  We'll hash on src,dst ip plus L4 ports.
	{
		q := t.getDmaReq()

		const (
			enable_ip_tcp_udp_ports = 1 << 22
			enable_ip_dst_address   = 1 << 12
		)

		v := t.rx_pipe_regs.hash_control.getDo(q, sbus.Duplicate)
		v |= enable_ip_dst_address | enable_ip_tcp_udp_ports
		t.rx_pipe_regs.hash_control.set(q, v)
		q.Do()
	}
}

func (am *adjacency_main) get_adj(ai uint) *adjacency { return &am.adjacency_pool.entries[ai] }
func (am *adjacency_main) get_ip_adj(im *ip.Main, adj ip.Adj) (ai uint, a *adjacency) {
	ai = uint(am.family_adj_index_by_ai[im.Family][uint(adj)])
	a = am.get_adj(ai)
	return
}

func (t *fe1a) adj_add_del(im *ip.Main, adj ip.Adj, isDel bool) {
	am := &t.adjacency_main

	if isDel {
		ai, a := am.get_ip_adj(im, adj)
		if a.counter.is_valid() {
			a.counter.free(t, BlockTxPipe)
		}
		am.adjacency_pool.PutIndex(ai)
		am.family_adj_index_by_ai[im.Family][uint(ai)] = ^uint32(0)
		*a = adjacency{}
		return
	}

	as := im.GetAdj(adj)
	if n_adj := len(as); n_adj > 1 {
		// nhs := im.NextHopsForAdj(adj)
		panic("ecmp") // not yet
	}

	a := &adjacency{}

	// Nothing to do for special adjacencies.
	tx := &l3_unicast_tx_next_hop{}
	pipe_mask := uint(1)<<n_pipe - 1
	disable_counter := false
	switch as[0].LookupNextIndex {
	case ip.LookupNextRewrite:
		si := as[0].Rewrite.Si
		port := port_for_si(t.Vnet, si)
		if port == nil {
			return
		}
		intf := t.l3_interface_for_si(si)
		a.rx = rx_next_hop_entry{
			rx_next_hop_type: rx_next_hop_type_tunnel, // index is l3_oif
			index:            uint16(intf.index()),
		}
		a.rx.LogicalPort.Set(uint(port.physical_port_number.toGpp()))
		tx.l3_intf_index = uint16(intf.index())
		h := (*ethernet.Header)(as[0].Rewrite.GetData())
		tx.dst_ethernet_address = m.EthernetAddress(h.Dst)
		pipe_mask = 1 << port.physical_port_number.pipe()
	case ip.LookupNextDrop:
		// Unfortunately no way to attach counter to drop adjacency.
		a.rx.drop = true
		disable_counter = true
	case ip.LookupNextPunt, ip.LookupNextLocal, ip.LookupNextGlean:
		// Switch to cpu port.
		port := t.port_by_phys_port[phys_port_cpu]
		intf := t.l3_interface_for_si(port.HwIf.Si())
		a.rx = rx_next_hop_entry{
			rx_next_hop_type: rx_next_hop_type_tunnel, // index is l3_oif
			index:            uint16(intf.index()),
		}
		a.rx.LogicalPort.Set(uint(port.physical_port_number.toGpp()))
		tx.l3_intf_index = uint16(intf.index())
		tx.disable_dst_ethernet_address_rewrite = true
		tx.disable_src_ethernet_address_rewrite = true
		tx.disable_l3_unicast_vlan_rewrite = true
		tx.disable_ip_ttl_decrement = true
		pipe_mask = 1 << port.physical_port_number.pipe()
	default:
		panic("unknown adj")
	}

	ai := am.adjacency_pool.GetIndex()
	if !disable_counter {
		a.counter.alloc(t, pipe_counter_pool_tx_adjacency, pipe_mask, BlockTxPipe)
		tx.pipe_counter_ref = a.counter
	}
	a.tx = tx
	*am.get_adj(ai) = *a

	am.family_adj_index_by_ai[im.Family].Validate(uint(ai))
	am.family_adj_index_by_ai[im.Family][adj] = uint32(ai)

	q := t.getDmaReq()
	t.rx_pipe_mems.l3_next_hop[ai].set(q, &a.rx)
	t.tx_pipe_mems.l3_next_hop[ai].set(q, a.tx)
	q.Do()
}

func (t *fe1a) adj_sync_counters(m *ip.Main) {
	t.update_pool_counter_values(pipe_counter_pool_tx_adjacency, BlockTxPipe)
}

func (t *fe1a) adj_get_counter(im *ip.Main, adj ip.Adj, f ip.AdjGetCounterHandler) {
	_, a := t.get_ip_adj(im, adj)
	v := a.counter.get_value(t, BlockTxPipe)
	f(tx_counter_prefix, v)
}

type ip4_fib_key struct {
	ip.FibIndex
	ip4.Address
}

type ip4_fib_prefix_len struct {
	index_by_key   map[ip4_fib_key]uint32
	len            uint
	index          uint
	n_half_entries uint32
	base_index     uint32
	elib.Pool
}

type ip4_fib_main struct {
	// /32 /31 ... /0
	prefix_lens      [33]ip4_fib_prefix_len
	fib_tcam_entries [n_fib_tcam_entries]fib_tcam_entry
}

func (t *fe1a) ip4_fib_init() {
	fm := &t.ip4_fib_main
	for i := range fm.prefix_lens {
		fm.prefix_lens[i].index = uint(i)
		fm.prefix_lens[i].len = uint(32 - i)
	}
}

const log2_fib_tcam_alloc_unit = 5 // 32 double entries

func (pl *ip4_fib_prefix_len) shift_up(t *fe1a, q *DmaRequest, n_half_entries uint32) {
	fm := &t.ip4_fib_main
	for a, i := range pl.index_by_key {
		iʹ := i + n_half_entries
		i0, i0ʹ := i/2, iʹ/2
		e0, e0ʹ := &fm.fib_tcam_entries[i0], &fm.fib_tcam_entries[i0ʹ]
		*e0ʹ = *e0
		t.rx_pipe_mems.fib_tcam[i0ʹ].set(q, e0ʹ)
		e0[0].is_valid = false
		e0[1].is_valid = false
		t.rx_pipe_mems.fib_tcam[i0].set(q, e0)
		pl.index_by_key[a] = iʹ
	}
	if len(q.Commands) > 256 {
		q.Do()
	}
}

func (pl *ip4_fib_prefix_len) alloc(t *fe1a) (ei uint32) {
	fm := &t.ip4_fib_main
	l := uint(pl.n_half_entries)
	var i uint
	q := t.getDmaReq()
	if i = pl.Pool.GetIndex(l); i == l {
		u := uint(log2_fib_tcam_alloc_unit)
		if pl.len < log2_fib_tcam_alloc_unit {
			u = pl.len
		}
		delta := uint32(2) << u // 2 since each entry is 2 half entries
		pl.n_half_entries += delta
		for j := uint(len(fm.prefix_lens)) - 1; j > pl.index; j-- {
			plʹ := &fm.prefix_lens[j]
			if plʹ.base_index+delta >= 2*n_fib_tcam_entries {
				panic("fib tcam overflow; more than 16k entries")
			}
			if plʹ.n_half_entries > 0 {
				plʹ.shift_up(t, q, delta)
			}
			plʹ.base_index += delta
		}

		// Perform any remaining dma commands.
		q.Do()

		// Free all entries after first.
		for j := uint(1); j < uint(delta); j++ {
			pl.Pool.PutIndex(i + j)
		}
	}
	ei = uint32(i) + pl.base_index
	return
}

func (m *ip4_fib_main) free(t *fe1a, i uint32) {
	i0, i1 := i/2, i%2
	e := &m.fib_tcam_entries[i0]
	e[i1].is_valid = false
	var f fib_tcam_tcam_only_entry
	f[0] = e[0].fib_tcam_tcam_search
	f[1] = e[1].fib_tcam_tcam_search
	q := t.getDmaReq()
	t.rx_pipe_mems.fib_tcam_only[i0].set(q, &f)
	q.Do()
}

func (t *fe1a) ip4_fib_add_del(fib_index ip.FibIndex, p *ip4.Prefix, adj ip.Adj, isDel bool, isRemap bool) {
	am := &t.adjacency_main
	fm := &t.ip4_fib_main
	fl := &fm.prefix_lens[32-p.Len]

	key := ip4_fib_key{Address: p.Address, FibIndex: fib_index}

	if isDel {
		if i, ok := fl.index_by_key[key]; !ok {
			panic("not found")
		} else {
			fm.free(t, i)
			delete(fl.index_by_key, key)
			fl.Pool.PutIndex(uint(i - fl.base_index))
		}
	} else {
		ai := uint16(am.family_adj_index_by_ai[ip.Ip4][adj])
		if i, ok := fl.index_by_key[key]; ok {
			i0, i1 := i/2, i%2
			e := &fm.fib_tcam_entries[i0]
			if e[i1].next_hop.IsECMP {
				panic("ecmp")
			}
			if e[i1].next_hop.Index != ai {
				e[i1].next_hop.Index = ai
				var f fib_tcam_tcam_data_only_entry
				f[0] = e[0].fib_tcam_tcam_data
				f[1] = e[1].fib_tcam_tcam_data
				q := t.getDmaReq()
				t.rx_pipe_mems.fib_tcam_data_only[i0].set(q, &f)
				q.Do()
			}
		} else {
			i := fl.alloc(t)
			i0, i1 := i/2, i%2
			e := &fm.fib_tcam_entries[i0]
			e[i1] = fib_tcam_half_entry{
				fib_tcam_tcam_search: fib_tcam_tcam_search{
					key: fib_tcam_tcam_key{
						key_type:   fib_tcam_ip4,
						Vrf:        m.Vrf(fib_index),
						Ip4Address: m.Ip4Address(p.Address),
					},
					mask: fib_tcam_tcam_key{
						key_type:   0xff,
						Vrf:        ^m.Vrf(0),
						Ip4Address: m.Ip4Address(p.MaskAsAddress()),
					},
					is_valid: true,
				},
				fib_tcam_tcam_data: fib_tcam_tcam_data{
					next_hop: m.NextHop{Index: ai},
				},
			}
			q := t.getDmaReq()
			t.rx_pipe_mems.fib_tcam[i0].set(q, e)
			q.Do()
			if fl.index_by_key == nil {
				fl.index_by_key = make(map[ip4_fib_key]uint32)
			}
			fl.index_by_key[key] = i
		}
	}
}
