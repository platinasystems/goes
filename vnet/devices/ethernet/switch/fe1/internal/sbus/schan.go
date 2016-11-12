// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sbus

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/hw"

	"fmt"
)

type Sbus struct {
	PIO
	Dma
	FifoDma
}

type PIO struct {
	Controller     *PIOController
	FastController *FastPIOController
	requestFifo    chan *request
}

type request struct {
	// Request
	Command Command
	Address Address
	Tx      []uint32
	// Reply
	Status control
	Rx     []uint32
	// Pointer sent when reply is ready.
	Done chan *request
}

type PIOController struct {
	/* start [0] done [1] abort [2] */
	control                  control
	n_data_words_in_last_ack hw.U32
	// [31:26] opcode, [25:20] dst block, [19:14] src block,
	// [13:7] data len, [6] is error [5:4] error code, [0] nack
	error   hw.U32
	command command_u32
	message [21]hw.U32
}

type Opcode uint8
type Block uint8

const (
	ReadMemory       Opcode = 0x07
	ReadMemoryAck    Opcode = 0x08
	WriteMemory      Opcode = 0x09
	WriteMemoryAck   Opcode = 0x0a
	ReadRegister     Opcode = 0x0b
	ReadRegisterAck  Opcode = 0x0c
	WriteRegister    Opcode = 0x0d
	WriteRegisterAck Opcode = 0x0e
	FifoPop          Opcode = 0x2a
	FifoPopAck       Opcode = 0x2b
	FifoPush         Opcode = 0x2c
	FifoPushAck      Opcode = 0x2d
)

//go:generate stringer -type=Opcode

type AccessType uint8

const (
	Unique0          AccessType = 0 // 0-7
	Duplicate        AccessType = 9
	AddressSplitDist AccessType = 10
	AddressSplit     AccessType = 12
	DataSplit0       AccessType = 14 // 0-3
	Single           AccessType = 20
)

func Unique(i uint) AccessType {
	return Unique0 + AccessType(i)
}

type command_u32 uint32

const (
	command_nack command_u32 = 1 << 0
	command_dma  command_u32 = 1 << 3
)

type Command struct {
	Opcode     Opcode
	Block      Block
	AccessType AccessType
	Size       uint // in bytes
	nack       bool
	dma        bool
}

func (c *Command) String() string {
	s := c.Opcode.String()
	if c.Size > 0 {
		s += fmt.Sprintf(" size:%d", c.Size)
	}
	if c.nack {
		s += fmt.Sprintf(" nack")
	}
	if c.dma {
		s += fmt.Sprintf(" dma")
	}
	return s
}

func (c *command_u32) set(a Command) {
	r := command_u32(uint32(a.Opcode)<<26 |
		uint32(a.Block)<<19 |
		uint32(a.AccessType)<<14 |
		uint32(a.Size)<<7)
	if a.dma {
		r |= command_dma
	}
	hw.StoreUint32((*uint32)(c), uint32(r))
}

func (c *command_u32) get() (a Command) {
	r := command_u32(hw.LoadUint32((*uint32)(c)))
	a.Opcode = Opcode(r >> 26)
	a.Block = Block((r >> 19) & 0x7f)
	a.AccessType = AccessType((r >> 14) & 0x1f)
	a.Size = uint((r >> 7) & 0x7f)
	if r&command_nack != 0 {
		a.nack = true
	}
	if r&command_dma != 0 {
		a.dma = true
	}
	return
}

type Address uint32

func (a Address) String() string {
	v := uint32(a)
	if a&GenReg != 0 {
		return fmt.Sprintf("0x%08x", v)
	} else {
		// For per-port registers separate port.
		return fmt.Sprintf("0x%08x[%d]", v&^0xff, v&0xff)
	}
}

const (
	PortReg Address = 0 << 25
	GenReg  Address = 1 << 25
)

func (a *Address) set(v Address) {
	hw.StoreUint32((*uint32)(a), uint32(v))
}

func (a *Address) get() Address {
	return Address(hw.LoadUint32((*uint32)(a)))
}

type FastPIOController struct {
	enable_command hw.U32
	write_is_busy  hw.U32
	command        command_u32
	address        Address
	data32         hw.U32
	data64         [2]hw.U32 // lo/hi
}

