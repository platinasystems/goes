package ucd9090

const ()

type reg8 byte
type reg16 [2]byte
type regi16 reg16

// Memory map
type pwmRegs struct {
	Page         reg8
	_            [0x1f]byte
	VoutMode     reg8
	VoutCommand  reg16
	_            [0x7]byte
	VoutScaleMon reg16
	_            [0x5f]byte
	ReadVout     reg16
	ReadIout     reg16
	ReadTemp1    reg16
	ReadTemp2    reg16
}

type genRegs struct {
	Reg reg8
}
