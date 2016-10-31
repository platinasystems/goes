package sbus

// SBUS control dma.

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/hw"

	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type dma_control uint32

func (x *dma_control) get() dma_control  { return dma_control(((*hw.Reg32)(x)).Get()) }
func (x *dma_control) set(v dma_control) { ((*hw.Reg32)(x)).Set(uint32(v)) }

const (
	dma_start            dma_control = 1 << 0
	dma_abort            dma_control = 1 << 1
	dma_mode_descriptors dma_control = 1 << 2
	dma_mode_registers   dma_control = 0 << 2
	dma_append_cpu       dma_control = 1 << 28
	dma_jump             dma_control = 1 << 29 // jump to descriptor at given cpu address; rest of descriptor is ignored
	dma_skip             dma_control = 1 << 30 // skip to next descriptor; ignore this descriptor.
	dma_last_in_chain    dma_control = 1 << 31
)

type dma_descriptor struct {
	// As above.
	control dma_control

	/* [31] decrement sbus address instead of increment
	   [30] disable increment of cpu address (same data for all sbus entries)
	   [29] disable increment of sbus address
	   [28:24] shift for sbus address increment
	   [23:16] core clocks between pending sbus commands
	   [13] disable write to cpu memory
	   [12] word swap 64 data
	   [11] cpu mem write big endian
	   [10] cpu mem read big endian
	   [9:5] number of 32 bit data words in request
	   [4:0] number of 32 bit data words in response */
	options hw.Reg32

	/* Number of operations to execute. */
	count hw.Reg32

	command command_reg

	/* Starting SBUS and cpu memory address for dma. */
	sbus_address Address
	cpu_address  hw.Reg32
}

type dma_data uint32

//go:generate gentemplate -d Package=sbus -id dma_descriptor -d VecType=dma_descriptor_vec -d Type=dma_descriptor github.com/platinasystems/go/elib/hw/dma_mem.tmpl
//go:generate gentemplate -d Package=sbus -id dma_data -d VecType=dma_data_vec -d Type=dma_data github.com/platinasystems/go/elib/hw/dma_mem.tmpl

type dma_status uint32

func (x *dma_status) get() dma_status  { return dma_status(hw.LoadUint32((*uint32)(x))) }
func (x *dma_status) set(v dma_status) { hw.StoreUint32((*uint32)(x), uint32(v)) }

func (q *DmaRequest) toError() error {
	if q.status&dma_error == 0 {
		return nil
	}
	return q
}

const (
	dma_done dma_status = 1 << iota
	dma_error
	dma_cpu_write_error
	dma_cpu_read_error
	dma_parity_error
	dma_ack_wrong_beat_count
	dma_ack_wrong_opcode
	dma_nack
	dma_ack_error
	dma_ack_timeout
	dma_desc_read_error
	dma_active
	dma_descriptor_active
	dma_multi_bit_ecc_error
	dma_all_errors dma_status = (dma_cpu_write_error |
		dma_cpu_read_error |
		dma_parity_error |
		dma_ack_wrong_beat_count |
		dma_ack_wrong_opcode |
		dma_nack |
		dma_ack_error |
		dma_ack_timeout |
		dma_desc_read_error |
		dma_multi_bit_ecc_error)
)

var dma_status_details = []string{
	2:  "CPU write error",
	3:  "CPU read error",
	4:  "parity error",
	5:  "wrong beat count",
	6:  "wrong opcode",
	7:  "nack",
	8:  "ack with error bit set",
	9:  "timeout",
	10: "descriptor read error",
	13: "multi bit ecc error",
}

func (q *DmaRequest) Error() (s string) {
	s = "success"
	x := q.status
	if x&dma_error != 0 {
		ei := q.error_command_index
		cmd := q.Commands[ei].GetCmd()
		s = fmt.Sprintf("[%d] %s %s: ", q.error_command_index, BlockToString(cmd.Command.Block), cmd)
		if details := x & dma_all_errors; details != 0 {
			s += elib.FlagStringer(dma_status_details, elib.Word(details))
		}
	}
	return
}

