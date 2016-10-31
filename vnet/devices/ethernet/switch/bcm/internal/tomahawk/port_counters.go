// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/port"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/sbus"

	"fmt"
)

type port_counter_main struct {
	vnet.InterfaceCounterNames

	// Port (MIB) counters.
	kindByPortCounterType map[port.Counter_type]vnet.HwIfCounterKind
	portCounterTypeByKind map[vnet.HwIfCounterKind]port.Counter_type

	// Rx pipe counters
	kindByRxCounterType map[rx_counter_type]vnet.HwIfCounterKind
	rxCounterTypeByKind map[vnet.HwIfCounterKind]rx_counter_type

	// Tx pipe counters
	kindByTxCounterType map[tx_counter_type]vnet.HwIfCounterKind
	txCounterTypeByKind map[vnet.HwIfCounterKind]tx_counter_type

	kindForPortFlexCounter [vnet.NRxTx]vnet.HwIfCounterKind

	// 2 * 2 * 10 counters: unicast/multicast packets/bytes 10 queues
	txPerqCounterKind    [m.N_cast]vnet.HwIfCounterKind
	txPerqCpuCounterKind vnet.HwIfCounterKind

	// Shadow values of rx_pipe/tx_pipe counters.
	rx_counters_last_value [n_pipe_ports][n_rx_counters]uint64
	tx_counters_last_value [n_pipe_ports][n_tx_counters]uint64

	// Mmu counters

	// 2 counters packets, bytes
	mmu_rx_drops_kind vnet.HwIfCounterKind

	// 2 * 10 counters packets, bytes
	mmu_tx_drops_kind [m.N_cast]vnet.HwIfCounterKind

	mmu_wred_drops_kind vnet.HwIfCounterKind
}

