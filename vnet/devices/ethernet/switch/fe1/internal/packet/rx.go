// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package packet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/vnet"

	"fmt"
	"io"
	"sync/atomic"
)

type RxDescriptor struct {
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

	reason uint64

	cpu_cos_queue rx_next
	_             [3]uint8

	src_port uint8
	_        [3]uint8

	// Hi-gig module header.
	_ [4]uint32

	// [2:0] inner priority
	// [3] dvp valid (else next hop)
	// [4] packet modified needs new crc
	// [21:5] next hop index
	// [24] vfi valid
	dw10 uint32

	_ [4]uint32

	// Actual number of bytes transfered.
	rx_buffer_bytes uint16

	// [15] rx done
	// [2] rx cell error
	// [1] rx start of packet
	// [0] rx end of packet
	rx_flags rxDescFlag
}

func (r *RxDescriptor) getNextHopIndex() (nhi uint32, isECMP bool) {
	v := (r.dw10 >> 5) & 0x1ffff
	nhi = v & 0xffff
	if v&(1<<16) != 0 {
		isECMP = true
	}
	return
}

type rxDescFlag uint16

const (
	log2_rx_flag_start_of_packet            = 1
	log2_rx_flag_end_of_packet              = 0
	rx_flag_end_of_packet        rxDescFlag = 1 << log2_rx_flag_end_of_packet
	rx_flag_start_of_packet      rxDescFlag = 1 << log2_rx_flag_start_of_packet
	rx_flag_cell_error           rxDescFlag = 1 << 2
	rx_flag_done                 rxDescFlag = 1 << 15
)

var rxDescFlagStrings = [...]string{
	0:  "eop",
	1:  "sop",
	2:  "cell error",
	15: "done",
}

func (x rxDescFlag) String() string { return elib.FlagStringer(rxDescFlagStrings[:], elib.Word(x)) }

func (d *RxDescriptor) String() (s string) {
	// avoid acces to potentially dma memory
	c := *d
	s = fmt.Sprintf("0x%08x, %d/%d len/cap, flags: %s rx flags: %s, src %d, cpu-cos %d",
		c.memory_address, c.rx_buffer_bytes, c.n_buffer_bytes,
		c.flags, c.rx_flags,
		c.src_port, c.cpu_cos_queue)
	if c.reason != 0 {
		s += fmt.Sprintf(", reason 0x%x", c.reason)
	}
	return
}

//go:generate gentemplate -d Package=packet -id RxDescriptor -d Type=RxDescriptor -d VecType=RxDescriptorVec github.com/platinasystems/go/elib/hw/dma_mem.tmpl

// Size ring with enough buffers so that we can process one complete vector's worth while receiving more on the ring.
const (
	// 1 extra for halt descriptor.
	rx_ring_len_without_reload_descriptor = 1 + 8*vnet.MaxVectorLen
	// 1 extra for reload descriptor.
	rx_ring_len_with_reload_descriptor = rx_ring_len_without_reload_descriptor + 1
	nBufs                              = 2 * rx_ring_len_without_reload_descriptor
)

type rx_next uint8

const (
	Rx_next_error rx_next = iota
	Rx_next_punt
)

type rx_error uint

const (
	rx_error_none = iota
)

type rxNode struct {
	vnet.InputNode
	pool             vnet.BufferPool
	refs             [nBufs]vnet.Ref
	ports            []Port
	si_by_port       [256]vnet.Si
	rx_port_counters [256]vnet.CombinedCounter
	sequence         uint
	s                inputState
	channel          *dma_channel
	ring_index       uint
	halt_index       uint
	rx_desc          RxDescriptorVec
	desc_id          elib.Index
	active_count     int32
}

func (d *RxDescriptor) Set(n *rxNode, r *vnet.Ref) {
	*d = RxDescriptor{
		memory_address: uint32(r.DataPhys()),
		n_buffer_bytes: uint16(n.pool.Size),
		flags:          desc_next_valid | desc_scatter_gather | desc_controlled_interrupt,
	}
}

type Port struct {
	vnet.Si
	Name string
}

func (n *rxNode) rxDescPhysAddr(i uint) uint32 { return uint32(n.rx_desc[i].PhysAddress()) }
func (n *rxNode) rxDescIndex(pa uint32) uint   { return uint(pa-n.rxDescPhysAddr(0)) / 64 }

