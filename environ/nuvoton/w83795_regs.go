package w83795

const ()

type reg8 byte
type reg16 [2]byte
type regi16 reg16

// Memory map
type hwmRegs struct {
	_           [0x4]byte
	TempCntl1   reg8
	TempCntl2   reg8
	_           [0x1b]byte
	FrontTemp   reg8
	RearTemp    reg8
	_           [0x0b]byte
	FanCount    [14]reg8
	FractionLSB reg8
}

type genRegs struct {
	Reg reg8
}
