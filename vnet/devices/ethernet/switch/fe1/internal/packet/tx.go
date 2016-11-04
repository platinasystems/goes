// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package packet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"

	"fmt"
	"sync"
)

type TxDescriptor struct {
	memory_address uint32

	// Tx: number of bytes to send.
	// Rx: number of bytes that can be received in this buffer.
	n_buffer_bytes uint16

	// [0] chain 1 => continue and process next descriptor; 0 => stop dma processing, end of chain
	// [1] 0 => end-of-packet, 1 => not end-of-packet (scatter/gather)
	// [2] reload 1 => memory_address contains address of next descriptor
	// [3] higig/sobmh module header valid; if not valid packet is an ethernet packet.
	// [4] 1 => tx stats update
	// [5] tx pause
	// [6] tx purge
	// [7] all desc interrupt mode
	// [8] controlled interrupt mode
	flags descFlag

	// [0] [31:30] start of frame: 2 => internal packet, 3 => higig header
	//     [29:24] header type
	//     [15:0] lm counter index ?
	// [1] [7:0] local dst port for epipe to send packet out
	//           if set to loopback port then packet must have loopback header (before ethernet header).
	// [2] [31:28] priority
	//     [27:24] cos
	//     [18] unicast
	//     [23:18] logical egress port for packets from epipe
	//     [17:8] queue number
	//     [7:0] src module id (must be set to my module id register)
	// [3] written by hardware on rx.  not used on tx.
	raw_module_header

	_ [10]uint32
}

func (d *TxDescriptor) String() (s string) {
	s = fmt.Sprintf("0x%08x, %d bytes, %v", d.memory_address, d.n_buffer_bytes, d.flags)
	if d.flags&desc_module_header_valid != 0 {
		s += fmt.Sprintf(", module header %x", &d.raw_module_header)
	}
	return
}

//go:generate gentemplate -d Package=packet -id TxDescriptor -d Type=TxDescriptor -d VecType=TxDescriptorVec github.com/platinasystems/go/elib/hw/dma_mem.tmpl

type InterfaceNode struct {
	// Packet interfaces are output only.  Input is via rx node.
	vnet.OutputInterfaceNode
	dma                             *Dma
	tx_module_header                raw_module_header
	tx_loopback_header              raw_loopback_header
	loopback_header_valid           bool
	tx_desc_flag_by_next_valid_flag [vnet.NextValid + 1]descFlag
}

type InterfaceNodeConfig struct {
	Src_port, Dst_port   uint8
	Loopback_port        uint8
	Use_module_header    bool
	Use_loopback_header  bool
	Is_visibility_packet bool
}

func (n *InterfaceNode) Init(dma *Dma, hi vnet.Hi, cf *InterfaceNodeConfig) {
	n.SetHi(hi)
	n.dma = dma

	module_header_valid := cf.Use_module_header || cf.Use_loopback_header

	mh_dst_port := cf.Dst_port
	if cf.Use_loopback_header {
		mh_dst_port = cf.Loopback_port
	}
	h := tx_module_header{
		tx_module_header_type: tx_module_header_type_from_cpu,
		dst_port:              mh_dst_port,
	}
	n.tx_module_header.encodeTx(&h)

	{
		h := tx_loopback_header{
			loopback_header_type: loopback_header_ethernet,
			is_visibility_packet: cf.Is_visibility_packet,
			src_port:             uint16(cf.Src_port),
			dst_port:             cf.Dst_port,
		}
		n.loopback_header_valid = cf.Use_loopback_header
		n.tx_loopback_header.encode(&h)
	}

	{
		base := desc_next_valid | desc_tx_stats_update
		if module_header_valid {
			base |= desc_module_header_valid
		}
		n.tx_desc_flag_by_next_valid_flag[0] = base | desc_controlled_interrupt
		n.tx_desc_flag_by_next_valid_flag[vnet.NextValid] = base | desc_scatter_gather
	}
}

type tx_in struct {
	in   *vnet.RefVecIn
	node *InterfaceNode
	free chan *vnet.RefVecIn
}

func (x *tx_in) Len() uint { return x.in.Refs.Len() }

type txNode struct {
	channel *dma_channel

	mu sync.Mutex

	// Tx descriptor ring and heap identifier.
	tx_desc TxDescriptorVec
	desc_id elib.Index

	// Ring head (ring_index) and tail (halt_index).
	ring_index uint
	halt_index uint

	txRing vnet.TxDmaRing
}

