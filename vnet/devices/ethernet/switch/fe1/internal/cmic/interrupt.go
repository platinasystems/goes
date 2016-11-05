// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmic

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"

	"fmt"
	"io"
	"sort"
)

const (
	schan_done_interrupt                    interrupt = 20
	mdio_done_interrupt                     interrupt = 7
	cross_coupled_memory_dma_done_interrupt interrupt = 21

	sbus_dma0_interrupt          interrupt = 1
	sbus_dma1_interrupt          interrupt = 0
	sbus_dma2_interrupt          interrupt = 6
	sbus_dma_ecc_error_interrupt interrupt = 26

	fifo_dma0_interrupt interrupt = 5
	fifo_dma1_interrupt interrupt = 4
	fifo_dma2_interrupt interrupt = 3
	fifo_dma3_interrupt interrupt = 2

	packet_dma3_desc_done_interrupt interrupt = 8
	packet_dma2_desc_done_interrupt interrupt = 10
	packet_dma1_desc_done_interrupt interrupt = 12
	packet_dma0_desc_done_interrupt interrupt = 14

	packet_dma3_chain_done_interrupt interrupt = 9
	packet_dma2_chain_done_interrupt interrupt = 11
	packet_dma1_chain_done_interrupt interrupt = 13
	packet_dma0_chain_done_interrupt interrupt = 15

	packet_dma3_coalesce_interrupt interrupt = 16
	packet_dma2_coalesce_interrupt interrupt = 17
	packet_dma1_coalesce_interrupt interrupt = 18
	packet_dma0_coalesce_interrupt interrupt = 19

	packet_dma0_desc_controlled_interrupt interrupt = 27
	packet_dma1_desc_controlled_interrupt interrupt = 28
	packet_dma2_desc_controlled_interrupt interrupt = 29
	packet_dma3_desc_controlled_interrupt interrupt = 30

	sw0_interrupt interrupt = 22
	sw1_interrupt interrupt = 23
	sw2_interrupt interrupt = 24
	sw3_interrupt interrupt = 25

	i2c_interrupt                          interrupt = 32 + 0
	pcie_ecc_error_interrupt               interrupt = 32 + 1
	time_sync_interrupt                    interrupt = 32 + 2
	phy_scan_link_status_interrupt         interrupt = 32 + 3
	phy_scan_pause_status_interrupt        interrupt = 32 + 4
	spi_interrupt                          interrupt = 32 + 5
	uart0_interrupt                        interrupt = 32 + 6
	uart1_interrupt                        interrupt = 32 + 7
	common_schan_op_done_interrupt         interrupt = 32 + 8
	common_mii_op_done_interrupt           interrupt = 32 + 9
	gpio_interrupt                         interrupt = 32 + 10
	chip_func0_interrupt                   interrupt = 32 + 11
	chip_func1_interrupt                   interrupt = 32 + 12
	chip_func2_interrupt                   interrupt = 32 + 13
	chip_func3_interrupt                   interrupt = 32 + 14
	chip_func4_interrupt                   interrupt = 32 + 15
	chip_func5_interrupt                   interrupt = 32 + 16
	chip_func6_interrupt                   interrupt = 32 + 17
	chip_func7_interrupt                   interrupt = 32 + 18
	uc0_pmu_interrupt                      interrupt = 32 + 19
	uc1_pmu_interrupt                      interrupt = 32 + 20
	wdt0_interrupt                         interrupt = 32 + 21
	wdt1_interrupt                         interrupt = 32 + 22
	tim00_interrupt                        interrupt = 32 + 23
	tim01_interrupt                        interrupt = 32 + 24
	tim10_interrupt                        interrupt = 32 + 25
	tim11_interrupt                        interrupt = 32 + 26
	pcie_interface_needs_cleanup_interrupt interrupt = 32 + 27
	sram_ecc_interrupt                     interrupt = 32 + 28
	ser_irq_interrupt                      interrupt = 32 + 29
)