func (t *tomahawk) get_port_counters(p *Port, th *vnet.InterfaceThread) {
	cm := &t.port_counter_main
	hi := p.Hi()

	if p.PortBlock != nil { // cpu port for example does not have a port block.
		pb := p.PortBlock.(*port.PortBlock)
		var v [port.N_counters]uint64
		pb.GetCounters(p, &v)
		for i := range v {
			if kind, ok := cm.kindByPortCounterType[port.Counter_type(i)]; ok {
				kind.Add64(th, hi, v[i])
			}
		}
	}

	q := t.getDmaReq()
	phys_port := p.physical_port_number
	pipe_port := phys_port.toPipe()
	pipe := uint(phys_port.pipe())

	// Read Rx_pipe counters.
	{
		var v [n_rx_counters]uint64
		for i := range rx_counter_names {
			t.rx_pipe_regs.counters[i][pipe_port].geta(q, sbus.AddressSplit, &v[i])
		}
		q.Do()
		for i := range rx_counter_names {
			if kind, ok := cm.kindByRxCounterType[rx_counter_type(i)]; ok {
				// Bit [32] is parity bit; [31:0] is counter.
				x := v[i] & 0xffffffff
				kind.Add64(th, hi, x-cm.rx_counters_last_value[pipe_port][i])
				cm.rx_counters_last_value[pipe_port][i] = x
			}
		}
	}

	// Read Rx_pipe per port flex counter.
	{
		v := t.rx_pipe_port_table[pipe_port].flex_counter.update_value(t, pipe, BlockRxPipe)
		(cm.kindForPortFlexCounter[vnet.Rx] + 0).Add64(th, hi, v.Packets)
		(cm.kindForPortFlexCounter[vnet.Rx] + 1).Add64(th, hi, v.Bytes)
	}

	// Read Tx_pipe counters.
	{
		var v [n_tx_counters]uint64
		for i := range tx_counter_names {
			t.tx_pipe_regs.counters[i][pipe_port].geta(q, sbus.AddressSplit, &v[i])
		}
		q.Do()
		for i := range tx_counter_names {
			if kind, ok := cm.kindByTxCounterType[tx_counter_type(i)]; ok {
				// Bit [32] is parity bit; [31:0] is counter.
				x := v[i] & 0xffffffff
				kind.Add64(th, hi, x-cm.tx_counters_last_value[pipe_port][i])
				cm.tx_counters_last_value[pipe_port][i] = x
			}
		}
	}

	// Read Tx_pipe per port flex counter.
	{
		v := t.tx_pipe_port_table[pipe_port].flex_counter.update_value(t, pipe, BlockTxPipe)
		(cm.kindForPortFlexCounter[vnet.Tx] + 0).Add64(th, hi, v.Packets)
		(cm.kindForPortFlexCounter[vnet.Tx] + 1).Add64(th, hi, v.Bytes)
	}

	// Tx_pipe per q counters.
	{
		_, tx_pipe := phys_port.toIdb()
		// No need to simulate clear on read; since we've initialized all flex counters to clear on read.
		if phys_port == phys_port_cpu {
			var v [mmu_n_cpu_queues]flex_counter_entry
			for txq := 0; txq < mmu_n_cpu_queues; txq++ {
				t.tx_pipe_mems.per_tx_queue_counters.cpu[txq].geta(q, uint(tx_pipe), &v[txq])
			}
			q.Do()
			for txq := 0; txq < mmu_n_cpu_queues; txq++ {
				k := vnet.HwIfCounterKind(txq)
				(cm.txPerqCpuCounterKind + 2*k + 0).Add64(th, hi, v[txq].Packets)
				(cm.txPerqCpuCounterKind + 2*k + 1).Add64(th, hi, v[txq].Bytes)
			}
		} else {
			var v [m.N_cast][mmu_n_tx_queues]flex_counter_entry
			pi := pipe_port % 34
			for txq := 0; txq < mmu_n_tx_queues; txq++ {
				t.tx_pipe_mems.per_tx_queue_counters.ports[pi].unicast[txq].geta(q, uint(tx_pipe), &v[m.Unicast][txq])
				t.tx_pipe_mems.per_tx_queue_counters.ports[pi].multicast[txq].geta(q, uint(tx_pipe), &v[m.Multicast][txq])
			}
			q.Do()
			for cast := m.Cast(0); cast < m.N_cast; cast++ {
				for txq := 0; txq < mmu_n_tx_queues; txq++ {
					k := vnet.HwIfCounterKind(txq)
					(cm.txPerqCounterKind[cast] + 2*k + 0).Add64(th, hi, v[cast][txq].Packets)
					(cm.txPerqCounterKind[cast] + 2*k + 1).Add64(th, hi, v[cast][txq].Bytes)
				}
			}
		}

		// Tx perq counters may be in eviction fifo; so sync it here.
		t.flex_counter_eviction_fifo_sync()
	}

	// MMU rx port counters
	{
		var (
			v    [2]mmu_rx_counter_entry
			zero mmu_rx_counter_entry
		)

		mmu_port, mmu_rx_pipe := phys_port.toMmu(t)
		// [0] => xpe 0 pipes 0 [0..33] & 3 [34..67]
		// [1] => xpe 1 pipes 0 [0..33] & 3 [34..67]
		// [2] => xpe 2 pipes 1 & 2
		// [3] => xpe 3 pipes 1 & 2
		xpe, i := 0, mmu_rx_pipe&1
		if mmu_rx_pipe == 1 || mmu_rx_pipe == 2 {
			xpe = 2
			i = (mmu_rx_pipe + 1) & 1
		}
		for x := 0; x < 2; x++ {
			t.mmu_xpe_mems.rx_drops[xpe+x].entries[i][mmu_port].get(q, &v[x])
			// clear on read
			t.mmu_xpe_mems.rx_drops[xpe+x].entries[i][mmu_port].set(q, &zero)
		}
		q.Do()
		(cm.mmu_rx_drops_kind + 0).Add64(th, hi, v[0].Packets+v[1].Packets)
		(cm.mmu_rx_drops_kind + 1).Add64(th, hi, v[0].Bytes+v[1].Bytes)
	}

	{
		var (
			v    [m.N_cast][mmu_n_tx_queues][2]mmu_tx_counter_entry
			zero mmu_tx_counter_entry
		)

		mmu_port, mmu_tx_pipe := phys_port.toMmu(t)
		// tx_pipes 0 and 1 => xpe 0/2; else xpe 1/3
		baseXpe := 0
		if mmu_tx_pipe >= 2 {
			baseXpe = 1
		}
		for cast := m.Cast(0); cast < m.N_cast; cast++ {
			for x := 0; x < 2; x++ {
				xpe := uint(baseXpe + 2*x)
				for txq := 0; txq < mmu_n_tx_queues; txq++ {
					if cast == m.Unicast {
						t.mmu_xpe_mems.unicast_tx_drops[mmu_tx_pipe].entries[mmu_port][txq].get(q, xpe, &v[cast][txq][x])
						t.mmu_xpe_mems.unicast_tx_drops[mmu_tx_pipe].entries[mmu_port][txq].set(q, xpe, &zero)
					} else {
						switch {
						case mmu_port < n_mmu_data_port:
							t.mmu_xpe_mems.multicast_tx_drops[mmu_tx_pipe].ports[mmu_port][txq].get(q, xpe, &v[cast][txq][x])
							t.mmu_xpe_mems.multicast_tx_drops[mmu_tx_pipe].ports[mmu_port][txq].set(q, xpe, &zero)
						case mmu_port == idb_mmu_port_loopback:
							t.mmu_xpe_mems.multicast_tx_drops[mmu_tx_pipe].loopback[txq].get(q, xpe, &v[cast][txq][x])
							t.mmu_xpe_mems.multicast_tx_drops[mmu_tx_pipe].loopback[txq].set(q, xpe, &zero)
						case mmu_port == idb_mmu_port_any_pipe_cpu_or_management:
							t.mmu_xpe_mems.multicast_tx_drops[mmu_tx_pipe].cpu[txq].get(q, xpe, &v[cast][txq][x])
							t.mmu_xpe_mems.multicast_tx_drops[mmu_tx_pipe].cpu[txq].set(q, xpe, &zero)
						}
					}
				}
			}
			q.Do()

			for txq := 0; txq < mmu_n_tx_queues; txq++ {
				k := vnet.HwIfCounterKind(txq)
				(cm.mmu_tx_drops_kind[cast] + 2*k + 0).Add64(th, hi, v[cast][txq][0].Packets+v[cast][txq][1].Packets)
				(cm.mmu_tx_drops_kind[cast] + 2*k + 1).Add64(th, hi, v[cast][txq][0].Bytes+v[cast][txq][1].Bytes)
			}
		}
	}

	{
		var (
			v    [mmu_n_tx_queues][2]mmu_wred_counter_entry
			zero mmu_wred_counter_entry
		)

		mmu_port, mmu_tx_pipe := phys_port.toMmu(t)
		baseXpe := 0
		if mmu_tx_pipe >= 2 {
			baseXpe = 1
		}
		for x := 0; x < 2; x++ {
			xpe := uint(baseXpe + 2*x)
			for txq := 0; txq < mmu_n_tx_queues; txq++ {
				if mmu_port < n_mmu_data_port {
					t.mmu_xpe_mems.wred_drops[mmu_tx_pipe].data_port_entries[mmu_port][txq].get(q, xpe, &v[txq][x])
					t.mmu_xpe_mems.wred_drops[mmu_tx_pipe].data_port_entries[mmu_port][txq].set(q, xpe, &zero)
				} else if txq < mmu_n_tx_cos_queues {
					i := mmu_port - n_mmu_data_port
					t.mmu_xpe_mems.wred_drops[mmu_tx_pipe].cpu_loopback_port_entries[i][txq].get(q, xpe, &v[txq][x])
					t.mmu_xpe_mems.wred_drops[mmu_tx_pipe].cpu_loopback_port_entries[i][txq].set(q, xpe, &zero)
				}
			}
		}
		q.Do()

		for txq := 0; txq < mmu_n_tx_queues; txq++ {
			k := vnet.HwIfCounterKind(txq)
			(cm.mmu_wred_drops_kind + k).Add64(th, hi, uint64(v[txq][0]+v[txq][1]))
		}
	}
}

