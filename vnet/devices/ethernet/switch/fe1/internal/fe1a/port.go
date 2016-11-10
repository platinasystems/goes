// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/tsc"

	"fmt"
)

// Global physical port (GPP) number (used in multiple module stack).
// Used to match in feature lookup.
type global_physical_port_number uint16

const (
	n_global_physical_port = 256
)

// Port number as used by rx/tx pipes.  Also known as "device port number".
type pipe_port_number uint16

const (
	n_pipe_ports      = 136
	pipe_port_invalid = 0xffff

	// Pseudo-port (index) used for cpu hi gig.
	pipe_port_cpu_hi_gig_index pipe_port_number = 136
)

var pipe_port_number_by_phys [n_phys_ports]pipe_port_number
var phys_by_pipe_port_number [n_pipe_ports]physical_port_number

func (p physical_port_number) toPipe() pipe_port_number {
	return pipe_port_number_by_phys[p]
}

func (p pipe_port_number) toPhys() physical_port_number {
	return phys_by_pipe_port_number[p]
}

func init() {
	t := &pipe_port_number_by_phys
	for i := range t {
		t[i] = pipe_port_invalid
	}

	t[phys_port_cpu] = 0
	t[phys_port_loopback_for_pipe(0)] = 33
	t[phys_port_loopback_for_pipe(1)] = 67
	t[phys_port_loopback_for_pipe(2)] = 101
	t[phys_port_loopback_for_pipe(3)] = 135 // can also be 134
	t[phys_port_management_0] = 66
	t[phys_port_management_1] = 100

	for i := 0; i < n_idb_data_port; i++ {
		// Phys 1-31 clport0-7 => pipe ports 1-31
		t[1+n_idb_data_port*0+i] = pipe_port_number(i) + first_data_port_for_pipe[0]
		// Phys 33-64 clport8-15 => pipe ports 34-65
		t[1+n_idb_data_port*1+i] = pipe_port_number(i) + first_data_port_for_pipe[1]
		// Phys 65-96 clport16-23 => pipe ports 68-99
		t[1+n_idb_data_port*2+i] = pipe_port_number(i) + first_data_port_for_pipe[2]
		// Phys 97-128 clport24-31 => pipe ports 102-133
		t[1+n_idb_data_port*3+i] = pipe_port_number(i) + first_data_port_for_pipe[3]
	}

	for i := range phys_by_pipe_port_number {
		phys_by_pipe_port_number[i] = phys_port_invalid
	}
	for i := range pipe_port_number_by_phys {
		if pipe := pipe_port_number_by_phys[i]; pipe != pipe_port_invalid {
			phys_by_pipe_port_number[pipe] = physical_port_number(i)
		}
	}
}

var first_data_port_for_pipe = [n_pipe]pipe_port_number{1, 34, 68, 102}

func (mod34 pipe_port_number) mod34_to_pipe(pipe uint) (p pipe_port_number) {
	if mod34 >= 34 {
		panic("34")
	}
	p = first_data_port_for_pipe[pipe] + mod34

	// Pipe 0 ports are numbered from 1-31 since cpu port is always 0.
	// So mod34 == 1 and pipe == 0 => pipe port 1.
	if pipe == 0 {
		if mod34 == 0 {
			panic("0")
		}
		p--
	}
	return
}

// Physical port numbers.
type physical_port_number uint16

const (
	n_phys_data_ports       = 32 * tsc.N_lane // each serdes lane can be a physical port.
	n_phys_management_ports = 2
)

const (
	// punt path to cpu via PCI bus
	phys_port_cpu physical_port_number = iota

	// 128 physical data ports numbered 1 through 128
	phys_port_data_lo physical_port_number = 1
	phys_port_data_hi physical_port_number = phys_port_data_lo + n_data_ports - 1
	n_data_ports                           = 128

	// 2 management ports
	phys_port_management_0 physical_port_number = 129
	phys_port_management_1 physical_port_number = 131

	// 132 ... 135 per-pipe loopback ports
	phys_port_loopback_pipe_0 physical_port_number = 132

	phys_port_invalid physical_port_number = 0xffff

	n_phys_ports = 136
)

func (p physical_port_number) is_loopback_port() bool {
	return p >= phys_port_loopback_pipe_0 && p < phys_port_loopback_pipe_0+n_pipe
}
func phys_port_loopback_for_pipe(i uint) physical_port_number {
	return phys_port_loopback_pipe_0 + physical_port_number(i)
}

func (p physical_port_number) is_data_port_in_range(lo, n_ports physical_port_number) bool {
	return p >= phys_port_data_lo+lo && p < phys_port_data_lo+lo+n_ports
}