type interrupt uint

type interruptHandler struct {
	handler func()
	count   uint64
}

var interruptNames = []string{
	schan_done_interrupt:                    "schan done",
	mdio_done_interrupt:                     "mdio done",
	cross_coupled_memory_dma_done_interrupt: "cc memory dma done",
	sbus_dma0_interrupt:                     "sbus dma ch 0 done",
	sbus_dma1_interrupt:                     "sbus dma ch 1 done",
	sbus_dma2_interrupt:                     "sbus dma ch 2 done",
	sbus_dma_ecc_error_interrupt:            "sbus dma ecc error",
	fifo_dma0_interrupt:                     "fifo dma ch 0 done",
	fifo_dma1_interrupt:                     "fifo dma ch 1 done",
	fifo_dma2_interrupt:                     "fifo dma ch 2 done",
	fifo_dma3_interrupt:                     "fifo dma ch 3 done",

	packet_dma0_desc_done_interrupt: "packet dma ch 0 desc done",
	packet_dma1_desc_done_interrupt: "packet dma ch 1 desc done",
	packet_dma2_desc_done_interrupt: "packet dma ch 2 desc done",
	packet_dma3_desc_done_interrupt: "packet dma ch 3 desc done",

	packet_dma0_chain_done_interrupt: "packet dma ch 0 chain done",
	packet_dma1_chain_done_interrupt: "packet dma ch 1 chain done",
	packet_dma2_chain_done_interrupt: "packet dma ch 2 chain done",
	packet_dma3_chain_done_interrupt: "packet dma ch 3 chain done",

	packet_dma0_coalesce_interrupt: "packet dma ch 0 coalesce",
	packet_dma1_coalesce_interrupt: "packet dma ch 1 coalesce",
	packet_dma2_coalesce_interrupt: "packet dma ch 2 coalesce",
	packet_dma3_coalesce_interrupt: "packet dma ch 3 coalesce",

	packet_dma0_desc_controlled_interrupt: "packet dma ch 0 desc controlled",
	packet_dma1_desc_controlled_interrupt: "packet dma ch 1 desc controlled",
	packet_dma2_desc_controlled_interrupt: "packet dma ch 2 desc controlled",
	packet_dma3_desc_controlled_interrupt: "packet dma ch 3 desc controlled",

	sw0_interrupt: "software interrupt 0",
	sw1_interrupt: "software interrupt 1",
	sw2_interrupt: "software interrupt 2",
	sw3_interrupt: "software interrupt 3",

	i2c_interrupt:                          "i2c",
	pcie_ecc_error_interrupt:               "pcie error",
	time_sync_interrupt:                    "time sync",
	phy_scan_link_status_interrupt:         "phy scan link status",
	phy_scan_pause_status_interrupt:        "phy scan pause status",
	spi_interrupt:                          "spi",
	uart0_interrupt:                        "uart0",
	uart1_interrupt:                        "uart1",
	common_schan_op_done_interrupt:         "common schan done",
	common_mii_op_done_interrupt:           "common mii done",
	gpio_interrupt:                         "gpio",
	chip_func0_interrupt:                   "chip function 0",
	chip_func1_interrupt:                   "chip function 1",
	chip_func2_interrupt:                   "chip function 2",
	chip_func3_interrupt:                   "chip function 3",
	chip_func4_interrupt:                   "chip function 4",
	chip_func5_interrupt:                   "chip function 5",
	chip_func6_interrupt:                   "chip function 6",
	chip_func7_interrupt:                   "chip function 7",
	uc0_pmu_interrupt:                      "ucontroller 0 pmu",
	uc1_pmu_interrupt:                      "ucontroller 1 pmu",
	wdt0_interrupt:                         "watchdog 0",
	wdt1_interrupt:                         "watchdog 1",
	tim00_interrupt:                        "tim00 interrupt",
	tim01_interrupt:                        "tim01 interrupt",
	tim10_interrupt:                        "tim10 interrupt",
	tim11_interrupt:                        "tim11 interrupt",
	pcie_interface_needs_cleanup_interrupt: "pcie interface needs cleanup",
	sram_ecc_interrupt:                     "sram ecc",
	ser_irq_interrupt:                      "ser",
}