func (f *FastPIOController) read32(b Block, addr Address) uint32 {
	f.command.set(Command{Opcode: ReadRegister, Block: b})
	f.address.set(addr)
	hw.MemoryBarrier()
	return f.data32.Get()
}

func (f *FastPIOController) read64(b Block, addr Address) uint64 {
	f.command.set(Command{Opcode: ReadRegister, Block: b})
	f.address.set(addr)
	hw.MemoryBarrier()
	lo, hi := f.data64[0].Get(), f.data64[1].Get()
	return uint64(lo) | (uint64(hi) << 32)
}

func (f *FastPIOController) write32(b Block, addr Address, value uint32) {
	f.command.set(Command{Opcode: WriteRegister, Block: b, Size: 4})
	f.address.set(addr)
	f.data32.Set(value)
	hw.MemoryBarrier()
	for f.write_is_busy.Get() != 0 {
	}
}

func (f *FastPIOController) write64(b Block, addr Address, value uint64) {
	f.command.set(Command{Opcode: WriteRegister, Block: b, Size: 8})
	f.address.set(addr)
	f.data64[1].Set(uint32(value >> 32))
	f.data64[0].Set(uint32(value >> 0))
	hw.MemoryBarrier()
	for f.write_is_busy.Get() != 0 {
	}
}

func (s *PIO) FastRead32(b Block, addr Address) uint32 {
	return s.FastController.read32(b, addr)
}
func (s *PIO) FastWrite32(b Block, addr Address, value uint32) {
	s.FastController.write32(b, addr, value)
}

type control uint32

func (x *control) get() control  { return control(hw.LoadUint32((*uint32)(x))) }
func (x *control) set(v control) { hw.StoreUint32((*uint32)(x), uint32(v)) }

const (
	start         control = 1 << 0
	done          control = 1 << 1 // Clear bit to clear done interrupt.
	abort         control = 1 << 2
	error_present control = 1 << 23 // One of error detail bits set.
	parity_error  control = 1 << 20 // Error detail bits
	nack          control = 1 << 21
	timeout       control = 1 << 22
)

var error_details = []string{
	20: "parity error", 21: "nack", 22: "timeout",
}

func (x control) Error() (s string) {
	s = "success"
	if x&error_present != 0 {
		s = "abort" // no detail bits => abort
		if details := x & (parity_error | nack | timeout); details != 0 {
			s = elib.FlagStringer(error_details, elib.Word(details))
		}
	}
	return
}

func (x control) toError() error {
	if x&error_present == 0 {
		return nil
	}
	return x
}

func (a *request) start(s *PIO) {
	// Size in bytes of message.
	a.Command.Size = uint(len(a.Tx)) * 4

	r := s.Controller
	r.command.set(a.Command)
	r.message[0].Set(uint32(a.Address))
	for i := range a.Tx {
		r.message[1+i].Set(a.Tx[i])
	}
	hw.MemoryBarrier()
	r.control.set(start)
}

func (a *request) finish(s *PIO) {
	// Fetch request status
	r := s.Controller
	a.Status = r.control.get()
	for i := range a.Rx {
		a.Rx[i] = r.message[i].Get()
	}
	if a.Done != nil {
		a.Done <- a
	}

	// Either start next request or leave hardware idle.
	select {
	case b := <-s.requestFifo:
		b.start(s)
	default:
		r.control.set(0)
	}
}

func (a *request) do(s *PIO) {
	if s.requestFifo == nil {
		s.requestFifo = make(chan *request, 64)
	}
	if a.Done == nil {
		a.Done = make(chan *request, 1)
	}
	s.requestFifo <- a
	if len(s.requestFifo) == 1 {
		a.start(s)
	}
}

func (s *PIO) DoneInterrupt() {
	select {
	case a := <-s.requestFifo:
		a.finish(s)
	default:
		s.Controller.control.set(0)
	}
}

