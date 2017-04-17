package w83795d

const ()

type reg8 byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16

// Memory map
type regsBank0 struct {
	BankSelect    reg8
	Configuration reg8
	_             [0x2]byte
	TempCntl1     reg8 //0x04
	TempCntl2     reg8 //0x05
	_             [0x1b]byte
	FrontTemp     reg8 //0x21
	RearTemp      reg8 //0x22
	_             [0x0b]byte
	FanCount      [14]reg8 //0x2e
	FractionLSB   reg8     //0x3c
}

type regsBank2 struct {
	BankSelect            reg8
	FanControlModeSelect1 reg8
	TempToFanMap1         reg8
	TempToFanMap2         reg8
	_                     [0x4]byte
	FanControlModeSelect2 reg8 //0x08
	_                     [0x4]byte
	FanStepUpTime         reg8 //0x0d
	FanStepDownTime       reg8 //0x0e
	FanOutputModeControl  reg8 //0x0f
	FanOutValue1          reg8 //0x10
	FanOutValue2          reg8 //0x11
	_                     [0x6]byte
	FanPwmPrescale1       reg8 //0x18
	FanPwmPrescale2       reg8 //0x19
	_                     [0x6]byte
	FanStartValue1        reg8 //0x20
	FanStartValue2        reg8 //0x21
	_                     [0x6]byte
	FanStopValue1         reg8 //0x28
	FanStopValue2         reg8 //0x29
	_                     [0x6]byte
	FanStopTime1          reg8 //0x30
	FanStopTime2          reg8 //0x31
	_                     [0x2e]byte
	TargetTemp1           reg8 //0x60
	TargetTemp2           reg8 //0x61
	_                     [0x6]byte
	FanCritTemp1          reg8 //0x68
	FanCritTemp2          reg8 //0x69
	_                     [0x6]byte
	TempHyster1           reg8 //0x70
	TempHyster2           reg8 //0x71
}
