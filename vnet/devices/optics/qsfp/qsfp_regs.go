package qsfp

type reg8 byte
type reg8b byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16

// Memory map
// offsets are 16-bit to accomodate the mixed 8-bit 16-bit accesses
// the offset function has a divide by two to restore to proper address
type regs struct {
	_                 [0x83 * 2]byte
	SpecCompliance    reg8 // 0x83
	_                 byte
	_                 [0x10 * 2]byte
	VendorName        reg8b // 0x94
	_                 byte
	_                 [0x13 * 2]byte
	VendorPN          reg8b // 0xA8
	_                 byte
	_                 [0x17 * 2]byte
	ExtSpecCompliance reg8 // 0xC0
	_                 byte
	_                 [0x3 * 2]byte
	VendorSN          reg8b // 0xC4
}