const tx_ring_len = 4*vnet.MaxVectorLen + 1

func (t *txNode) txDescPhysAddr(i uint) uint32 { return uint32(t.tx_desc[i].PhysAddress()) }
func (t *txNode) txDescIndex(pa uint32) uint   { return uint(pa-t.txDescPhysAddr(0)) / 64 }

func (t *txNode) start(v *vnet.Vnet) {
	c := t.channel

	t.txRing.Init(v)

	l := uint(tx_ring_len + 1) // 1 extra for reload descriptor.
	t.tx_desc, t.desc_id = TxDescriptorAlloc(l)
	t.tx_desc = t.tx_desc[:l]

	// Zero descriptors for hygiene.
	for i := range t.tx_desc {
		t.tx_desc[i] = TxDescriptor{}
	}

	// Last descriptor in ring is reload to descriptor 0.
	start := t.txDescPhysAddr(0)
	t.tx_desc[l-1] = TxDescriptor{
		memory_address: start,
		flags:          desc_reload | desc_next_valid | desc_scatter_gather | desc_controlled_interrupt,
	}

	// Start dma engine; initially halted on first descriptor.
	c.regs.start_descriptor_address[c.index].Set(start)
	t.halt_index = 0
	c.regs.halt_descriptor_address[c.index].Set(t.txDescPhysAddr(t.halt_index))
	hw.MemoryBarrier()
	c.regs.control[c.index].Set(uint32(c.start_control))
}

func (d0 *TxDescriptor) adjustModuleheader(r0 *vnet.Ref, nextFlag0 vnet.BufferFlag, is_sopʹ bool) (is_sop bool) {
	is_eop := nextFlag0 == 0
	if is_sopʹ {
		h0 := (*ethernet.Header)(r0.Data())
		d0.raw_module_header.setUnicast(h0.Dst.IsUnicast())
		// Hardware requires padding of runts.
		if is_eop && d0.n_buffer_bytes < 64 {
			d0.n_buffer_bytes = 64
		}
	}
	is_sop = is_eop
	return
}

func (node *InterfaceNode) set1Desc(rs []vnet.Ref, ds []TxDescriptor, is_sopʹ bool, ri, di uint) (is_sop bool) {
	is_sop = is_sopʹ
	r0, d0 := &rs[ri+0], &ds[di+0]
	d0.memory_address = uint32(r0.DataPhys())
	d0.n_buffer_bytes = uint16(r0.DataLen())
	d0.raw_module_header = node.tx_module_header
	f0 := r0.NextValidFlag()
	d0.flags = node.tx_desc_flag_by_next_valid_flag[f0]
	is_sop = d0.adjustModuleheader(r0, f0, is_sop)
	return
}

func (node *InterfaceNode) set4Desc(rs []vnet.Ref, ds []TxDescriptor, is_sopʹ bool, ri, di uint) (is_sop bool) {
	is_sop = is_sopʹ
	r0, r1, r2, r3 := &rs[ri+0], &rs[ri+1], &rs[ri+2], &rs[ri+3]
	d0, d1, d2, d3 := &ds[di+0], &ds[di+1], &ds[di+2], &ds[di+3]

	d0.memory_address = uint32(r0.DataPhys())
	d1.memory_address = uint32(r1.DataPhys())
	d2.memory_address = uint32(r2.DataPhys())
	d3.memory_address = uint32(r3.DataPhys())

	d0.n_buffer_bytes = uint16(r0.DataLen())
	d1.n_buffer_bytes = uint16(r1.DataLen())
	d2.n_buffer_bytes = uint16(r2.DataLen())
	d3.n_buffer_bytes = uint16(r3.DataLen())

	d0.raw_module_header = node.tx_module_header
	d1.raw_module_header = node.tx_module_header
	d2.raw_module_header = node.tx_module_header
	d3.raw_module_header = node.tx_module_header

	f0 := r0.NextValidFlag()
	f1 := r1.NextValidFlag()
	f2 := r2.NextValidFlag()
	f3 := r3.NextValidFlag()

	d0.flags = desc_next_valid | node.tx_desc_flag_by_next_valid_flag[f0]
	d1.flags = desc_next_valid | node.tx_desc_flag_by_next_valid_flag[f1]
	d2.flags = desc_next_valid | node.tx_desc_flag_by_next_valid_flag[f2]
	d3.flags = desc_next_valid | node.tx_desc_flag_by_next_valid_flag[f3]

	is_sop = d0.adjustModuleheader(r0, f0, is_sop)
	is_sop = d1.adjustModuleheader(r1, f1, is_sop)
	is_sop = d2.adjustModuleheader(r2, f2, is_sop)
	is_sop = d3.adjustModuleheader(r3, f3, is_sop)

	return
}