func (t *tomahawk) add_zero_rx_pipe_port_counters_cmd(q *DmaRequest, p pipe_port_number) {
	var (
		zero [2 * n_rx_counters]uint32
	)
	zerocmd := sbus.DmaCmd{
		Command: sbus.Command{
			Opcode:     sbus.WriteRegister,
			Block:      BlockRxPipe,
			AccessType: sbus.AddressSplit,
		},
		Address: t.rx_pipe_regs.counters[0][p].address(),
		Tx:      zero[:],
		Count:   uint(n_rx_counters),
		Log2SbusAddressIncrement: m.Log2NRegPorts,
	}
	q.Add(&zerocmd)
}

// Configure rx debug counters.
func (t *tomahawk) rx_pipe_port_counter_init(kind vnet.HwIfCounterKind) vnet.HwIfCounterKind {
	m := &t.port_counter_main
	nm := &m.InterfaceCounterNames
	q := t.getDmaReq()

	rx := [...]struct {
		de   rx_debug_counter_type
		name string
	}{
		{rx_debug_zero_port_bitmap_drops, "zero port bitmap drops"},
		{rx_debug_vlan_drops, "unknown vlan drops"},
		{rx_debug_dst_discard_drop, "dst discard drops"},
	}
	all_mask := [2]uint32{}

	// first 10 counters are duplicated between debug counters and normal counters.
	for i := uint(0); i < uint(rx_debug_0); i++ {
		all_mask[0] |= 1 << i
	}

	for i := range rx {
		di, mask := rx[i].de/32, uint32(1<<(rx[i].de%32))
		if i >= len(t.rx_pipe_regs.debug_counter_select[di]) {
			break
		}
		t.rx_pipe_regs.debug_counter_select[di][i].set(q, mask)
		all_mask[di] |= mask
		rx_counter_names[int(rx_debug_0)+i] = rx[i].name
	}
	if len(rx) < len(t.rx_pipe_regs.debug_counter_select[0]) {
		t.rx_pipe_regs.debug_counter_select[0][len(rx)].set(q, ^uint32(0)&^all_mask[0])
		t.rx_pipe_regs.debug_counter_select[1][len(rx)].set(q, ^uint32(0)&^all_mask[1])
	}
	q.Do()

	m.kindForPortFlexCounter[vnet.Rx] = kind
	nm.Single = append(nm.Single, rx_counter_prefix+"port table packets")
	nm.Single = append(nm.Single, rx_counter_prefix+"port table bytes")
	kind += 2

	m.kindByRxCounterType = make(map[rx_counter_type]vnet.HwIfCounterKind)
	m.rxCounterTypeByKind = make(map[vnet.HwIfCounterKind]rx_counter_type)
	for i := range rx_counter_names {
		ct := rx_counter_type(i)
		n := rx_counter_names[i]
		nm.Single = append(nm.Single, rx_counter_prefix+n)
		m.kindByRxCounterType[ct] = kind
		m.rxCounterTypeByKind[kind] = ct
		kind++
	}

	for p := pipe_port_number(0); p < n_pipe_ports; p++ {
		t.add_zero_rx_pipe_port_counters_cmd(q, p)
	}
	q.Do()
	return kind
}

