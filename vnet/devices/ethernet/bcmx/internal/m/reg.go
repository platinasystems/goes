package m

import (
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"

	"unsafe"
)

var (
	RegsBasePointer = hw.RegsBasePointer
	RegsBaseAddress = hw.RegsBaseAddress
)

type Greg32 byte
type Greg64 byte
type Preg32 byte
type Preg64 byte

const Log2NRegPorts = 8

type Reg32 [1 << Log2NRegPorts]Greg32
type Reg64 [1 << Log2NRegPorts]Greg64
type PortReg32 [1 << Log2NRegPorts]Preg32
type PortReg64 [1 << Log2NRegPorts]Preg64

func (r *Greg32) Offset() uint { return uint(uintptr(unsafe.Pointer(r)) - RegsBaseAddress) }
func (r *Greg64) Offset() uint { return (*Greg32)(r).Offset() }
func (r *Preg32) Offset() uint { return (*Greg32)(r).Offset() }
func (r *Preg64) Offset() uint { return (*Greg32)(r).Offset() }

func (r *Reg32) Offset() uint { return r[0].Offset() }
func (r *Reg64) Offset() uint { return r[0].Offset() }

func (r *Preg32) Address() sbus.Address { return sbus.Address(r.Offset()) | sbus.PortReg }
func (r *Reg32) Address() sbus.Address  { return sbus.Address(r.Offset()) | sbus.GenReg }
func (r *Preg64) Address() sbus.Address { return sbus.Address(r.Offset()) | sbus.PortReg }
func (r *Reg64) Address() sbus.Address  { return sbus.Address(r.Offset()) | sbus.GenReg }

func (r *Reg32) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint32) {
	q.GetReg32(v, b, a|r.Address(), c)
}
func (r *Reg32) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint32) {
	q.SetReg32(v, b, a|r.Address(), c)
}

func (r *Reg64) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint64) {
	q.GetReg64(v, b, a|r.Address(), c)
}
func (r *Reg64) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint64) {
	q.SetReg64(v, b, a|r.Address(), c)
}

func (r *Preg32) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint32) {
	q.GetReg32(v, b, a|r.Address(), c)
}
func (r *Preg32) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint32) {
	q.SetReg32(v, b, a|r.Address(), c)
}

func (r *Preg64) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint64) {
	q.GetReg64(v, b, a|r.Address(), c)
}
func (r *Preg64) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint64) {
	q.SetReg64(v, b, a|r.Address(), c)
}

func (r *Greg32) address() sbus.Address { return sbus.Address(r.Offset()) | sbus.GenReg }
func (r *Greg64) address() sbus.Address { return sbus.Address(r.Offset()) | sbus.GenReg }

func (r *Greg32) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint32) {
	q.GetReg32(v, b, a|r.address(), c)
}
func (r *Greg32) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint32) {
	q.SetReg32(v, b, a|r.address(), c)
}

func (r *Greg64) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint64) {
	q.GetReg64(v, b, a|r.address(), c)
}
func (r *Greg64) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint64) {
	q.SetReg64(v, b, a|r.address(), c)
}