type DmaRegs struct {
	// Single descriptor registers for register mode.
	desc dma_descriptor
	// Descriptor address.  Incremented by hardware as descriptors are processed.
	desc_address hw.Reg32
	status       dma_status

	current struct {
		cpu_address        hw.Reg32
		sbus_address       hw.Reg32
		descriptor_address hw.Reg32

		/* From descriptor. */
		request hw.Reg32

		/* Number of operations to execute. */
		count hw.Reg32

		/* Starting SBUS and host mem address. */
		sbus_start_address Address
		cpu_start_address  hw.Reg32

		command command_reg

		debug             hw.Reg32
		debug_clear       hw.Reg32
		ecc_error_address hw.Reg32
		ecc_error_control hw.Reg32
	}
}

type DmaCmd struct {
	// Request
	Command Command
	Address Address

	FixedCpuAddress  bool // keep cpu/sbus address the same
	FixedSbusAddress bool

	DecrementSbusAddress     bool  // decrement instead of increment
	Log2SbusAddressIncrement uint8 // increment sbus address by 1 << Log2SbusAddressIncrement

	CoreClocksBetweenCommands uint8
	Count                     uint // number of commands to issue
	Tx                        []uint32
	Rx                        []uint32
}

type DmaCmdInterface interface {
	GetCmd() *DmaCmd
	Pre()
	Post()
	String() string
}

func (d *DmaCmd) IsRead() bool {
	return d.Command.Opcode == ReadRegister || d.Command.Opcode == ReadMemory
}
func (d *DmaCmd) IsWrite() bool   { return !d.IsRead() }
func (d *DmaCmd) GetCmd() *DmaCmd { return d }
func (d *DmaCmd) Pre()            {}
func (d *DmaCmd) Post()           {}

func (d *DmaCmd) String() (s string) {
	s = fmt.Sprintf("%s %s", &d.Command, d.Address)
	if d.Count != 1 {
		s += fmt.Sprintf(" count %d", d.Count)
	}
	if len(d.Tx) > 0 {
		s += fmt.Sprintf(" tx %x", d.Tx)
	}
	if len(d.Rx) > 0 {
		s += fmt.Sprintf(" rx #%d", len(d.Rx))
	}
	return s
}

type DmaRequest struct {
	dma *Dma

	Commands []DmaCmdInterface

	// Value of dma status register when request finished.
	status dma_status

	// If status indicates error index into req.Commands where error occurred.
	error_command_index uint

	// Go error from above status.  Nil if no error.
	Err error

	// Pointer sent when all descriptors are done.
	Done chan *DmaRequest

	// Used to mark finished requests so they can be reused again to start new requests.
	isDone bool
}

type dma_channel struct {
	index        uint
	regs         *DmaRegs
	desc         dma_descriptor_vec
	desc_heap_id elib.Index
	data         dma_data_vec
	data_heap_id elib.Index

	// Protects following.
	mu         sync.Mutex
	reqFifo    chan *DmaRequest
	currentReq *DmaRequest
}

type Dma struct {
	// Current channel; increments round robin through all channels
	channel_index uint32
	Channels      []dma_channel
}

var BlockToString func(b Block) string

