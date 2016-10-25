package ledgpio

const ()

type reg8 byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16

// Memory map
type ledRegs struct {
	Input0    reg8
	Input1    reg8
	Output0   reg8
	Output1   reg8
	Polarity0 reg8
	Polarity1 reg8
	Config0   reg8
	Config1   reg8
}

type genRegs struct {
	Reg reg8
}