func (i interrupt) String() string {
	if int(i) < len(interruptNames) {
		return interruptNames[i]
	}
	return fmt.Sprintf("%d %d", i/32, i%32)
}

//go:generate gentemplate -d Package=cmic -id interruptHandler -d VecType=interruptHandlerVec -d Type=interruptHandler github.com/platinasystems/go/elib/vec.tmpl

func (c *Cmic) setInterruptHandler(which interrupt, h func()) {
	c.interruptHandlers.Validate(uint(which))
	c.interruptHandlers[which].handler = h
}

func (c *Cmic) Interrupt() (nInt uint) {
	r := &c.regs.cmc[0]
	c.interruptCount++
	for i := range r.irq_status0 {
		s := r.irq_status0[i].Get()
		if s != 0 {
			nInt += c.intr(uint(i), s)
		}
	}
	for i := range r.irq_status1 {
		s := r.irq_status1[i].Get()
		if s != 0 {
			nInt += c.intr(uint(len(r.irq_status0)+i), s)
		}
	}
	return
}

func (c *Cmic) intr(i uint, status uint32) (nInt uint) {
	s := elib.Word(status)
	for s != 0 {
		b := s.FirstSet()
		l := b.MinLog2()
		irq := l + 32*i
		h := &c.interruptHandlers[irq]
		h.count++
		nInt++
		if h.handler != nil {
			if elog.Enabled() {
				e := irqEvent{i: uint32(irq)}
				e.Log()
			}
			h.handler()
		}
		s ^= b
	}
	return
}

// Event logging.
type irqEvent struct {
	i uint32
}

func (e *irqEvent) String() string          { return fmt.Sprintf("fe1 irq %s", interrupt(e.i).String()) }
func (e *irqEvent) Encode(b []byte) int     { return elog.EncodeUint32(b, e.i) }
func (e *irqEvent) Decode(b []byte) (i int) { e.i, i = elog.DecodeUint32(b, i); return }

//go:generate gentemplate -d Package=cmic -id irqEvent -d Type=irqEvent github.com/platinasystems/go/elib/elog/event.tmpl

func (c *Cmic) interrupt_init() {
	r := &c.regs.cmc[0]

	c.interruptHandlers.Validate(uint(32*(len(r.irq_status0)+len(r.irq_status1)) - 1))

	// Enable MSI
	r.pcie_msi_config.Set(1<<4 | r.pcie_msi_config.Get())

	// Enable all interrupts
	enable := ^uint32(0)
	for i := range r.irq_enable0[0] {
		r.irq_enable0[0][i].Set(enable)
	}
	for i := range r.irq_enable1[0] {
		r.irq_enable1[0][i].Set(enable)
	}
}

func (cm *Cmic) interruptEnable(i, mask uint32, enable bool) {
	v := cm.regs.cmc[0].irq_enable0[0][i].Get()
	if enable {
		v |= mask
	} else {
		v &^= mask
	}
	cm.regs.cmc[0].irq_enable0[0][i].Set(v)
}

type interruptSummary struct {
	Interrupt string `format:"%-30s"`
	Count     uint64 `format:"%16d"`
}
type byName []interruptSummary

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Interrupt < a[j].Interrupt }

func (c *Cmic) WriteInterruptSummary(w io.Writer) {
	data := byName{}
	for i, h := range c.interruptHandlers {
		if h.count != 0 {
			data = append(data, interruptSummary{
				Interrupt: interrupt(i).String(),
				Count:     h.count,
			})
		}
	}
	sort.Sort(data)
	data = append(data, interruptSummary{Interrupt: "Total", Count: c.interruptCount})
	elib.TabulateWrite(w, data)
}