func (d *Dma) StartRx(v *vnet.Vnet, node_name string, ports []Port) {
	n := &d.rxNode

	t := &n.pool.BufferTemplate
	*t = *hw.DefaultBufferTemplate
	n.pool.Name = node_name
	v.AddBufferPool(&n.pool)

	n.Next = []string{
		Rx_next_error: "error",
		Rx_next_punt:  "punt",
	}
	n.Errors = []string{
		rx_error_none: "no error",
	}

	n.ports = ports
	for i := range n.si_by_port {
		n.si_by_port[i] = ^vnet.Si(0)
		if i < len(ports) {
			n.si_by_port[i] = ports[i].Si
		}
	}

	v.RegisterInputNode(n, node_name)
	v.RegisterSwIfCounterSyncHook(n.syncSwIfCounters)

	c := d.rx_channels[0]
	n.channel = c

	// Allocate ring descriptors and fill with buffers.
	l := uint(rx_ring_len_with_reload_descriptor)

	n.rx_desc, n.desc_id = RxDescriptorAlloc(l)
	n.rx_desc = n.rx_desc[:l]

	n.pool.AllocRefs(n.refs[:])

	// Put even buffers on ring; odd buffers will be used for refill.
	i0 := n.sequence & 1
	for i := uint(0); i < nBufs; i += 2 {
		n.rx_desc[i/2].Set(n, &n.refs[i+i0])
	}

	start := n.rxDescPhysAddr(0)

	// Last descriptor in ring is reload to descriptor 0.
	n.rx_desc[l-1] = RxDescriptor{
		memory_address: start,
		flags:          desc_reload | desc_next_valid | desc_scatter_gather,
	}

	// Accept all cos values for this channel.
	c.regs.rx_cos_control[c.index][0].Set(^uint32(0))
	c.regs.rx_cos_control[c.index][1].Set(^uint32(0))

	// Start dma engine.
	c.regs.start_descriptor_address[c.index].Set(start)
	n.halt_index = rx_ring_len_without_reload_descriptor - 1
	c.regs.halt_descriptor_address[c.index].Set(n.rxDescPhysAddr(n.halt_index))
	hw.MemoryBarrier()
	c.regs.control[c.index].Set(uint32(c.start_control))
}

func (n *rxNode) syncSwIfCounters(v *vnet.Vnet) {
	t := v.GetIfThread(0)
	for pi := range n.ports {
		p := &n.ports[pi]
		c := &n.rx_port_counters[pi]
		vnet.IfRxCounter.Add64(t, p.Si, c.Packets, c.Bytes)
		c.Zero()
	}
}

type inputState struct {
	out              *vnet.RefOut
	chain            vnet.RefChain
	next             rx_next
	n_next           uint
	last_miss_next   rx_next
	n_last_miss_next uint
}

func (n *rxNode) slowPath(r0 *vnet.Ref, f0 rxDescFlag, next0 rx_next,
	nextʹ rx_next, n_nextʹ uint) (next rx_next, n_next uint) {

	next, n_next = nextʹ, n_nextʹ
	s := &n.s

	// Append buffer to current chain.
	s.chain.Append(r0)

	// If at end of packet, enqueue packet to next graph node.
	if f0&rx_flag_end_of_packet == 0 {
		return
	}

	// Enqueue packet.
	ref := s.chain.Done()
	in := &s.out.Outs[next0]

	// Cache empty?
	if n_next == 0 {
		next = next0
	}

	// Cache hit?
	if next0 == next {
		in.Refs[n_next] = ref
		n_next++
		return
	}

	n_next0 := in.Len()
	in.SetPoolAndLen(n.Vnet, &n.pool, n_next0+1)
	in.Refs[n_next0] = ref
	n_next0++

	// Switch cached next after enough repeats of cache miss with same next.
	if next0 == s.last_miss_next {
		s.n_last_miss_next++
		if s.n_last_miss_next >= 4 {
			if n_next > 0 {
				s.out.Outs[next].SetPoolAndLen(n.Vnet, &n.pool, n_next)
			}
			next = next0
			n_next = n_next0
		}
	} else {
		s.last_miss_next = next0
		s.n_last_miss_next = 1
	}
	return
}

const (
	rx_done_not_done = iota
	rx_done_vec_len
	rx_done_found_hw_owned_descriptor
)

var rx_done_code_strings = [...]string{
	rx_done_not_done:                  "not-done",
	rx_done_vec_len:                   "vec-len",
	rx_done_found_hw_owned_descriptor: "hw-owned",
}

type rx_done_code uint8

func (c rx_done_code) String() string { return elib.Stringer(rx_done_code_strings[:], int(c)) }