func (node *InterfaceNode) setDesc(rs []vnet.Ref, ds []TxDescriptor, is_sopʹ bool,
	ri0, di0, rn, dn uint) (is_sop bool, ri, di, nd uint) {
	ri, di = ri0, di0
	is_sop = is_sopʹ
	for ri+4 <= rn && di+4 <= dn {
		is_sop = node.set4Desc(rs, ds, is_sop, ri, di)
		ri += 4
		di += 4
	}
	for ri < rn && di < dn {
		is_sop = node.set1Desc(rs, ds, is_sop, ri, di)
		ri += 1
		di += 1
	}
	nd = ri - ri0
	return
}

func (node *InterfaceNode) output(t *txNode, in *vnet.TxRefVecIn) {
	nr := in.Len()

	head, tail := t.ring_index, t.halt_index
	// Free slots are after tail and before head.
	n_free := head - tail
	if int(n_free) <= 0 {
		n_free += tx_ring_len
	}
	// Leave empty slot for halt index.
	n_free -= 1

	// No room?
	if n_free < nr {
		panic("ga")
		node.Vnet.FreeTxRefIn(in)
		return
	}

	ds, rs := t.tx_desc, in.Refs

	ri := uint(0)
	n_tx := uint(0)

	// From tail (halt index) to end of ring.
	di := tail
	n_end := n_free
	if tail+n_end > tx_ring_len {
		n_end = tx_ring_len - tail
	}
	is_sop := true
	if n_end > 0 {
		var nd uint
		is_sop, ri, di, nd = node.setDesc(rs, ds, is_sop, ri, di, nr, di+n_end)
		n_free -= nd
		n_tx += nd
	}

	// From start of ring to head.
	n_start := n_free
	if n_start > head {
		n_start = head
	}
	if n_start > 0 && ri < nr {
		var nd uint
		is_sop, ri, di, nd = node.setDesc(rs, ds, is_sop, ri, 0, nr, n_start)
		n_free -= nd
		n_tx += nd
	}

	// Ring wrap.
	if di >= tx_ring_len {
		di = 0
	}

	if elog.Enabled() && n_tx > 0 {
		elog.GenEventf("fe1 tx %d halt %d head %d tail %d", n_tx, di, head, tail)
	}

	hw.MemoryBarrier()

	// Re-start dma engine when tail advances.
	if di != t.halt_index {
		c := t.channel
		t.halt_index = di
		c.regs.halt_descriptor_address[c.index].Set(t.txDescPhysAddr(di))
	}

	t.txRing.ToInterrupt <- in
}

func (node *InterfaceNode) InterfaceOutput(in *vnet.TxRefVecIn) {
	t := &node.dma.txNode

	if node.loopback_header_valid {
		node.addLoopbackHeader(in)
	}

	// Start on first call.
	if t.channel == nil {
		t.channel = node.dma.tx_channels[0]
		t.start(node.Vnet)
	}

	node.output(t, in)
}

func (t *txNode) DescDoneInterrupt() {
	// Mutually excludes real interrupt and polled calls from interfering with each other.
	t.mu.Lock()
	defer t.mu.Unlock()

	c := t.channel
	di := t.txDescIndex(c.regs.current_descriptor_address[c.index].Get())
	cur := di
	if di == tx_ring_len {
		di = tx_ring_len - 1
	}

	n_advance := di - t.ring_index
	if di < t.ring_index {
		n_advance += tx_ring_len
	}

	if elog.Enabled() {
		v := c.regs.halt_status.Get()
		elog.GenEventf("fe1 tx adv %d halt %d/%d head %d/%d %d", n_advance, t.halt_index, 1&(v>>(27+c.index)), di, cur, t.ring_index)
	}

	t.ring_index = di
	t.txRing.InterruptAdvance(n_advance)

	c.ack_desc_controlled_interrupt()
}

