package packet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/hw"

	"fmt"
	"sync/atomic"
)

type descFlag uint16

const (
	desc_next_valid, desc_next_valid_bit descFlag = 1 << iota, iota
	// scatter/gather bit in spec
	desc_scatter_gather, desc_scatter_gather_bit
	desc_reload, desc_reload_bit
	desc_module_header_valid, desc_module_header_valid_bit
	desc_tx_stats_update, desc_tx_stats_update_bit
	desc_tx_pause, desc_tx_pause_bit
	desc_tx_purge, desc_tx_purge_bit
	desc_descriptor_interrupt, desc_descriptor_interrupt_bit
	desc_controlled_interrupt, desc_controlled_interrupt_bit
)

var descFlagStrings = [...]string{
	desc_next_valid_bit:           "next-valid",
	desc_scatter_gather_bit:       "scatter-gather",
	desc_reload_bit:               "reload",
	desc_module_header_valid_bit:  "module header valid",
	desc_tx_stats_update_bit:      "tx stats update",
	desc_tx_pause_bit:             "tx pause",
	desc_tx_purge_bit:             "tx purge",
	desc_descriptor_interrupt_bit: "desc interrupt",
	desc_controlled_interrupt_bit: "controlled interrupt",
}

func (x descFlag) String() string { return elib.FlagStringer(descFlagStrings[:], elib.Word(x)) }

type control_reg uint32

const (
	control_tx = 1 << iota
	control_start
	control_abort
	// 1 => interrupt per descriptor; 0 => interrupt per packet
	control_interrupt_per_descriptor
	control_big_endian_packets
	control_big_endian_descriptors
	control_rx_drop_on_no_descriptors_available
	control_reload_status_update_disable
	control_interrupt_based_on_descriptor
	control_continuous
	control_rx = 0 << 0
)

type status_reg int32

const (
	bad_status_write_address   status_reg = 1 << (4 * 3)
	bad_packet_address         status_reg = 1 << (4 * 4)
	bad_descriptor_address     status_reg = 1 << (4 * 5)
	rx_status_buffer_ecc_error status_reg = 1 << (4 * 6)
	rx_packet_buffer_ecc_error status_reg = 1 << (4 * 7)
	status_reg_all_errors      status_reg = 0x11111000
)

const n_channel = 4

type DmaRegs struct {
	// [8] enable => backpressure mmu if outstanding cell count exceeds threshold (default: 1)
	// [7:0] threshold 1 <= t <= 128 (default 4)
	rx_buffer_threshold [n_channel]hw.Reg32

	// In continuous mode, hardware stops fetching if current descriptor address matches this address.
	halt_descriptor_address [n_channel]hw.Reg32

	// [30:27] channel halt because halt_address = curr_desc_address
	// Read-only; cleared when descriptor_halt_address moves.
	halt_status hw.Reg32

	_ [3]hw.Reg32

	// [9] enable continuous dma (enables channel in halt interrupt)
	// [8] 0 => desc done interrupt for every done descriptor; 1 => only when interrupt bit is set in descriptor.
	// [7] reload status update disable
	// [6] 1 => drop packet on end of chain (no descriptors available); 0 => block waiting for descriptor
	// [5] 1 => big endian descriptors
	// [4] 1 => big endian packet operations
	// [3] 1 => interrupt after descriptor; 0 => interrupt after complete packet.
	// [2] abort dma
	// [1] enable dma
	// [0] direction 0 => rx, 1 => tx
	control [n_channel]hw.Reg32

	// [31:28] ecc error reading rx packet buffer; cleared by control enable
	// [27:24] ecc error reading rx status buffer; cleared by control enable
	// [23:20] descriptor read address error; cleared by control enable
	// [19:16] packet read/write address error; cleared by control enable
	// [15:12] status write address error; cleared by control enable
	// [11:8]  dma channel active
	// [7:4]   current descriptor done
	// [3:0]   descriptor chain done
	status hw.Reg32

	_ [1]hw.Reg32

	// Chain start address.
	start_descriptor_address [n_channel]hw.Reg32

	// Bitmap of cos accepted for channel.
	rx_cos_control [n_channel][2]hw.Reg32

	// [31] enable
	// [30:16] descriptor count
	// [15:0] timer (units? probably 125MHz core clock).
	interrupt_coallesce [n_channel]hw.Reg32

	// ??
	rx_buffer_threshold_config hw.Reg32

	// 64 bit cos mask for this cmc.  Masks cos from this cmc.
	programmable_cos_mask [2]hw.Reg32

	// [11:8] clear descriptor controlled interrupt
	// [7:4]  clear coallescing interrupt
	// [3:0]  clear descriptor read complete
	status_write_1_to_clear hw.Reg32

	// Incremented as descriptors are processed by hardware.
	current_descriptor_address [n_channel]hw.Reg32

	_ [2]hw.Reg32
}

const (
	status_write_1_to_clear_read_complete   = 0
	status_write_1_to_clear_coallescing     = 4
	status_write_1_to_clear_desc_controlled = 8
)

type dma_channel struct {
	dma  *Dma
	regs *DmaRegs

	// Channel index 0-3
	index uint32

	tx_node *txNode

	start_control control_reg
}

type Dma struct {
	regs        *DmaRegs
	Channels    [n_channel]dma_channel
	rx_channels []*dma_channel
	tx_channels []*dma_channel
	txNode
	rxNode
	InterruptEnable   func(enable bool)
	interruptsEnabled bool
}

type dma_error struct {
	memory_address uint32
	index          uint32
	status         uint32
}

func (e *dma_error) toError() error {
	if e.status == 0 {
		return nil
	}
	return e
}

var status_reg_details = []string{
	12: "bad rx status write address",
	16: "bad packet read/write address",
	20: "bad descriptor read address",
	24: "rx status buffer ecc error",
	28: "rx packet buffer ecc error",
}

func (e *dma_error) Error() (s string) {
	s = "success"
	x := e.status
	if x != 0 {
		s = fmt.Sprintf("[%d] 0x%x: ", e.index, e.memory_address)
		s += elib.FlagStringer(status_reg_details, elib.Word(x))
	}
	return
}

func (c *dma_channel) ack_desc_controlled_interrupt() {
	c.regs.status_write_1_to_clear.Set(1 << (c.index + status_write_1_to_clear_desc_controlled))
}

func (c *dma_channel) DescControlledInterrupt() {
	v := c.regs.status.Get()
	if v>>12 != 0 {
		panic(fmt.Errorf("dma error status %s", elib.FlagStringer(status_reg_details, elib.Word(v))))
	}
	d := c.dma
	if !d.interruptsEnabled {
		return
	}
	if c.start_control&control_tx != 0 {
		d.txNode.DescDoneInterrupt()
	} else {
		n := &d.rxNode
		n.Activate(true)
		atomic.AddInt32(&n.active_count, 1)
	}
}

func (m *Dma) InitChannels(regs *DmaRegs) {
	m.interruptsEnabled = true
	for i := range m.Channels {
		c := &m.Channels[i]
		c.index = uint32(i)
		c.dma = m
		c.regs = regs
		is_rx := i == 0
		c.start_control = control_start | control_continuous | control_interrupt_based_on_descriptor
		if is_rx {
			m.rx_channels = append(m.rx_channels, c)
			c.start_control |= control_rx
		} else {
			m.tx_channels = append(m.tx_channels, c)
			c.start_control |= control_tx
		}
	}
}