// Configure tx debug counters.
func (t *tomahawk) tx_pipe_port_counter_init(kind vnet.HwIfCounterKind) vnet.HwIfCounterKind {
	cm := &t.port_counter_main
	nm := &cm.InterfaceCounterNames
	q := t.getDmaReq()

	tx := [...]struct {
		de   tx_debug_counter_type
		name string
	}{
		{tx_debug_packets_dropped, "packets dropped"},
		{tx_debug_invalid_vlan_drops, "invalid vlan drops"},
		{tx_debug_spanning_tree_state_not_forwarding_drops, "spanning tree state not forwarding drops"},
		{tx_debug_packet_aged_drops, "packet aged drops"},
		{tx_debug_ip4_unicast_packets, "ip4 unicast packets"},
		{tx_debug_ip4_unicast_aged_and_drop_packets, "ip4 unicast aged and dropped packets"},
		{tx_debug_ip_length_check_drops, "ip length check drops"},
		{tx_debug_vlan_tagged_packets, "vlan tagged packets"},
	}
	all_mask := [2]uint32{}
	for i := range tx {
		di, mask := tx[i].de/32, uint32(1<<tx[i].de)
		if di > 1 {
			break
		}
		if di == 0 {
			t.tx_pipe_regs.debug_counter_select[i].set(q, mask)
		} else {
			t.tx_pipe_regs.debug_counter_select_hi[i].set(q, mask)
		}
		all_mask[di] |= mask
		tx_counter_names[int(tx_debug_0)+i] = tx[i].name
	}
	if len(tx) < len(t.tx_pipe_regs.debug_counter_select) {
		t.tx_pipe_regs.debug_counter_select[len(tx)].set(q, ^uint32(0)&^all_mask[0])
		t.tx_pipe_regs.debug_counter_select_hi[len(tx)].set(q, ^uint32(0)&^all_mask[1])
	}
	q.Do()

	cm.kindByTxCounterType = make(map[tx_counter_type]vnet.HwIfCounterKind)
	cm.txCounterTypeByKind = make(map[vnet.HwIfCounterKind]tx_counter_type)
	for i := range tx_counter_names {
		ct := tx_counter_type(i)
		n := tx_counter_names[i]
		nm.Single = append(nm.Single, tx_counter_prefix+n)
		cm.kindByTxCounterType[ct] = kind
		cm.txCounterTypeByKind[kind] = ct
		kind++
	}

	for cast := m.Cast(0); cast < m.N_cast; cast++ {
		cm.txPerqCounterKind[cast] = kind
		for i := 0; i < mmu_n_tx_queues; i++ {
			nm.Single = append(nm.Single, fmt.Sprintf(tx_counter_prefix+"%s queue %s packets", cast, mmu_tx_queue_names[i]))
			nm.Single = append(nm.Single, fmt.Sprintf(tx_counter_prefix+"%s queue %s bytes", cast, mmu_tx_queue_names[i]))
		}
		kind += 2 * mmu_n_tx_queues
	}

	cm.txPerqCpuCounterKind = kind
	for i := 0; i < mmu_n_cpu_queues; i++ {
		nm.Single = append(nm.Single, fmt.Sprintf(tx_counter_prefix+"cpu queue %d packets", i))
		nm.Single = append(nm.Single, fmt.Sprintf(tx_counter_prefix+"cpu queue %d bytes", i))
	}
	kind += 2 * mmu_n_cpu_queues

	cm.kindForPortFlexCounter[vnet.Tx] = kind
	nm.Single = append(nm.Single, tx_counter_prefix+"port table packets")
	nm.Single = append(nm.Single, tx_counter_prefix+"port table bytes")
	kind += 2

	return kind
}

