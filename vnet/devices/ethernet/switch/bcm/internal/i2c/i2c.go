// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This code implements an SMBus driver interface for a BCM i2c
// controller, operating in master mode, utilizing the CMIC.
// << NOTE: the SMBus does not implement all i2c commands >>
//
package i2c

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/iproc"

	"fmt"
	"sync"
)

const (
	// SMBus modes - if you ain't the master, then your a slave! (-;
	master = 0
	slave  = 1
)

type Operation int

const (
	// Standard SMBUS operations.
	// See Documentation/i2c/smbus-protocol from linux kernel tree.
	// NOTE: Currently unimplemented SMBus operations: Alert, ARP, Notify, PEC

	// Send a single bit to the device, at the place of the Rd/Wr bit.
	Quick Operation = iota

	// Write/Read a single byte to/from a device, without specifying a device register.
	SendByte
	RcvByte

	// Write/Read a single byte from a device, from a designated register.
	// The register is specified through the Comm byte (1st byte sent).
	WriteByteData
	ReadByteData

	// As above, but using 2 bytes of data.
	WriteWordData
	ReadWordData

	// Write/Read a block of up to 32 bytes from a device, from a
	// designated register that is specified through the Comm byte. The amount
	// of data is specified by first byte of data.
	WriteBlockData
	ReadBlockData

	// This command selects a device register (through the Comm byte), sends
	// 16 bits of data to it, and reads 16 bits of data in return.
	ProcessCall

	// This command selects a device register (through the Comm byte), sends
	// 1 to 31 bytes of data to it, and reads 1 to 31 bytes of data in return.
	BlockProcessCall
)

type request struct {
	op Operation

	// Space for max block size (32) plus comm byte plus block length.
	data [32 + 2]byte

	// Number of bytes to send for request; number of bytes received in response.
	nData int

	status status

	interrupt_status uint32

	// Pointer sent when reply is ready.
	done chan *request
}

type status int

const (
	ok status = iota
	lost_arbitration
	slave_nack_not_present
	slave_nack_busy
	slave_timeout
	tx_fifo_underrun
	rx_fifo_overrun
)

var status_strings = []string{
	ok:                     "OK",
	lost_arbitration:       "lost arbitration",
	slave_nack_not_present: "slave nack device not present",
	slave_nack_busy:        "slave nack device busy",
	slave_timeout:          "slave timeout",
	tx_fifo_underrun:       "tx fifo underrun",
	rx_fifo_overrun:        "rx fifo overrun",
}

var ErrDeviceNotPresent = slave_nack_not_present.ToError()

func (x status) ToError() error {
	if x == ok {
		return nil
	}
	return x
}

func (x status) Error() (s string) { return status_strings[x] }

func (r *request) addData(b byte) {
	r.data[r.nData] = b
	r.nData++
}

type I2c struct {
	*iproc.I2cRegs
	IprocRegs   *iproc.Regs
	requestFifo chan *request
	lock        sync.Mutex
}

func (r *request) start(bus *I2c) {
	regs := bus.I2cRegs

	// Fill up fifo
	for i := 0; i < r.nData; i++ {
		x := uint32(r.data[i])
		if i+1 == r.nData {
			// Set write end for last data byte.
			x |= 1 << 31
		}
		regs.Data_fifo[master].Write.Set(bus.IprocRegs, x)
	}

	// Set operation plus start bit.
	cmd := uint32(r.op)<<9 | 1<<31
	regs.Command[master].Set(bus.IprocRegs, cmd)
}

func (q *request) finish(s *I2c) {
	// Fetch request status
	regs := s.I2cRegs
	q.status = status((regs.Command[master].Get(s.IprocRegs) >> 25) & 0x7)

	// Fetch data from rx fifo.
	if q.status == ok {
		q.nData = 0
		for {
			x := regs.Data_fifo[master].Read.Get(s.IprocRegs)
			status := x >> 30
			const (
				empty = iota
				start
				middle
				end
			)
			if status != empty {
				q.addData(byte(x & 0xff))
			}
			if status == empty || status == end {
				break
			}
		}
	}

	if q.done != nil {
		q.done <- q
	}

	// Either start next request or leave hardware idle.
	select {
	case b := <-s.requestFifo:
		b.start(s)
	default:
	}
}

func (i *I2c) Interrupt() {
	// Fetch status; clear interrupt.
	status := i.Interrupt_status_write_1_to_clear.Get(i.IprocRegs)
	i.Interrupt_status_write_1_to_clear.Set(i.IprocRegs, status)

	select {
	case a := <-i.requestFifo:
		a.interrupt_status = status
		a.finish(i)
	default:
	}
}

func (a *request) do(s *I2c) {
	if s.requestFifo == nil {
		s.requestFifo = make(chan *request, 64)
	}
	if a.done == nil {
		a.done = make(chan *request, 1)
	}

	s.lock.Lock()
	s.requestFifo <- a
	l := len(s.requestFifo)
	s.lock.Unlock()

	if l == 1 {
		a.start(s)
	}
}

type RW int

const (
	Write RW = iota
	Read
)

type Data [32]byte

func (s *I2c) Do(rw RW, address byte, op Operation, command byte, data *Data, nData int) (nRead int, err error) {
	req := request{
		op: op,
	}

	// First send 7 bit i2c bus address + read/write bit.
	req.addData(address<<1 | byte(rw))

	// For some operations send Comm byte.
	switch op {
	case Quick, SendByte, RcvByte:
		// no Comm byte for these commands.
	default:
		// Otherwise insert Comm byte.
		req.addData(command)
	}

	// For block operations send block length in bytes.
	switch op {
	case WriteBlockData, ReadBlockData:
		if nData > 32 {
			panic(fmt.Errorf("block too large %d bytes", nData))
		}
		req.addData(byte(nData))
	}

	// Copy caller's data.
	copy(req.data[req.nData:], data[:nData])
	req.nData += nData

	// Perform request and block until done.
	req.do(s)
	<-req.done

	nRead = req.nData
	copy(data[:], req.data[:nRead])
	err = req.status.ToError()

	return
}

func (i *I2c) Init() {
	ir := i.IprocRegs

	// Set bus enable bit.
	i.Config.Set(ir, i.Config.Get(ir)|1<<30)

	// Enable all interrupts
	enable := ^uint32(0)
	i.Interrupt_enable.Set(ir, enable)
}