func (s *PIO) rw(cmd Command, a Address, v uint64, nBits int, isWrite, panicError bool) (u uint64, err error) {
	var buf [4]uint32
	req := request{
		Command: cmd,
		Address: a,
	}
	if isWrite {
		buf[0] = uint32(v)
		buf[1] = uint32(v >> 32)
		req.Tx = buf[:nBits/32]
	} else {
		req.Rx = buf[:nBits/32]
	}
	req.do(s)
	<-req.Done
	err = req.Status.toError()
	if isWrite {
		u = v
	} else {
		u = uint64(buf[0])
		if nBits > 32 {
			u |= uint64(buf[1]) << 32
		}
	}
	if err != nil && panicError {
		panic(err)
	}
	return
}

func (s *PIO) read(b Block, a Address, access AccessType, nBits int) (x uint64) {
	cmd := Command{Opcode: ReadRegister, Block: b, AccessType: access}
	x, _ = s.rw(cmd, a, 0, nBits, false, true)
	return
}

func (s *PIO) write(b Block, a Address, access AccessType, nBits int, v uint64) {
	cmd := Command{Opcode: WriteRegister, Block: b, AccessType: access}
	s.rw(cmd, a, v, nBits, true, true)
}

func (s *PIO) Read32A(b Block, a Address, c AccessType) uint32     { return uint32(s.read(b, a, c, 32)) }
func (s *PIO) Read64A(b Block, a Address, c AccessType) uint64     { return s.read(b, a, c, 64) }
func (s *PIO) Write64A(b Block, a Address, c AccessType, v uint64) { s.write(b, a, c, 64, v) }
func (s *PIO) Write32A(b Block, a Address, c AccessType, v uint32) { s.write(b, a, c, 32, uint64(v)) }

func (s *PIO) Read32(b Block, a Address) uint32     { return s.Read32A(b, a, Unique0) }
func (s *PIO) Write32(b Block, a Address, v uint32) { s.Write32A(b, a, Unique0, v) }
func (s *PIO) Read64(b Block, a Address) uint64     { return s.Read64A(b, a, Unique0) }
func (s *PIO) Write64(b Block, a Address, v uint64) { s.Write64A(b, a, Unique0, v) }

func (s *PIO) Read128(b Block, addr Address, buf []uint32) error {
	req := request{
		Command: Command{Opcode: ReadMemory, Block: b},
		Address: addr,
		Rx:      buf[:],
	}
	req.do(s)
	<-req.Done
	return req.Status.toError()
}

func (s *PIO) Write128(b Block, addr Address, data []uint32) error {
	var buf [4]uint32
	req := request{
		Command: Command{Opcode: WriteMemory, Block: b},
		Address: addr,
		Tx:      buf[:],
	}
	for i := range data {
		req.Tx[i] = data[i]
	}
	req.do(s)
	<-req.Done
	return req.Status.toError()
}

// Event logging.
type schanEvent struct {
	Address Address
	Opcode  Opcode
	Block   Block
	Channel byte
	Tag     [elog.EventDataBytes - 4 - 3*1]byte
}

func (e *schanEvent) String() string {
	return fmt.Sprintf("fe1 schan %d: %s %s %s %s", e.Channel,
		elog.String(e.Tag[:]), e.Opcode.String(), BlockToString(e.Block), e.Address.String())
}
func (e *schanEvent) Encode(b []byte) int {
	i := 0
	i, b[0] = i+1, e.Channel
	i += elog.EncodeUint32(b[i:], uint32(e.Opcode))
	i += elog.EncodeUint32(b[i:], uint32(e.Block))
	i += elog.EncodeUint32(b[i:], uint32(e.Address))
	copy(b[i:], e.Tag[:])
	return i
}
func (e *schanEvent) Decode(buf []byte) (i int) {
	var o, b, a uint32
	i, e.Channel = i+1, buf[0]
	o, i = elog.DecodeUint32(buf, i)
	b, i = elog.DecodeUint32(buf, i)
	a, i = elog.DecodeUint32(buf, i)
	copy(e.Tag[:], buf[i:])
	e.Opcode, e.Block, e.Address = Opcode(o), Block(b), Address(a)
	return
}

//go:generate gentemplate -d Package=sbus -id schanEvent -d Type=schanEvent github.com/platinasystems/go/elib/elog/event.tmpl