func (p physical_port_number) speedBitsPerSec(t *fe1a) float64 {
	if p.is_loopback_port() {
		return 10e9
	}
	switch p {
	case phys_port_cpu, phys_port_management_0, phys_port_management_1:
		return 10e9
	default:
		if x := t.port_by_phys_port[p]; x != nil {
			return x.SpeedBitsPerSec
		}
		return 0
	}
}

type idb_mmu_port_number uint16

const (
	// 32 data ports per idb/mmu
	idb_mmu_port_data_0 idb_mmu_port_number = iota
	// idb port 32 is cpu/management ports depending on pipe number
	idb_mmu_port_pipe_0_cpu                 idb_mmu_port_number = 32
	idb_mmu_port_pipe_1_management_0        idb_mmu_port_number = 32
	idb_mmu_port_pipe_2_management_1        idb_mmu_port_number = 32
	idb_mmu_port_any_pipe_cpu_or_management idb_mmu_port_number = 32
	// loopback port is number 33 in all 4 pipes.
	idb_mmu_port_loopback idb_mmu_port_number = 33
	n_idb_mmu_port                            = 34

	// Aliases
	n_idb_data_port = 32
	n_mmu_data_port = 32
	n_idb_port      = n_idb_mmu_port
	n_mmu_port      = n_idb_mmu_port

	// Pseudo-ports appearing in tdm calendars.
	// Over subscription round robin.
	idb_mmu_port_over_subscription = n_idb_mmu_port + 0
	// Null psuedo-port: no port pick, no opportunistic.
	idb_mmu_port_null = n_idb_mmu_port + 1
	// Pseudo-port for mem reset, l2 management, etc.
	idb_mmu_port_idle1 = n_idb_mmu_port + 2
	// Port for guaranteed for refresh
	idb_mmu_port_idle2 = n_idb_mmu_port + 3
	// Port number used to tell hardware that table entries are not valid.
	idb_mmu_port_invalid = 0x3f
)

// 8 bit global mmu port number: 2 bit pipe + 6 bit mmu port number
type mmu_global_port_number uint16
type rx_tx_pipe uint8 // rx/tx pipe index [0,3]
type mmu_pipe uint8   // mmu pipe (xpe) index [0,3]

const mmu_global_port_number_invalid = 0xffff

var phys_by_idb_and_pipe [n_idb_port][n_pipe]physical_port_number

func init() {
	for i := 0; i < n_idb_data_port; i++ {
		for p := 0; p < n_pipe; p++ {
			phys_by_idb_and_pipe[i][p] = phys_port_data_lo + physical_port_number(n_idb_data_port*p+i)
		}
	}
	// Loopback ports for each pipe.
	for p := 0; p < n_pipe; p++ {
		phys_by_idb_and_pipe[idb_mmu_port_loopback][p] = phys_port_loopback_pipe_0 + physical_port_number(p)
	}
	// Cpu and management ports.
	phys_by_idb_and_pipe[idb_mmu_port_pipe_0_cpu][0] = phys_port_cpu
	phys_by_idb_and_pipe[idb_mmu_port_pipe_1_management_0][1] = phys_port_management_0
	phys_by_idb_and_pipe[idb_mmu_port_pipe_2_management_1][2] = phys_port_management_1
	// Invalidate unused slots
	phys_by_idb_and_pipe[idb_mmu_port_pipe_0_cpu][3] = phys_port_invalid
}

func (i idb_mmu_port_number) toPhys(pipe uint) physical_port_number {
	return phys_by_idb_and_pipe[i][pipe]
}

var (
	idb_by_phys     [n_phys_ports]idb_mmu_port_number
	ie_pipe_by_phys [n_phys_ports]rx_tx_pipe
)

func init() {
	for p := physical_port_number(0); p < n_phys_ports; p++ {
		if i := p - phys_port_data_lo; i < n_data_ports {
			idb_by_phys[p] = idb_mmu_port_number(i % 32)
			ie_pipe_by_phys[p] = rx_tx_pipe(i / 32)
		} else if i := p - phys_port_loopback_pipe_0; i < n_pipe {
			idb_by_phys[p] = idb_mmu_port_loopback
			ie_pipe_by_phys[p] = rx_tx_pipe(i)
		} else {
			switch p {
			case phys_port_cpu:
				idb_by_phys[p] = idb_mmu_port_pipe_0_cpu
				ie_pipe_by_phys[p] = 0
			case phys_port_management_0:
				idb_by_phys[p] = idb_mmu_port_pipe_1_management_0
				ie_pipe_by_phys[p] = 1
			case phys_port_management_1:
				idb_by_phys[p] = idb_mmu_port_pipe_2_management_1
				ie_pipe_by_phys[p] = 2
			default:
				idb_by_phys[p] = idb_mmu_port_invalid
				ie_pipe_by_phys[p] = ^rx_tx_pipe(0)
			}
		}
	}
}