func (ch *dma_channel) start(req *DmaRequest) {
	ch.currentReq = req
	nDesc := uint(len(req.Commands))
	if c := uint(cap(ch.desc)); c < nDesc {
		if c > 0 {
			ch.desc.Free(ch.desc_heap_id)
		}
		ch.desc, ch.desc_heap_id = dma_descriptorAlloc(nDesc)
	} else {
		ch.desc = ch.desc[:nDesc]
	}

	for i := range req.Commands {
		req.Commands[i].Pre()
	}

	n := uint(0)
	for i := range req.Commands {
		dc := req.Commands[i].GetCmd()
		n += uint(len(dc.Rx) + len(dc.Tx))
	}
	if c := uint(cap(ch.data)); c < n {
		if c > 0 {
			ch.data.Free(ch.data_heap_id)
		}
		ch.data, ch.data_heap_id = dma_dataAlloc(n)
	} else {
		ch.data = ch.data[:n]
	}

	idata := 0
	for idesc := range ch.desc {
		d := dma_descriptor{}
		var o DmaCmd
		o = *req.Commands[idesc].GetCmd()

		if idesc == 0 {
			d.cpu_address = hw.Reg32(ch.data[0].PhysAddress())
		} else {
			d.control |= dma_append_cpu
		}
		if uint(idesc)+1 == nDesc {
			d.control |= dma_last_in_chain
		}

		count := o.Count
		if count == 0 {
			count = 1
		}
		if o.Command.Size == 0 {
			o.Command.Size = uint(len(o.Tx)) * 4 / count
		}
		o.Command.dma = true
		d.command.set(o.Command)
		d.sbus_address = o.Address
		d.count = hw.Reg32(count)

		if o.DecrementSbusAddress {
			d.options |= 1 << 31
		}
		if o.FixedCpuAddress {
			d.options |= 1 << 30
		}
		if o.FixedSbusAddress {
			d.options |= 1 << 29
		}

		if o.Log2SbusAddressIncrement > 31 {
			panic(fmt.Errorf("sbus address increment too large: %d > 31", o.Log2SbusAddressIncrement))
		}
		d.options |= hw.Reg32(o.Log2SbusAddressIncrement) << 24

		if o.CoreClocksBetweenCommands > 0xff {
			panic(fmt.Errorf("core clocks too large: %d > 255", o.CoreClocksBetweenCommands))
		}
		d.options |= hw.Reg32(o.CoreClocksBetweenCommands) << 16

		if o.Command.Size/4 > 31 {
			panic(fmt.Errorf("message tx words too large: %d > 31", len(o.Tx)))
		}
		if d.count > 1 && len(o.Rx)%int(d.count) != 0 {
			panic(fmt.Errorf("message rx words %d must be divisible by count %d", len(o.Rx), d.count))
		}
		rxSize := len(o.Rx) / int(d.count)
		if rxSize > 31 {
			panic(fmt.Errorf("message rx words too large: %d > 31", rxSize))
		}
		d.options |= hw.Reg32(o.Command.Size/4) << 5
		d.options |= hw.Reg32(rxSize) << 0

		for i := range o.Tx {
			ch.data[idata] = dma_data(o.Tx[i])
			idata++
		}
		for _ = range o.Rx {
			ch.data[idata] = 0xdeadbeef // poison
			idata++
		}

		ch.desc[idesc] = d
	}

	ch.regs.desc_address.Set(uint32(ch.desc[0].PhysAddress()))
	hw.MemoryBarrier()
	ch.regs.desc.control.set(dma_start | dma_mode_descriptors)
}

// Get a request from fifo if there is one.
func (ch *dma_channel) getReq() (req *DmaRequest) {
	select {
	case req = <-ch.reqFifo:
		break
	default:
	}
	return
}

// Add a request return the number of requests pending (including given request).
func (ch *dma_channel) putReq(req *DmaRequest) (newLen int) {
	ch.reqFifo <- req
	newLen = len(ch.reqFifo)
	return
}

func (ch *dma_channel) finish(req *DmaRequest) {
	// Get status and received message data.
	req.status = ch.regs.status

	req.Err = req.toError()
	if req.Err != nil {
		// If status indicates error find offending command from hardware.
		hi := uintptr(ch.regs.desc_address.Get())
		lo := ch.desc[0].PhysAddress()
		req.error_command_index = uint((hi - lo) / unsafe.Sizeof(ch.desc[0]))
	}

	idata := 0
	for idesc := range req.Commands {
		o := req.Commands[idesc].GetCmd()
		idata += len(o.Tx)
		for j := range o.Rx {
			o.Rx[j] = uint32(ch.data[idata])
			idata++
		}
	}

	for i := range req.Commands {
		req.Commands[i].Post()
	}

	if req.Err != nil {
		panic(req.Err)
	}

	// Stop dma hardware & ack interrupt.
	ch.regs.desc.control.set(0)

	// Either start next request or leave hardware idle.
	if nextReq := ch.getReq(); nextReq != nil {
		ch.start(nextReq)
	} else {
		ch.currentReq = nil
	}

	req.isDone = true
	req.Done <- req
}

