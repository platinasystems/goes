package ucd9090d

const ()

type reg8 byte
type reg8b byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16

// Memory map
type regs struct {
	Page              reg8 //0x00
	_                 byte
	_                 [0x10 * 2]byte
	StoreDefaultAll   reg8 //0x11
	_                 byte
	_                 [0xe * 2]byte
	VoutMode          reg8 //0x20
	_                 byte
	VoutCommand       reg16r //0x21
	_                 [0x8 * 2]byte
	VoutScaleMon      reg16r //0x2a
	_                 [0x60 * 2]byte
	ReadVout          reg16r //0x8b
	ReadIout          reg16r //0x8c
	ReadTemp1         reg16r //0x8d
	ReadTemp2         reg16r //0x8e
	_                 [0x48 * 2]byte
	RunTimeClock      reg8b //0xd7
	_                 byte
	_                 [0x12 * 2]byte
	LoggedFaults      reg8b //0xea
	_                 byte
	LoggedFaultIndex  reg16 //0xeb
	LoggedFaultDetail reg8b //0xec
	_                 byte
	_                 [0xf * 2]byte
	MiscConfig        reg8b //0xfc
}