func (t *txNode) DumpTxRing(tag string) {
	c := t.channel
	var w [3]uint32
	w[0] = c.regs.halt_status.Get()
	w[1] = c.regs.status.Get()
	w[2] = c.regs.current_descriptor_address[c.index].Get()
	fmt.Printf("%s: index %d, halt %d, halt %d status %x cur %d\n", tag, t.ring_index, t.halt_index,
		1&(w[0]>>(27+c.index)), w[1], t.txDescIndex(w[2]))
	for i := uint(0); i <= tx_ring_len; i++ {
		c := ':'
		if i == t.ring_index {
			c = '{'
		}
		if i == t.halt_index {
			c = '}'
		}
		fmt.Printf("%2d%c 0x%x %s\n", i, c, t.tx_desc[i].PhysAddress(), &t.tx_desc[i])
	}
}

type tx_module_header_type uint8

const (
	tx_module_header_type_internal tx_module_header_type = (2 << 6)
	// Found in packets received from epipe.
	tx_module_header_type_from_epipe tx_module_header_type = tx_module_header_type_internal | (0 << 0)
	// Used for packets sent from cpu.
	tx_module_header_type_from_cpu tx_module_header_type = tx_module_header_type_internal | (1 << 0)
)

type tx_module_header struct {
	tx_module_header_type

	// Physical port number to send packet out.
	dst_port uint8

	priority uint8
	cos      uint8

	rqe_queue uint8

	service_pool uint8

	service_pool_override bool
}

type raw_module_header [16]uint8

func (m *raw_module_header) encodeTx(t *tx_module_header) {
	m[3] = uint8(t.tx_module_header_type)
	m[4] = t.dst_port
	if t.service_pool_override {
		m[11] |= 1 << 0
		m[10] |= uint8(t.service_pool) << 6
	}
	m[11] |= t.priority << 1
	m[9] |= t.cos & 0x3f
	m[10] |= uint8(t.rqe_queue & 0xf)
}

func (m *raw_module_header) setUnicast(isUnicast bool) {
	if isUnicast {
		// set unicast bit.
		m[9] |= 1 << 6
	} else {
		// set l2 bitmap select for multicast.
		m[9] |= 1 << 7
	}
}

type loopback_header_type uint8

const (
	loopback_header_tunnel   loopback_header_type = 0
	loopback_header_ethernet loopback_header_type = 3
)

type tx_loopback_header struct {
	is_visibility_packet bool

	loopback_header_type
	input_priority uint8

	is_src_virtual_port_valid bool
	src_port                  uint16

	// 3 bit index into ipipe cpu_pkt_profile registers.
	visibility_packet_profile_index uint8

	dst_port uint8
}

func (node *InterfaceNode) addLoopbackHeader(in *vnet.TxRefVecIn) {
	rs := in.Refs
	n := rs.Len()
	i := uint(0)
	for i+4 <= n && false {
		r0, r1, r2, r3 := &rs[i+0], &rs[i+1], &rs[i+2], &rs[i+3]

		r0.Advance(-loopback_header_bytes)
		r1.Advance(-loopback_header_bytes)
		r2.Advance(-loopback_header_bytes)
		r3.Advance(-loopback_header_bytes)

		h0 := (*raw_loopback_header)(r0.Data())
		h1 := (*raw_loopback_header)(r1.Data())
		h2 := (*raw_loopback_header)(r2.Data())
		h3 := (*raw_loopback_header)(r3.Data())

		*h0 = node.tx_loopback_header
		*h1 = node.tx_loopback_header
		*h2 = node.tx_loopback_header
		*h3 = node.tx_loopback_header
		i += 4
	}

	for i < n {
		r0 := &rs[i+0]
		r0.Advance(-loopback_header_bytes)
		h0 := (*raw_loopback_header)(r0.Data())
		*h0 = node.tx_loopback_header
		i += 1
	}
}

const loopback_header_bytes = 16

type raw_loopback_header [loopback_header_bytes]byte

func (r *raw_loopback_header) encode(h *tx_loopback_header) {
	r[0] = 0xfb
	r[1] = h.input_priority<<4 | byte(h.loopback_header_type)
	r[2] = 1<<6 | byte(h.src_port&0x3f)
	r[3] = byte(h.src_port >> 6)
	r[4] = byte(h.src_port>>14) << 6
	if h.is_visibility_packet {
		r[4] |= 1<<5 | byte(h.visibility_packet_profile_index)<<2
	}
	r[15] = h.dst_port
}