func (q *DmaRequest) Reset() {
	// Reset length so request may be reused.
	if len(q.Commands) > 0 {
		q.Commands = q.Commands[:0]
	}
}

func (q *DmaRequest) Add(c DmaCmdInterface) {
	if q.isDone {
		q.Reset()
		q.isDone = false
	}
	q.Commands = append(q.Commands, c)
}

func (c *dma_channel) Interrupt() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.currentReq != nil {
		c.finish(c.currentReq)
	} else {
		status := c.regs.status.get()
		c.regs.desc.control.set(0)
		panic(fmt.Errorf("request fifo empty: channel %d dma status 0x%x", c.index, status))
	}
}

func (s *Dma) chooseChannel() (c *dma_channel) {
	// Round robin among channels.
	var i0, i1 uint32
	for {
		i0 = atomic.LoadUint32(&s.channel_index)
		i1 = i0 + 1
		if i1 >= uint32(len(s.Channels)) {
			i1 = 0
		}
		if atomic.CompareAndSwapUint32(&s.channel_index, i0, i1) {
			break
		}
	}
	c = &s.Channels[i1]
	return
}

func (a *DmaRequest) do(s *Dma) {
	if a.dma == nil {
		a.dma = s
	}

	if a.Done == nil {
		a.Done = make(chan *DmaRequest, 1)
	}

	c := s.chooseChannel()

	c.mu.Lock()
	defer c.mu.Unlock()

	if elog.Enabled() {
		a.log(c.index, "do")
	}

	if c.currentReq != nil {
		c.putReq(a)
	} else {
		c.start(a)
	}
}

func (s *Dma) Start(req *DmaRequest) { req.do(s) }

func (s *Dma) Do(req *DmaRequest) {
	req.do(s)
	<-req.Done
}

func (m *Dma) InitChannels(regs []DmaRegs) {
	m.Channels = make([]dma_channel, len(regs))
	for i := range m.Channels {
		c := &m.Channels[i]
		c.index = uint(i)
		c.regs = &regs[i]
		c.reqFifo = make(chan *DmaRequest, 64)
	}
}

// Synchronous DMA read or write.
func (s *Dma) rw(cmd Command, a Address, v uint64, nBits int, isWrite, panicError bool) (u uint64, err error) {
	var buf [2]uint32
	dc := DmaCmd{
		Command: cmd,
		Address: a,
	}
	if isWrite {
		buf[0] = uint32(v)
		buf[1] = uint32(v >> 32)
		dc.Tx = buf[:nBits/32]
	} else {
		dc.Rx = buf[:nBits/32]
	}
	req := DmaRequest{
		Commands: []DmaCmdInterface{&dc},
	}
	req.do(s)
	<-req.Done
	if isWrite {
		u = v
	} else {
		u = uint64(buf[0])
		if nBits > 32 {
			u |= uint64(buf[1]) << 32
		}
	}
	err = req.toError()
	if err != nil && panicError {
		panic(err)
	}
	return
}

func (s *Dma) read(b Block, a Address, access AccessType, nBits int) (x uint64) {
	cmd := Command{Opcode: ReadRegister, Block: b, AccessType: access}
	x, _ = s.rw(cmd, a, 0, nBits, false, true)
	return
}

func (s *Dma) write(b Block, a Address, access AccessType, nBits int, v uint64) {
	cmd := Command{Opcode: WriteRegister, Block: b, AccessType: access}
	s.rw(cmd, a, v, nBits, true, true)
}

func (s *Dma) Read32A(b Block, a Address, c AccessType) uint32     { return uint32(s.read(b, a, c, 32)) }
func (s *Dma) Read64A(b Block, a Address, c AccessType) uint64     { return s.read(b, a, c, 64) }
func (s *Dma) Write64A(b Block, a Address, c AccessType, v uint64) { s.write(b, a, c, 64, v) }
func (s *Dma) Write32A(b Block, a Address, c AccessType, v uint32) { s.write(b, a, c, 32, uint64(v)) }

