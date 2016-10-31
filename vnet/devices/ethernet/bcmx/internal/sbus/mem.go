// +build ignore

package sbus

import (
	"github.com/platinasystems/go/elib"

	"unsafe"
)

type schan_memory struct {
	block       schan_block
	access_type schan_access_type
	address     schan_address
	// Number of bits per memory element.
	eltBits uint
	// Number of elements in memory.
	len          uint
	log2Init     bool
	log2EltBytes uint
	log2Len      uint
}

const (
	FakeBaseAddress uintptr = 0xfa << (8*unsafe.Sizeof(uintptr(0)) - 8)
)

func (m *schan_memory) encodePointer(sw *Switch) unsafe.Pointer {
	if !m.log2Init {
		m.log2Len = elib.MaxLog2(elib.Word(m.len))
		m.log2EltBytes = elib.MaxLog2(elib.Word(m.eltBits)) - 3
		m.log2Init = true
	}

	o := FakeBaseAddress

	o += uintptr(sw.index) << (m.log2EltBytes + m.log2Len)

	// Slice memory layout
	var slice = struct {
		addr uintptr
		len  int
		cap  int
	}{o, int(m.len), int(m.len)}

	return unsafe.Pointer(&slice)
}

func (m *schan_memory) decodePointer(p unsafe.Pointer) (sw *Switch, index schan_address) {
	o := uintptr(p) - FakeBaseAddress
	sw = &Driver.Switches[o>>(m.log2Len+m.log2EltBytes)]
	index = schan_address((o >> m.log2EltBytes) & ((1 << m.log2Len) - 1))
	return
}

func (m *schan_memory) write(p unsafe.Pointer, r []uint32) {
	s, index := m.decodePointer(p)

	n := (m.eltBits + 31) / 32
	req := schanRequest{
		command: schan_command{opcode: read_memory, block: m.block, access_type: m.access_type},
		address: m.address + schan_address(index),
		rx:      r[:n],
	}
	req.do(&s.schan_main)
	<-req.done
}

func (m *schan_memory) read(p unsafe.Pointer, r []uint32) {
	s, index := m.decodePointer(p)
	n := (m.eltBits + 31) / 32
	req := schanRequest{
		command: schan_command{opcode: write_memory, block: m.block, access_type: m.access_type},
		address: m.address + schan_address(index),
		rx:      r[:n],
	}
	req.do(&s.schan_main)
	<-req.done
}