func (t *fe1a) set_mmu_pipe_map() {
	c := t.GetSwitchCommon()
	cf := &c.SwitchConfig

	// Allow configuration to specify which port blocks belong to which mmu pipes (xpes).
	for p := range t.mmu_pipe_by_phys_port {
		t.mmu_pipe_by_phys_port[p] = mmu_pipe(ie_pipe_by_phys[p])
	}
	for i := range cf.MMUPipeByPortBlock {
		for j := 0; j < 4; j++ {
			p := phys_port_data_lo + physical_port_number(4*i+j)
			if got, want := int(cf.MMUPipeByPortBlock[i]), n_pipe; got >= want {
				panic(fmt.Errorf("mmu pipe index out of rante %d >= %d for block %d", got, want, i))
			}
			t.mmu_pipe_by_phys_port[p] = mmu_pipe(cf.MMUPipeByPortBlock[i])
		}
	}
}

func (i idb_mmu_port_number) gratuitous_mmu_port_scrambling() idb_mmu_port_number {
	if i < n_idb_data_port {
		i = i>>1 | (i&1)<<4
	}
	return i
}

func (p physical_port_number) toIdb() (i idb_mmu_port_number, pipe rx_tx_pipe) {
	i, pipe = idb_by_phys[p], ie_pipe_by_phys[p]
	return
}

func (p physical_port_number) pipe() rx_tx_pipe { return ie_pipe_by_phys[p] }

func (p physical_port_number) toMmu(t *fe1a) (i idb_mmu_port_number, m mmu_pipe) {
	i, m = idb_by_phys[p], t.mmu_pipe_by_phys_port[p]
	i = i.gratuitous_mmu_port_scrambling()
	return
}

func (p physical_port_number) toGlobalMmu(t *fe1a) mmu_global_port_number {
	i, pipe := p.toMmu(t)
	if pipe != ^mmu_pipe(0) {
		return mmu_global_port_number(i) | mmu_global_port_number(pipe)<<6
	} else {
		return mmu_global_port_number_invalid
	}
}

type port_speed uint8

const (
	// Port speed codes: 20G or less, etc.
	port_speed_lt_20g = iota + 1
	port_speed_lt_25g
	port_speed_lt_40g
	port_speed_lt_50g
	port_speed_lt_100g
	port_speed_ge_100g // >= 100G
	port_speed_first   = 1
)

const n_port_speed = 7

func port_speed_code(bps float64) port_speed {
	switch {
	case bps < 20e9:
		return port_speed_lt_20g
	case bps < 25e9:
		return port_speed_lt_25g
	case bps < 40e9:
		return port_speed_lt_40g
	case bps < 50e9:
		return port_speed_lt_50g
	case bps < 100e9:
		return port_speed_lt_100g
	default:
		return port_speed_ge_100g
	}
}

type port_bitmap [5]uint32

type port_bitmap_entry struct{ port_bitmap }

func make_port_bitmap_entry(p port_bitmap) port_bitmap_entry { return port_bitmap_entry{port_bitmap: p} }

func (r *port_bitmap_entry) MemBits() int { return int(n_phys_ports) }
func (r *port_bitmap_entry) MemGetSet(b []uint32, isSet bool) {
	if isSet {
		copy(b, r.port_bitmap[:])
	} else {
		copy(r.port_bitmap[:], b)
	}
}

func (p *port_bitmap) MemGetSet(b []uint32, i int, isSet bool) int {
	l := len(p)
	for j := 0; j < l-1; j++ {
		i = m.MemGetSetUint32(&p[j], b, i+31, i, isSet)
	}
	i = m.MemGetSetUint32(&p[l-1], b, i+(n_phys_ports%32)-1, i, isSet)
	return i
}

func (b *port_bitmap) isSet(p pipe_port_number) (ok bool) {
	if p != pipe_port_invalid {
		i0, i1 := p/32, p%32
		ok = b[i0]&(1<<i1) != 0
	}
	return
}

func (b *port_bitmap) add(p pipe_port_number) {
	i0, i1 := p/32, p%32
	b[i0] |= 1 << i1
}

func (b *port_bitmap) del(p pipe_port_number) {
	i0, i1 := p/32, p%32
	b[i0] &^= 1 << i1
}

func (b *port_bitmap) or(c *port_bitmap) {
	for i := range b {
		b[i] |= c[i]
	}
}

func (b *port_bitmap) and(c *port_bitmap) {
	for i := range b {
		b[i] &= c[i]
	}
}

func (b *port_bitmap) andNot(c *port_bitmap) {
	for i := range b {
		b[i] &^= c[i]
	}
}

type port_bitmap_mem m.MemElt

func (r *port_bitmap_mem) geta(q *DmaRequest, v *port_bitmap_entry, b sbus.Block, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, b, t)
}
func (r *port_bitmap_mem) seta(q *DmaRequest, v *port_bitmap_entry, b sbus.Block, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, b, t)
}
