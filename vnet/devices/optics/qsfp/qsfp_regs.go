package qsfp

type reg8 byte
type reg8b byte
type reg16 [2]byte
type reg16r [2]byte
type regi16 reg16

type blocks struct {
	lpage0b reg8b //0
	_       [31]byte
	lpage1b reg8b //32
	_       [31]byte
	lpage2b reg8b //64
	_       [31]byte
	lpage3b reg8b //96
	_       [31]byte
	upage0b reg8b //128
	_       [31]byte
	upage1b reg8b //160
	_       [31]byte
	upage2b reg8b //192
	_       [31]byte
	upage3b reg8b //224
	_       [31]byte
}

// Memory map
type regsUpage0 struct {
	_                 [128]byte
	identifier        reg8 //128
	extIdentifier     reg8 //129
	connectorType     reg8 //130
	SpecCompliance    reg8 //131
	_                 [2]byte
	gigEthCompliance  reg8  //134
	fibreLinkLength   reg8  //135
	fibreTxTech       reg8  //136
	fibreTxMedia      reg8  //137
	fibreSpeed        reg8  //138
	encoding          reg8  //139
	brNominal         reg8  //140
	extRateSelectTag  reg8  //141
	lengthSmf         reg8  //142
	lengthOm3         reg8  //143
	lengthOm2         reg8  //144
	lengthOm1         reg8  //145
	lengthDacAocOm4   reg8  //146
	deviceTech        reg8  //147
	VendorName        reg8b //148
	_                 [15]reg8
	extendedModule    reg8    //164
	vendorOui         [3]reg8 //165
	VendorPN          reg8b   //168
	_                 [15]reg8
	vendorRev         reg16   //184
	wlCuAtten         reg16   //186
	wlTolCuAtten      reg16   //188
	maxCaseTemp       reg8    //190
	cc_base           reg8    //191
	ExtSpecCompliance reg8    //192
	options           [3]reg8 //193
	VendorSN          reg8b   //196
	_                 [15]reg8
	vendorDate        reg8b //212
	_                 [7]byte
	diagMonType       reg8 //220
	enhancedOptions   reg8 //221
	brNominal2        reg8 //222
	cc_ext            reg8 //223
	vendorProm        reg8 //224-255
	_                 [31]byte
}

type regsLpage0 struct {
	id                       reg8
	status                   reg16
	_                        [83]byte
	txDisable                reg8 //86
	rxRateSelect             reg8 //87
	txRateSelect             reg8 //88
	rx4ApplicationSelect     reg8 //89
	rx3ApplicationSelect     reg8 //90
	rx2ApplicationSelect     reg8 //91
	rx1ApplicationSelect     reg8 //92
	highPowerClassEnable     reg8 //93
	tx4ApplicationSelect     reg8 //94
	tx3ApplicationSelect     reg8 //95
	tx2ApplicationSelect     reg8 //96
	tx1ApplicationSelect     reg8 //97
	cdrControl               reg8 //98
	_                        byte
	txLosMask                reg8 //100
	txAdaptEqFaultMask       reg8 //101
	txCdrLolMask             reg8 //102
	tempAlarm                reg8 //103
	vccAlarm                 reg8 //104
	_                        [3]byte
	propagationDelay         reg16 //108-109
	advancedLowPowerMode     reg8  //110
	pcisig                   reg16 //111-112
	farNearEndImplementation reg8  //113
	_                        [13]byte
	pageSelect               reg8 //127
}
type lpage0 struct {
	id                           byte    //0
	status                       uint16  //1-2
	channelStatusInterrupt       [3]byte //3-5
	freeMonitorInterruptFlags    [3]byte //6-8
	channelMonitorInterruptFlags [6]byte //9-14
	_                            [7]byte
	freeMonTemp                  uint16 //22-23
	_                            [2]byte
	freeMonVoltage               uint16 //26-27
	_                            [6]byte
	rxPower                      [8]byte //34-41
	txBias                       [8]byte //42-49
	txPower                      [8]byte //50-57
	_                            [26]byte
}

type upage3 struct {
	_                  [128]byte
	tempHighAlarm      uint16 //128-129
	tempLowAlarm       uint16 //130-131
	tempHighWarning    uint16 //132-133
	tempLowWarning     uint16 //134-135
	_                  [8]byte
	vccHighAlarm       uint16 //144-145
	vccLowAlarm        uint16 //146-147
	vccHighWarning     uint16 //148-149
	vccLowWarning      uint16 //150-151
	_                  [8]byte
	_                  [16]byte
	rxPowerHighAlarm   uint16 //176-177
	rxPowerLowAlarm    uint16 //178-179
	rxPowerHighWarning uint16 //180-181
	rxPowerLowWarning  uint16 //182-183
	txBiasHighAlarm    uint16 //184-185
	txBiasLowAlarm     uint16 //186-187
	txBiasHighWarning  uint16 //188-189
	txBiasLowWarning   uint16 //190-191
	txPowerHighAlarm   uint16 //192-193
	txPowerLowAlarm    uint16 //194-195
	txPowerHighWarning uint16 //196-197
	txPowerLowWarning  uint16 //198-199
}