func (t *tomahawk) mmu_port_counter_init(kind vnet.HwIfCounterKind) vnet.HwIfCounterKind {
	cm := &t.port_counter_main
	nm := &cm.InterfaceCounterNames

	// MMU rx per-port drop counters.
	cm.mmu_rx_drops_kind = kind
	nm.Single = append(nm.Single, "mmu rx threshold drop packets")
	nm.Single = append(nm.Single, "mmu rx threshold drop bytes")
	kind += 2

	// MMU tx per-port drop counters.
	cm.mmu_tx_drops_kind[m.Unicast] = kind
	for i := 0; i < mmu_n_tx_queues; i++ {
		nm.Single = append(nm.Single, fmt.Sprintf("mmu unicast tx %s drop packets", mmu_tx_queue_names[i]))
		nm.Single = append(nm.Single, fmt.Sprintf("mmu unicast tx %s drop bytes", mmu_tx_queue_names[i]))
	}
	kind += 2 * mmu_n_tx_queues

	cm.mmu_wred_drops_kind = kind
	for i := 0; i < mmu_n_tx_queues; i++ {
		nm.Single = append(nm.Single, fmt.Sprintf("mmu wred queue %s drop packets", mmu_tx_queue_names[i]))
	}
	kind += mmu_n_tx_queues

	cm.mmu_tx_drops_kind[m.Multicast] = kind
	for i := 0; i < mmu_n_tx_queues; i++ {
		nm.Single = append(nm.Single, fmt.Sprintf("mmu multicast tx %s drop packets", mmu_tx_queue_names[i]))
		nm.Single = append(nm.Single, fmt.Sprintf("mmu multicast tx %s drop bytes", mmu_tx_queue_names[i]))
	}
	kind += 2 * mmu_n_tx_queues

	return kind
}

func (t *tomahawk) mac_port_counter_init(rt vnet.RxTx, kind vnet.HwIfCounterKind) vnet.HwIfCounterKind {
	m := &t.port_counter_main
	nm := &m.InterfaceCounterNames

	if m.kindByPortCounterType == nil {
		m.kindByPortCounterType = make(map[port.Counter_type]vnet.HwIfCounterKind)
		m.portCounterTypeByKind = make(map[vnet.HwIfCounterKind]port.Counter_type)
	}

	for i := range port.Counter_order[rt] {
		ct := port.Counter_order[rt][i]
		n := port.CounterNames[ct]
		nm.Single = append(nm.Single, "port "+n)
		m.kindByPortCounterType[ct] = kind
		m.portCounterTypeByKind[kind] = ct
		kind++
	}
	return kind
}

func (t *tomahawk) port_counter_init() {
	kind := vnet.HwIfCounterKind(0)

	// Order counters: MAC rx, rx pipe, MMU, tx pipe, Mac tx to model packet processing pipeline.
	// Show commands will display counters in this order.
	kind = t.mac_port_counter_init(vnet.Rx, kind)
	kind = t.rx_pipe_port_counter_init(kind)
	kind = t.mmu_port_counter_init(kind)
	kind = t.tx_pipe_port_counter_init(kind)
	kind = t.mac_port_counter_init(vnet.Tx, kind)
}
