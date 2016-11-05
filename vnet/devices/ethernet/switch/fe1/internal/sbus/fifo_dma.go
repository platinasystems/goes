// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sbus

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/hw"

	"fmt"
	"sync"
)

const n_fifo_dma_channels = 4

type FifoDmaRegs struct {
	control [n_fifo_dma_channels]hw.Reg32

	sbus_start_address     [n_fifo_dma_channels]hw.Reg32
	host_mem_start_address [n_fifo_dma_channels]hw.Reg32

	host_mem [n_fifo_dma_channels]struct{ n_entries_read, n_entries_valid hw.Reg32 }

	ecc_error_address       [n_fifo_dma_channels]hw.Reg32
	ecc_error_control       [n_fifo_dma_channels]hw.Reg32
	host_mem_write_pointers [n_fifo_dma_channels]hw.Reg32

	_ [0x354 - 0x340]byte

	host_mem_interrupt_threshold [n_fifo_dma_channels]hw.Reg32

	status [n_fifo_dma_channels]fifo_dma_status

	status_clear [n_fifo_dma_channels]hw.Reg32

	sbus_opcode [n_fifo_dma_channels]command_reg

	debug hw.Reg32

	_ [0x3a0 - 0x398]byte
}

type fifo_dma_channel struct {
	index        uint
	sequence     uint
	nWords       uint
	log2MemElts  uint
	regs         *FifoDmaRegs
	data         dma_data_vec
	data_heap_id elib.Index
	free         chan elib.Uint32Vec
	resultFifo   chan FifoDmaData
	mu           sync.Mutex
}

type FifoDma struct {
	n_channels_used uint
	Channels        [n_fifo_dma_channels]fifo_dma_channel
}

type FifoDmaData struct {
	Data elib.Uint32Vec
	free chan elib.Uint32Vec
}

func (d *FifoDmaData) Free() { d.free <- d.Data }

type fifo_dma_status hw.Reg32

func (r *fifo_dma_status) get() (v fifo_dma_status) {
	v = fifo_dma_status((*hw.Reg32)(r).Get())
	return
}

func (c *fifo_dma_channel) Interrupt() {
	status := c.regs.status[c.index].get()
	if status&1 != 0 {
		c.regs.status_clear[c.index].Set(0x7)
		panic(fmt.Errorf("fifo dma channel %d status %x", c.index, status))
	}
	if status&0xff8 != 0 {
		panic(fmt.Errorf("fifo dma channel %d unexpected status 0x%x", c.index, status))
	}
	if status&4 != 0 {
		c.regs.status_clear[c.index].Set(1)
		c.resultFifo <- c.poll()
	}
}

func (c *fifo_dma_channel) get_buf() (v elib.Uint32Vec) {
	select {
	case v = <-c.free:
		v = v[:0]
	default:
	}
	return
}

func (c *fifo_dma_channel) start_address() uint32 { return uint32(c.data[0].PhysAddress()) }
func (c *fifo_dma_channel) write_index() uint {
	v := c.regs.host_mem_write_pointers[c.index].Get()
	return uint(v-c.start_address()) / (4 * c.nWords)
}

func (c *fifo_dma_channel) poll() (r FifoDmaData) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Circular arithmetic.
	mask := (uint(1) << c.log2MemElts) - 1
	head := c.sequence & mask
	tail := c.write_index()
	n := (tail - head) & mask
	if n == 0 {
		return
	}

	// Convert to words.
	hw, tw, nw := head*c.nWords, tail*c.nWords, n*c.nWords

	buf := c.get_buf()
	lw := buf.Len()
	buf.Resize(nw)
	b := buf[lw:]
	if tail > head {
		for i := uint(0); i < nw; i++ {
			b[i] = uint32(c.data[hw+i])
		}
	} else {
		// Ring wrapped.
		end := c.nWords << c.log2MemElts
		n_end := end - hw
		for i := uint(0); i < n_end; i++ {
			b[i] = uint32(c.data[hw+i])
		}
		for i := uint(0); i < tw; i++ {
			b[n_end+i] = uint32(c.data[i])
		}
	}
	c.regs.host_mem[c.index].n_entries_read.Set(uint32(n))
	c.sequence += uint(n)
	r.Data = buf
	r.free = c.free
	return r
}

func (d *FifoDma) FifoDmaSync(i uint) FifoDmaData { return d.Channels[i].poll() }

func (d *FifoDma) FifoDmaInit(resultFifo chan FifoDmaData, channel uint,
	a Address, cmd Command, nMemBits, log2MemElts uint) {
	const (
		min = 6
		max = 15
	)
	if log2MemElts < min {
		log2MemElts = min
	}
	if log2MemElts > max {
		log2MemElts = max
	}

	// Find a free fifo dma channel.
	i := channel
	ch := &d.Channels[i]
	if ch.resultFifo != nil {
		panic("channel in use")
	}
	ch.resultFifo = resultFifo
	ch.free = make(chan elib.Uint32Vec, 64)
	ch.log2MemElts = log2MemElts

	// Number of data words in fifo element.
	bits := nMemBits
	ch.nWords = uint(bits / 32)
	if bits%32 != 0 {
		ch.nWords++
	}
	n := ch.nWords << log2MemElts

	if c := uint(cap(ch.data)); c < n {
		if c > 0 {
			ch.data.Free(ch.data_heap_id)
		}
		ch.data, ch.data_heap_id = dma_dataAlloc(n)
	} else {
		ch.data = ch.data[:n]
	}

	ch.regs.sbus_start_address[i].Set(uint32(a))
	ch.regs.host_mem_start_address[i].Set(ch.start_address())
	ch.regs.sbus_opcode[i].set(cmd)
	// Interrupt when half full.
	ch.regs.host_mem_interrupt_threshold[i].Set(1 << (log2MemElts - 1))

	v := ch.regs.control[i].Get()
	v |= uint32(1 << 0) // enable bit
	v = (v &^ (0xf << 7)) | uint32(log2MemElts-min)<<7
	v = (v &^ (0x1f << 2)) | (uint32(ch.nWords) << 2)
	ch.regs.control[i].Set(v)
}

func (m *FifoDma) InitChannels(regs *FifoDmaRegs) {
	for i := range m.Channels {
		c := &m.Channels[i]
		c.index = uint(i)
		c.regs = regs
	}
}