func (n *rxNode) rx_no_wrap(out *vnet.RefOut, iʹ, n_doneʹ, l uint) (done rx_done_code, i, n_done uint) {
	rs, ds := n.refs[:], n.rx_desc
	i = iʹ
	n_done = n_doneʹ

	n_left := l - i
	if n_left+n_done > vnet.MaxVectorLen {
		n_left = vnet.MaxVectorLen - n_done
		done = rx_done_vec_len
	}
	n_done += n_left

	s := &n.s
	next := s.next
	n_next := s.n_next

	const fast_path_rx_flags = rx_flag_done | rx_flag_start_of_packet | rx_flag_end_of_packet

	ri0 := n.sequence & 1
	ri := 2 * i
	for n_left > 0 {
		d0 := &ds[i+0]
		f0 := d0.rx_flags

		// Never touch non cpu-owned descriptor.  e.g. don't zero rx flags below
		// Otherwise hardware race may occur.
		if f0&rx_flag_done == 0 {
			done = rx_done_found_hw_owned_descriptor
			break
		}

		src0 := d0.src_port

		// fmt.Printf("%s %d: %s\n", n.ports[src0].Name, i, d0)

		b0 := uint(d0.rx_buffer_bytes)

		// Process buffer with even index (ri ^ 0); refill with odd buffer (ri ^ 1).
		r0 := &rs[ri+(ri0^0)]
		r0.SetDataLen(b0)

		r0.Si = n.si_by_port[src0]
		next0 := d0.cpu_cos_queue

		// Refill with new buffer (odd index).
		d0.memory_address = uint32(rs[ri+(ri0^1)].DataPhys())

		// Clear done flag.
		d0.rx_flags = 0

		// Increment packet and byte count for src port.
		n.rx_port_counters[src0].Packets += uint64((f0 >> log2_rx_flag_start_of_packet) & 1)
		n.rx_port_counters[src0].Bytes += uint64(b0)

		n.SetError(r0, rx_error_none)

		out.Outs[next].Refs[n_next] = *r0
		n_next += 1
		i += 1
		n_left -= 1
		ri += 2 * 1

		// Slow path.
		if !(f0 == fast_path_rx_flags && next0 == next) {
			next, n_next = n.slowPath(r0, f0, next0, next, n_next-1)
		}
	}

	n_done -= n_left

	// Cache next index for next call.
	s.next = next
	s.n_next = n_next

	if elog.Enabled() && n_done > 0 {
		elog.GenEventf("fe1 rx head %d -> %d done %d %s", iʹ, i, n_done, done)
	}

	return
}

func (n *rxNode) DumpRxRing(w io.Writer, tag string) {
	c := n.channel
	var v [3]uint32
	v[0] = c.regs.halt_status.Get()
	v[1] = c.regs.status.Get()
	v[2] = c.regs.current_descriptor_address[c.index].Get()
	fmt.Fprintf(w, "%s: index %d, halt %d, halt status %x status %x cur %d\n", tag, n.ring_index, n.halt_index, v[0], v[1], n.rxDescIndex(v[2]))
	for i := uint(0); i < rx_ring_len_with_reload_descriptor; i++ {
		c := ':'
		if i == n.ring_index {
			c = '{'
		}
		if i == n.halt_index {
			c = '}'
		}
		fmt.Fprintf(w, "%2d%c 0x%x %s\n", i, c, n.rx_desc[i].PhysAddress(), &n.rx_desc[i])
	}
}

func (n *rxNode) InterruptEnable(enable bool) {
	d := n.channel.dma
	d.interruptsEnabled = enable
	d.InterruptEnable(enable)
}

func (n *rxNode) NodeInput(out *vnet.RefOut) {
	var done rx_done_code

	n.s.out = out
	n.s.n_next = 0
	n.s.last_miss_next = ^rx_next(0) // invalidate

	var i, iʹ uint
	iʹ = n.ring_index
	n_done := uint(0)
	done, i, n_done = n.rx_no_wrap(out, iʹ, n_done, rx_ring_len_without_reload_descriptor)
	if i >= rx_ring_len_without_reload_descriptor {
		i = 0

		// Allocate new re-fill buffers when ring wraps.
		ri0 := n.sequence & 1
		n.sequence++
		n.pool.AllocRefsStride(&n.refs[ri0], rx_ring_len_without_reload_descriptor, 2)
	}

	if done == rx_done_not_done && iʹ > 0 {
		done, i, n_done = n.rx_no_wrap(out, 0, n_done, iʹ)
	}
	n.ring_index = i

	// Set halt address to be ahead of current index by ring length less one for halt descriptor itself.
	c := n.channel
	n.halt_index = i + rx_ring_len_without_reload_descriptor - 1
	if n.halt_index >= rx_ring_len_without_reload_descriptor {
		n.halt_index -= rx_ring_len_without_reload_descriptor
	}
	c.regs.halt_descriptor_address[c.index].Set(n.rxDescPhysAddr(n.halt_index))

	if elog.Enabled() {
		elog.GenEventf("fe1 rx index %d, halt %d", n.ring_index, n.halt_index)
	}

	if n.s.n_next > 0 {
		out.Outs[n.s.next].SetPoolAndLen(n.Vnet, &n.pool, n.s.n_next)
	}

	// If interrupts are disabled poll clean tx ring.
	if d := n.channel.dma; !d.interruptsEnabled {
		if d.txNode.channel != nil {
			d.txNode.DescDoneInterrupt()
		}
	} else if done == rx_done_found_hw_owned_descriptor && n_done == 0 {
		if n.active_count <= 0 {
			panic("ga")
		}
		n.Activate(atomic.AddInt32(&n.active_count, -1) > 0)
	}

	n.channel.ack_desc_controlled_interrupt()
}