func (s *Dma) Read32(b Block, a Address) uint32     { return s.Read32A(b, a, Unique0) }
func (s *Dma) Write32(b Block, a Address, v uint32) { s.Write32A(b, a, Unique0, v) }
func (s *Dma) Read64(b Block, a Address) uint64     { return s.Read64A(b, a, Unique0) }
func (s *Dma) Write64(b Block, a Address, v uint64) { s.Write64A(b, a, Unique0, v) }

func (req *DmaRequest) ReadWrite(cmd Command, a Address, v []uint32, isWrite bool) {
	dc := DmaCmd{Command: cmd, Address: a}
	if isWrite {
		dc.Tx = v[:1]
	} else {
		dc.Rx = v[:1]
	}
	req.Add(&dc)
}

func (req *DmaRequest) Read(cmd Command, a Address, v []uint32)  { req.ReadWrite(cmd, a, v, false) }
func (req *DmaRequest) Write(cmd Command, a Address, v []uint32) { req.ReadWrite(cmd, a, v, true) }

type rw32 struct {
	DmaCmd
	result *uint32
	buf    [1]uint32
}

func (r *rw32) Pre() {
	if r.IsRead() {
		r.Rx = r.buf[:]
	} else {
		r.buf[0] = *r.result
		r.Tx = r.buf[:]
	}
}
func (r *rw32) Post() {
	if r.IsRead() {
		*r.result = r.buf[0]
	}
}

func (q *DmaRequest) doRw32(v *uint32, o Opcode, b Block, a Address, c AccessType) {
	r := rw32{
		DmaCmd: DmaCmd{
			Command: Command{Opcode: o, Block: b, AccessType: c},
			Address: a,
		},
		result: v,
	}
	q.Add(&r)
}
func (q *DmaRequest) GetReg32(v *uint32, b Block, a Address, c AccessType) {
	q.doRw32(v, ReadRegister, b, a, c)
}
func (q *DmaRequest) SetReg32(v uint32, b Block, a Address, c AccessType) {
	q.doRw32(&v, WriteRegister, b, a, c)
}
func (q *DmaRequest) GetMem32(v *uint32, b Block, a Address, c AccessType) {
	q.doRw32(v, ReadMemory, b, a, c)
}
func (q *DmaRequest) SetMem32(v uint32, b Block, a Address, c AccessType) {
	q.doRw32(&v, WriteMemory, b, a, c)
}

type rw64 struct {
	DmaCmd
	buf    [2]uint32
	result *uint64
}

func (r *rw64) Pre() {
	if r.IsRead() {
		r.Rx = r.buf[:]
	} else {
		v := *r.result
		r.buf[0] = uint32(v)
		r.buf[1] = uint32(v >> 32)
		r.Tx = r.buf[:]
	}
}
func (r *rw64) Post() {
	if r.IsRead() {
		*r.result = uint64(r.buf[0]) | uint64(r.buf[1])<<32
	}
}

func (q *DmaRequest) doRw64(v *uint64, o Opcode, b Block, a Address, c AccessType) {
	r := rw64{
		DmaCmd: DmaCmd{
			Command: Command{Opcode: o, Block: b, AccessType: c},
			Address: a,
		},
		result: v,
	}
	q.Add(&r)
}
func (q *DmaRequest) GetReg64(v *uint64, b Block, a Address, c AccessType) {
	q.doRw64(v, ReadRegister, b, a, c)
}
func (q *DmaRequest) SetReg64(v uint64, b Block, a Address, c AccessType) {
	q.doRw64(&v, WriteRegister, b, a, c)
}
func (q *DmaRequest) GetMem64(v *uint64, b Block, a Address, c AccessType) {
	q.doRw64(v, ReadMemory, b, a, c)
}
func (q *DmaRequest) SetMem64(v uint64, b Block, a Address, c AccessType) {
	q.doRw64(&v, WriteMemory, b, a, c)
}

// Event logging.
func (q *DmaRequest) log(channel uint, tag string) {
	cmd := q.Commands[0].GetCmd()
	e := schanEvent{
		Channel: byte(channel),
		Opcode:  cmd.Command.Opcode,
		Block:   cmd.Command.Block,
		Address: cmd.Address,
	}
	copy(e.Tag[:], tag)
	e.Log()
}
