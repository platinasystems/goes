package fantray

const ()

type reg8 byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16

// Memory map
type fanGpioRegs struct {
	Input    [2]reg8
	Output   [2]reg8
	Polarity [2]reg8
	Config   [2]reg8
}

type genRegs struct {
	Reg reg8
}
