package fspd

type reg8 byte
type reg8b byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16
type reg32B [32]byte

// Memory map
// offsets are 16-bit to accomodate the mixed 8-bit 16-bit accesses
// the offset function has a divide by two to restore to proper address
type regs struct {
	Page        reg8 // 0x00
	_           byte
	Operation   reg8 // 0x01
	_           byte
	OnOffConfig reg8 // 0x02
	_           byte
	ClearFaults reg8 // 0x03
	_           byte
	_           [0x1c * 2]byte
	VoutMode    reg8 // 0x20
	_           byte
	_           [0x58 * 2]byte
	StatusWord  reg16r // 0x79
	StatusVout  reg8   // 0x7a
	_           byte
	StatusIout  reg8 // 0x7b
	_           byte
	StatusInput reg8 // 0x7c
	_           byte
	StatusTemp  reg8 // 0x7d
	_           byte
	_           [0x03 * 2]byte
	StatusFans  reg8 // 0x81
	_           byte
	_           [0x04 * 2]byte
	Ein         reg8 // 0x86
	_           byte
	Eout        reg8 // 0x87
	_           byte
	Vin         reg16r // 0x88
	Iin         reg16r // 0x89
	_           [0x01 * 2]byte
	Vout        reg16r // 0x8b
	Iout        reg16r // 0x8c
	Temp1       reg16r // 0x8d
	Temp2       reg16r // 0x8e
	_           [0x01 * 2]byte
	FanSpeed    reg16 // 0x90
	_           [0x05 * 2]byte
	Pout        reg16r // 0x96
	Pin         reg16r // 0x97
	PMBusRev    reg8   // 0x98
	_           byte
	MfgId       reg8b // 0x99
	_           byte
	MfgMod      reg8b // 0x9a
	_           byte
}

type regsE struct {
	block [8]reg32B
}
