package nct7802yd

const ()

type reg8 byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16

// Memory map of nct7802 hwmonitor
type regsBank0 struct {
	BankSelect		reg8 //0x00
	RearTemp		reg8 //0x01
	FrontTemp     		reg8 //0x02
	_			[0x02]byte
	FractionLSB		reg8 //0x05
	_			[0x17]byte
	SmiTempStats		reg8 //0x1d
	_			[0x02]byte
	TcritRealStats		reg8 //0x20	
	StartCtrl		reg8 //0x21
	ModeSel			reg8 //0x22
	_			[0x01]byte
	FanEnable		reg8 //0x24
	VmonEnable		reg8 //0x25
	_			[0x09]byte
	SmiControl		reg8 //0x2f
	_			[0x0a]byte
	TcritThresh1		reg8 //0x3a
	TcritThresh2		reg8 //0x3b
	_			[0x14]byte
	SmiTempMask		reg8 //0x50
	_			[0x02]byte
	TcritMask		reg8 //0x53
	_			[0x0a]byte
	FanOutputModeControl	reg8 //0x5e
	_			[0x01]byte
	FanOutValue1		reg8 //0x60
	_			[0x03]byte
	TempToFanMap1		reg8 //0x64
	TempToFanMap2           reg8 //0x65
	FanCtrlConfig1		reg8 //0x66
	FanCtrlConfig2		reg8 //0x67
	FanCtrlConfig3		reg8 //0x68
	FanCtrlConfig4		reg8 //0x69
	_			[0x04]byte
	FanStepUpTime		reg8 //0x6e
	FanStepDownTime		reg8 //0x6f
	_			[0x01]byte
	FanPwmPrescale1		reg8 //0x71
	_			[0x02]byte
	TempHyster1		reg8 //0x74
	TempHyster2		reg8 //0x75
	_			[0x01]byte
	FanStartValue1		reg8 //0x77
	FanStopTime1		reg8 //0x78
	FanNonStopEn		reg8 //0x79
	_			[0x0a]byte
	FanCritTemp1		reg8 //0x84
	_			[0x0f]byte
	FanCritTemp2		reg8 //0x94
	_			[0x4e]byte
	TargetTemp1		reg8 //0xe3
	TargetTemp2		reg8 //0xe4
	_			[0x17]byte
	SoftReset		reg8 //0xfc
}
