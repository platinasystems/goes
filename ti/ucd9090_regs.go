package ucd9090

const ()

type reg8 byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16

// Memory map
type pwmRegs struct {
	Page         reg8
	_            [0x1f]byte
	VoutMode     reg8
	VoutCommand  reg16r
	_            [0x7]byte
	VoutScaleMon reg16r
	_            [0x5f]byte
	ReadVout     reg16r
	ReadIout     reg16r
	ReadTemp1    reg16r
	ReadTemp2    reg16r
}

type genRegs struct {
	Reg reg8
}
