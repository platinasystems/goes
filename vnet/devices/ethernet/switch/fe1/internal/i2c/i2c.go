// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package i2c

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/icpu"

	"fmt"
	"sync"
)

const (
	master = 0
	slave  = 1
)

type Operation int

const (
	Quick Operation = iota

	SendByte
	RcvByte

	WriteByteData
	ReadByteData

	WriteWordData
	ReadWordData

	WriteBlockData
	ReadBlockData

	ProcessCall

	BlockProcessCall
)

type request struct {
	op Operation

	data [32 + 2]byte

	nData int

	status status

	interrupt_status uint32

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
	*icpu.I2cController
	IcpuController *icpu.Controller
	requestFifo    chan *request
	lock           sync.Mutex
}

func (r *request) start(bus *I2c) {
	regs := bus.I2cController

	for i := 0; i < r.nData; i++ {
		x := uint32(r.data[i])
		if i+1 == r.nData {
			x |= 1 << 31
		}
		regs.Data_fifo[master].Write.Set(bus.IcpuController, x)
	}

	cmd := uint32(r.op)<<9 | 1<<31
	regs.Command[master].Set(bus.IcpuController, cmd)
}

func (q *request) finish(s *I2c) {
	regs := s.I2cController
	q.status = status((regs.Command[master].Get(s.IcpuController) >> 25) & 0x7)

	if q.status == ok {
		q.nData = 0
		for {
			x := regs.Data_fifo[master].Read.Get(s.IcpuController)
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

	select {
	case b := <-s.requestFifo:
		b.start(s)
	default:
	}
}

func (i *I2c) Interrupt() {
	status := i.Interrupt_status_write_1_to_clear.Get(i.IcpuController)
	i.Interrupt_status_write_1_to_clear.Set(i.IcpuController, status)

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

	req.addData(address<<1 | byte(rw))

	switch op {
	case Quick, SendByte, RcvByte:
	default:
		req.addData(command)
	}

	switch op {
	case WriteBlockData, ReadBlockData:
		if nData > 32 {
			panic(fmt.Errorf("block too large %d bytes", nData))
		}
		req.addData(byte(nData))
	}

	copy(req.data[req.nData:], data[:nData])
	req.nData += nData

	req.do(s)
	<-req.done

	nRead = req.nData
	copy(data[:], req.data[:nRead])
	err = req.status.ToError()

	return
}

func (i *I2c) Init() {
	ir := i.IcpuController

	i.Config.Set(ir, i.Config.Get(ir)|1<<30)

	enable := ^uint32(0)
	i.Interrupt_enable.Set(ir, enable)
}
